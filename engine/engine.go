package engine

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
)

type SearchEngine struct {
	Num      int    `json:"num"`
	Root     string `json:"Root"`
	Worker   []Obj
	reciver  chan []utils.Line
	sender   chan string
	topOrder chan string
	handle   func(lines []utils.Line)
}

func EngineInit(path ...string) *SearchEngine {
	root, _ := os.Getwd()
	if path != nil {
		root = path[0]
	}
	return &SearchEngine{
		Num:  20,
		Root: root,
	}
}

func (self *SearchEngine) Factory(listen func(lines []utils.Line)) {
	self.reciver = make(chan []utils.Line, self.Num)
	self.sender = make(chan string, 20)
	self.topOrder = make(chan string)
	waitFileArea := make(map[string]Obj)
	go func() {
		for {
			filepath.Walk(self.Root, func(root string, state os.FileInfo, err error) error {
				if !state.IsDir() {
					if strings.Contains(root, ".") {
						if obj, err := OpenObj(root); err != nil {
							log.Println("[Err]:", err)
							return err
						} else {
							if obj != nil {
								if _, ok := waitFileArea[root]; !ok {
									waitFileArea[root] = obj
									// 工人开始工作

									log.Println(Blue("[+]:"), Green(root))
									go obj.Work(self.sender, self.reciver)
								}
							}
						}
					}
				}
				return nil
			})
			time.Sleep(1 * time.Second)
		}

	}()

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

	for {
		op := <-self.topOrder
		for range waitFileArea {
			self.sender <- op
		}
	}
}

func (self *SearchEngine) SetResultListener(listen func(lines []utils.Line)) {
	self.handle = listen
}

func (self *SearchEngine) Search(key string) {
	self.topOrder <- key
}
