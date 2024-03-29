package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	//	"searcher/engine"
	"flag"

	"github.com/Qingluan/FrameUtils/engine"
	"github.com/Qingluan/FrameUtils/utils"
	"github.com/c-bata/go-prompt"
)

type Datas map[string]string

func Repl(label string, suggest Datas) string {
	return prompt.Input(label, func(d prompt.Document) (s []prompt.Suggest) {
		for k, v := range suggest {
			s = append(s, prompt.Suggest{
				Text:        k,
				Description: v,
			})
		}
		return prompt.FilterFuzzy(s, d.GetWordBeforeCursor(), true)
	})
}

func main() {
	cli := false
	frp := ""
	dst := ""
	out := ""
	filter := ""
	root := ""
	grepStr := ""
	tps := ""
	flag.BoolVar(&cli, "cli", false, "true to console")
	flag.StringVar(&frp, "fr", "", "set from file .")
	flag.StringVar(&dst, "to", "", "set to file.")
	flag.StringVar(&out, "out", "", "output path dir.")
	flag.StringVar(&root, "r", ".", " set root dir .")
	flag.StringVar(&grepStr, "s", "", " set grep str ")
	flag.StringVar(&tps, "t", "", " set typs str ")

	flag.StringVar(&filter, "grep", "", "fileter table if type is sql/xlsx.")
	flag.Parse()
	if grepStr != "" {
		sengine := engine.EngineInit(root)
		if tps != "" {
			tpss := strings.Split(tps, ",")
			sengine.SetFilter(tpss...)
		}
		go sengine.Factory(func(lines []utils.Line) {
			// log.Println(utils.Green("Found : "))
			for _, line := range lines {
				fmt.Println(utils.Yellow(line[0]))
			}
		}, true)
		sengine.SetResultListener(func(ls []utils.Line) {
			// for _, l := range ls {
			// fmt.Println(l)
			// }
		})

		sengine.Search(grepStr)
		time.Sleep(20 * time.Second)
		<-sengine.IfEnd

		return
	}
	if cli {
		sengine := engine.EngineInit()
		go sengine.Factory(nil, false)

		sengine.SetResultListener(func(ls []utils.Line) {
			for _, l := range ls {
				fmt.Println(l)
			}
		})

		for {
			e := Repl("search some >", Datas{"exit": "exit process"})
			if e == "exit" {
				break
			}
			sengine.Search(e)
		}
	} else {
		if frp != "" && dst != "" {
			obj, err := engine.OpenObj(frp)
			// fmt.Println("Opened obj:", obj)

			if err != nil {
				log.Fatal(err)
				time.Sleep(2 * time.Second)
			}
			// fmt.Println("Opened obj:", obj)
			// c := 0
			fsd := make(map[string]*os.File)
			// for _, d := range obj.AsJson() {
			// 	fmt.Println(d)
			// }
			var objs <-chan utils.Line
			if filter != "" {
				objs = obj.Iter(filter)
			} else {
				objs = obj.Iter()
			}
			for line := range objs {
				table := line[0]
				// fmt.Println(line)
				// break
				if fp, ok := fsd[table]; ok {
					fp.WriteString(strings.Join(line[1:], ",") + "\n")
				} else {
					tablePath := filepath.Join(out, table+".csv")
					fsd[table], err = os.Create(tablePath)
					header := obj.GetHeader(table)
					// fmt.Println("header:", header, "\nvalue:", line)
					header = header[:len(line)-1]
					// fmt.Println(header, "\n----\n")
					msg := strings.Join(header, ",") + "\n"
					_, err := fsd[table].WriteString(msg)
					if err != nil {
						log.Fatal("Err", err)
					}
					_, err = fsd[table].WriteString(strings.Join(line[1:], ",") + "\n")
					if err != nil {
						log.Fatal(err)
					}
				}
				// if c % 20000 == 0{
				// 	fmt.Printf("Flow: %d\r",c)
				// }
			}
			for _, fs := range fsd {
				fs.Close()
			}
		}
	}

}
