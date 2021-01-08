package webevent

import (
	"fmt"
)

// TableLayout: Ori: vertical / horizen
type TableLayout struct {
	Orientation string
	Id          string
	Items       []ElementAble
	Text        string
}

func (table *TableLayout) Content(c string) ElementAble {
	table.Text = c
	return table
}
func (table *TableLayout) GetID() string {
	return table.Id
}
func (table *TableLayout) String() string {
	if table.Orientation == "horizen" {
		tmps := ""
		for _, e := range table.Items {
			tmps += fmt.Sprintf(`
			<td valign="top" >
				%s
			</td>`, e.String())
		}
		return fmt.Sprintf(`<table>
		<tr>
			%s
		</tr>
	</table>`, tmps)
	} else {
		tmps := ""
		for _, e := range table.Items {
			tmps += fmt.Sprintf(`
			<tr >
				%s
			</tr>`, e.String())
		}
		return fmt.Sprintf(`<table>
			%s
	</table>`, tmps)
	}
}
