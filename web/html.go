package web

import "fmt"

func CardWrap(title string, content string) string {
	return fmt.Sprintf(`<div class="card">
    <h5 class="card-title">%s</h5>
    <div class="card-body">%s</div>
</div>`, title, content)
}
