package web

import (
	"bytes"
	"html/template"
	"log"
	"strings"

	"github.com/Qingluan/FrameUtils/asset"
	"github.com/tealeg/xlsx/v3"
)

type WebValue struct {
	Type  string
	Value string
}

func (v *WebValue) Info() string {
	if v.Type == "link" {
		return v.Value
	} else {
		return v.Type
	}
}

func (v *WebValue) String() string {
	if v.Type == "link" {
		if len(v.Value) > 14 {
			return v.Value[:14] + "..[link].."
		}
	}
	return v.Value

}

type WebTableRow struct {
	Cells map[int]*WebValue
}

func TableReadLine(lines []string) (e WebTableRow) {
	e.Cells = make(map[int]*WebValue)
	for i, k := range lines {
		k = strings.TrimSpace(k)
		if strings.HasPrefix(k, "http://") || strings.HasPrefix(k, "https://") {
			e.Cells[i] = &WebValue{Type: "link", Value: k}
		} else {
			e.Cells[i] = &WebValue{Value: k}
		}
	}
	return e
}

type WebTable struct {
	Keys   []string
	Values []WebTableRow
}

func TableReadLines(lines [][]string) (es WebTable) {
	for _, k := range lines {
		es.Values = append(es.Values, TableReadLine(k))
	}
	return es
}

var (
	TABLE_TEMPLATE, _      = asset.AssetAsFile("Res/templates/table.html")
	BASE_INDEX_TEMPLATE, _ = asset.AssetAsFile("Res/templates/base.html")
)

func (table *WebTable) Parse(name string) string {
	// d, _ := asset.Asset(TABLE_TEMPLATE)
	t, err := template.New(name).ParseFiles(TABLE_TEMPLATE)

	if err != nil {
		log.Fatal("Parse Table Err:", err, TABLE_TEMPLATE, "||")
	}
	buffer := bytes.NewBufferString("")
	err = t.ExecuteTemplate(buffer, "table", table)
	if err != nil {
		log.Fatal("Parse Table Err:", err, TABLE_TEMPLATE, "||")
	}

	return buffer.String()
}

func TableReadXlsx(name string) (es WebTable, err error) {
	wb, err := xlsx.OpenFile(name)
	if err != nil {
		return
	}

	rowVistor := func(r *xlsx.Row) error {
		i := 0
		datas := []string{}
		for {
			// fmt.Println(r.Sheet.MaxCol)

			if i < r.Sheet.MaxCol {
				datas = append(datas, r.GetCell(i).Value)
				i++
				continue
			}
			break
		}
		es.Values = append(es.Values, TableReadLine(datas))
		// r.ForEachCell(cells)
		// r.ForEachCell()
		return nil
	}
	for _, sh := range wb.Sheets {
		sh.ForEachRow(rowVistor)
		break
	}
	return
}
