package WebRender

import (
	"fmt"
	"strings"
)

var (
	leftRightContainerTMP = `
<div class="container h-100">
	<div class="row h-100">
		<div class="d-flex bd-highlight">
			<div class="bd-highlight" style=" background: #6a92bb; max-width: 400px;min-width: 400px;    padding: 10px;  border-right: 1px black solid;">
				<div class="accordion" id="accordionExample">
					%s
				</div>
			</div>
		</div>
		<div class="m-6 flex-fill bd-highlight;  padding: 10px;">
			%s
		</div>
	</div>
</div>
	`
)

type LeftRightContainer struct {
	LeftItems    map[string]string
	RightContent string
}

func (left *LeftRightContainer) String() string {
	items := []string{}
	for k, v := range left.LeftItems {
		items = append(items, fmt.Sprintf(`<a class="nav-link " id="tab-item-%s" data-toggle="pill" href="%s" role="tab" aria-selected="true">%s</a>`, strings.ReplaceAll(k, " ", "_"), v, k))
	}
	leftContent := fmt.Sprintf(`<div class="nav flex-column nav-pills" id="v-pills-tab" role="tablist" aria-orientation="vertical">%s</div>`, strings.Join(items, "\n"))
	return fmt.Sprintf(leftRightContainerTMP, leftContent, left.RightContent)
}
