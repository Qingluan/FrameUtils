package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Qingluan/FrameUtils/asset"
	"github.com/Qingluan/FrameUtils/utils"
	"github.com/tealeg/xlsx/v3"
)

type WebInput struct {
	ID      string
	Name    string
	Title   string
	Type    string
	Values  []string
	Default string
}

type WebForm struct {
	Action   string
	ID       string
	Uri      string
	Forms    []*WebInput
	Collects []TData
	Keys     []string
	IsBuild  bool
}

var (
	FORM_TEMPLATE, _ = asset.AssetAsFile("Res/templates/form.html")
)

/* ParseFrom

# normal :
// name , [default] ,  id = "xxx", type="xxx"

# select:
// name , id=xx , / options1 / options2
# date
// name , id=xx , type = "date"
# num
// name , id=xx , type ="num"
*/
func ParseFrom(line string) (w *WebInput) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "//") {
		return nil
	}
	fields := utils.SplitByIgnoreQuote(line[2:], "/")
	w = new(WebInput)

	args, kargs := utils.DecodeToOptions(fields[0])
	defaults := ""
	if len(args) > 1 {
		defaults = args[1]
	}
	name := args[0]
	w.Name = name
	w.Default = defaults
	if len(fields) > 1 {
		w.Type = "select"
		w.Values = fields[1:]
	} else {
		if v, ok := kargs["type"]; ok {
			w.Type = v.(string)
		}
		if v, ok := kargs["id"]; ok {
			w.ID = v.(string)
		}

		if v, ok := kargs["title"]; ok {
			w.Title = v.(string)
		} else {
			w.Title = w.Name
		}
	}
	return
}

func FormsParseFromXlsx(name string, uri ...string) (forms *WebForm) {
	wb, err := xlsx.OpenFile(name)
	if err != nil {
		return
	}
	forms = new(WebForm)

	for _, sh := range wb.Sheets {
		i := 0
		for {
			if i < sh.MaxCol {
				c, err := sh.Cell(0, i)
				i++
				if err != nil {
					log.Println("get cel err:", err)
					continue
				}
				if c.Value != "" {
					log.Println(c.Value)
					form := ParseFrom(c.Value)
					// form := ParseFrom(line)
					if form.Name != "" {
						forms.Forms = append(forms.Forms, form)
					}
				}

				continue
			}
			break
		}
		break
	}
	if uri != nil {
		forms.Uri = uri[0]
		forms.BuildHanlde(uri[0])
	}
	if forms == nil {
		log.Fatal("form s build err")
	}
	return
}

func FormsParseForms(lines string) (forms *WebForm) {
	forms = new(WebForm)
	for _, line := range utils.SplitByIgnoreQuote(lines, "\n") {
		if strings.HasPrefix(line, "//") {
			form := ParseFrom(line)
			if form.Name != "" {
				forms.Forms = append(forms.Forms, form)
			}
		}
	}
	return
}

func (forms *WebForm) SetID(id string) *WebForm {
	forms.ID = id
	return forms
}

func (forms *WebForm) SetAction(action string) *WebForm {
	forms.Action = action
	return forms
}

func (forms *WebForm) Parse(name string) string {
	t, err := template.New(name).ParseFiles(FORM_TEMPLATE)
	if err != nil {
		log.Fatal("Parse form Err:", err, FORM_TEMPLATE)
	}
	if forms.Action == "" {
		forms.Action = "json"
	}
	buffer := bytes.NewBufferString("")
	err = t.ExecuteTemplate(buffer, "form", forms)
	if err != nil {
		log.Fatal("Parse form Err:", err, " fomr:", FORM_TEMPLATE)
	}
	return buffer.String()
}

func (forms *WebForm) HandleCollect(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		data := TData{}
		if err = json.Unmarshal(bytes, &data); err == nil {
			for key := range data {
				if !utils.ArrayContains(forms.Keys, key) {
					forms.Keys = append(forms.Keys, key)
				}
			}
			forms.Collects = append(forms.Collects, data)
		}
		buf, _ := json.Marshal(&TData{
			"tp":  "msg",
			"msg": "Collect Ok!",
		})
		w.Write(buf)
	}
}

func (forms *WebForm) BuildHanlde(uri string) {
	if !forms.IsBuild {
		http.HandleFunc(uri, forms.HandleCollect)
		forms.IsBuild = true
	}
}

func (forms *WebForm) ToTable() (table *WebTable) {
	table = new(WebTable)
	table.Values = append(table.Values, TableReadLine(forms.Keys))

	for _, line := range forms.Collects {
		data := []string{}
		for _, name := range forms.Keys {
			data = append(data, line[name].(string))
		}
		table.Values = append(table.Values, TableReadLine(data))
	}
	return table
}

func (forms *WebForm) ToXlsx(name, sheet string) (err error) {
	wb := xlsx.NewFile()

	sh, err := wb.AddSheet(sheet)
	if err != nil {
		return err
	}
	row := sh.AddRow()
	row.SetHeight(12)
	for _, name := range forms.Keys {
		cell := row.AddCell()
		cell.Value = name
	}

	for _, line := range forms.Collects {
		// fmt.Println(line)
		row := sh.AddRow()
		row.SetHeight(12)
		for _, name := range forms.Keys {
			if val, ok := line[name]; ok {
				cell := row.AddCell()
				cell.Value = val.(string)
				// cell.SetString()
				// row.PushCell(cell)
				// fmt.Print("set:", val.(string))

				// row.PushCell(cell)
			}
		}
		// fmt.Println(row)

	}
	for i := range forms.Keys {

		sh.SetColAutoWidth(i, func(val string) float64 {
			return float64(len(val))
		})

	}
	err = wb.Save(name)
	if err != nil {
		log.Println("save xlsx err:", err)
	}
	return
}
