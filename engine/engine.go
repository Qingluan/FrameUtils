package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
)

type SearchEngine struct {
	Num       int    `json:"num"`
	Root      string `json:"Root"`
	Worker    []Obj
	reciver   chan []utils.Line
	sender    chan string
	fileTypes map[string]int
	topOrder  chan string
	handle    func(lines []utils.Line)
	IfEnd     chan bool
}

func EngineInit(path ...string) *SearchEngine {
	root, _ := os.Getwd()
	if path != nil {
		root = path[0]
	}
	return &SearchEngine{
		Num:       20,
		Root:      root,
		IfEnd:     make(chan bool),
		fileTypes: make(map[string]int),
	}
}

func (self *SearchEngine) SetFilter(fs ...string) {
	for _, f := range fs {
		self.fileTypes[f] = 1

	}
}

func (self *SearchEngine) Factory(listen func(lines []utils.Line), singleModel bool) {
	self.reciver = make(chan []utils.Line, self.Num)
	self.sender = make(chan string, 20)
	self.topOrder = make(chan string, 10)
	waitFileArea := make(map[string]Obj)
	go func(waitFileArea map[string]Obj) {
		for {
			filepath.Walk(self.Root, func(root string, state os.FileInfo, err error) error {
				if !state.IsDir() {

					if strings.Contains(root, ".") {

						if len(self.fileTypes) > 0 {
							tps := strings.Split(root, ".")
							tp := tps[len(tps)-1]
							if _, ok := self.fileTypes[tp]; !ok {
								return nil
							}
						}

						if obj, err := OpenObj(root); err != nil {
							log.Println("[Err]:", err)
							return err
						} else {
							if obj != nil {
								if _, ok := waitFileArea[root]; !ok {
									waitFileArea[root] = obj
									// 工人开始工作

									// log.Println(Blue("[+]:"), Green(root))
									go waitFileArea[root].Work(self.sender, self.reciver)
								}
							}
						}
					}
				}
				return nil
			})
			time.Sleep(2 * time.Second)
			log.Println(Blue("Finishe "))
			self.IfEnd <- true
			break
		}

	}(waitFileArea)

	if listen != nil {
		self.SetResultListener(listen)
	}
	go func() {
		for {
			rows := <-self.reciver
			if self.handle != nil {
				self.handle(rows)
			}

		}
	}()
	ifb := false
	for {
		if len(self.topOrder) == 0 {
			time.Sleep(1 * time.Second)
			// fmt.Print("wait: 1s\r")
			continue
		}
		op := <-self.topOrder
		ifb = true
		fmt.Println("do", op)
		for range waitFileArea {
			self.sender <- op
		}
		if singleModel && ifb {
			break
		}
	}
}

func (self *SearchEngine) SetResultListener(listen func(lines []utils.Line)) {
	self.handle = listen
}

func (self *SearchEngine) Search(key string) {
	// fmt.Println("do")
	if self.topOrder == nil {
		self.topOrder = make(chan string, 10)
	}
	// fmt.Println("do:", key)
	self.topOrder <- key
	// self.sender <- key
}
