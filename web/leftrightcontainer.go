package web

import (
	"fmt"
	"strings"
)

var (
	leftRightContainerTMP = `
<div class="container h-100" style="width: inherit;padding: 0px;max-width: 122000px;margin: 0px;">
	<div class="row h-100" style="width:100%%;">
		<div class="d-flex bd-highlight" style="width:100%%;" >
			<div class="bd-highlight" style=" background: #6a92bb; max-width: 400px;min-width: 400px;    padding: 10px;  border-right: 1px black solid;">
				<div class="accordion" id="accordionExample">
					%s
				</div>
			</div>
			<div class="m-6 flex-fill bd-highlight;  padding: 10px;">
				%s
			</div>
		</div>
	</div>
</div>
	`
)

type ColsContainer struct {
	Links   map[string]string
	Content string
}

func (left *ColsContainer) SetContent(str string) {
	left.Content = str
}

func (left ColsContainer) String() string {
	items := []string{}
	for k, v := range left.Links {
		items = append(items, fmt.Sprintf(`<a class="nav-link " id="tab-item-%s" data-toggle="pill" href="%s" role="tab" aria-selected="true">%s</a>`, strings.ReplaceAll(k, " ", "_"), v, k))
	}
	leftContent := fmt.Sprintf(`<div class="nav flex-column nav-pills" id="v-pills-tab" role="tablist" aria-orientation="vertical">%s</div>`, strings.Join(items, "\n"))
	return fmt.Sprintf(leftRightContainerTMP, leftContent, left.Content)
}

type Container interface {
	SetContent(str string)
	String() string
}
