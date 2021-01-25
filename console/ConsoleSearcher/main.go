package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	//	"searcher/engine"
	"flag"

	"github.com/Qingluan/FrameUtils/engine"
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
	flag.BoolVar(&cli, "cli", false, "true to console")
	flag.StringVar(&frp, "fr", "", "set from file .")
	flag.StringVar(&dst, "to", "", "set to file.")
	flag.Parse()
	if cli {
		sengine := engine.EngineInit()
		go sengine.Factory(nil)

		sengine.SetResultListener(func(ls []engine.Line) {
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
			if err != nil {
				log.Fatal(err)
			}
			// c := 0
			fsd := make(map[string]*os.File)
			// for _, d := range obj.AsJson() {
			// 	fmt.Println(d)
			// }
			for line := range obj.Iter() {
				table := line[0]
				if fp, ok := fsd[table]; ok {
					fp.WriteString(strings.Join(line[1:], ",") + "\n")
				} else {
					fsd[table], err = os.Create(table + ".csv")
					header := obj.GetHeader(table)
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
