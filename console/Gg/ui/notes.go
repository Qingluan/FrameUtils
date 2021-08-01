package ui

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sort"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	SEP = " • "
)

type Items struct {
	Tp     string
	Status string
	Value  interface{}
}

type Notes struct {
	Items      map[string]interface{}
	Now        map[string]interface{}
	lastCursor []int

	Page      int
	Index     []string
	cursor    int // which to-do list item our cursor is pointing at
	keys      string
	paginator paginator.Model
}

func (m *Notes) Init() tea.Cmd {
	return nil
}

func NewNotes(filepath string) (n *Notes) {
	n = new(Notes)
	n.Now = make(map[string]interface{})
	n.Items = make(map[string]interface{})
	n.Load(filepath)

	return
}

func (m *Notes) Load(filepath string) {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("load notes err must json:", err)
	}
	items := make(map[string]interface{})
	json.Unmarshal(buf, &items)
	m.Items = items
	m.Now = m.Items
	m.UpdateKeys()
	// fmt.Println(m.Items)

}

func (m *Notes) UpdateKeys() {
	m.Index = []string{}
	for k := range m.Now {
		m.Index = append(m.Index, k)
	}

	sort.Slice(m.Index, func(p, q int) bool {
		return m.Index[p][0] < m.Index[q][0]
	})
	// fmt.Println("page:", len(m.Index))
	if len(m.Index) > 10 {
		p := paginator.NewModel()
		p.Type = paginator.Dots
		p.PerPage = 10
		p.ActiveDot = utils.Green("•")
		p.InactiveDot = "•"
		m.paginator = p
		m.paginator.SetTotalPages(len(m.Index))

	} else {
		// m.paginator.SetTotalPages(0)
		m.paginator.TotalPages = 0
	}

	// n.paginator = p

}

func (m *Notes) GetItem() interface{} {
	return m.Now[m.Index[m.cursor]]
}

func (m *Notes) SetItem(newval interface{}) {
	m.Now[m.Index[m.cursor]] = newval
}

func (m *Notes) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.Index)-1 {
				m.cursor++
			}

		case "h":
			if m.paginator.TotalPages != 0 {

				m.paginator, cmd = m.paginator.Update(msg)
				_, end := m.paginator.GetSliceBounds(len(m.Index))
				m.cursor = end + end - m.cursor
				m.Page = m.paginator.Page
			}

		case "l":
			if m.paginator.TotalPages != 0 {

				start, _ := m.paginator.GetSliceBounds(len(m.Index))
				left := m.cursor - start
				m.paginator, cmd = m.paginator.Update(msg)
				start, _ = m.paginator.GetSliceBounds(len(m.Index))
				m.cursor = start + left
				m.Page = m.paginator.Page
			}
		case "backspace":
			// fmt.Println("ssss\n\nfsaf")
			keys := strings.Split(m.keys, SEP)
			keys = keys[:len(keys)-1]
			m.keys = strings.Join(keys, SEP)
			var s = m.Items
			for _, k := range keys {
				if _s, ok := s[k]; ok {
					s = _s.(map[string]interface{})
				}
			}
			if s != nil {
				m.Now = s
				m.UpdateKeys()
				if len(m.lastCursor) > 0 {
					m.cursor = m.lastCursor[len(m.lastCursor)-1]
					m.lastCursor = m.lastCursor[:len(m.lastCursor)-1]
				} else {
					m.cursor = 0
				}

			}

			// if m.paginator.
			// m.paginator.NextPage()
		case "enter":
			key := m.Index[m.cursor]
			if strings.HasPrefix(key, "[]") {
				n := m.Now[key].(map[string]interface{})["status"].(bool)
				if n {
					m.Now[key].(map[string]interface{})["status"] = false
				} else {
					m.Now[key].(map[string]interface{})["status"] = true
				}
			} else {
				if subv, ok := m.Now[key]; ok {
					switch subv.(type) {
					case map[string]interface{}:

						m.Now = subv.(map[string]interface{})
						m.UpdateKeys()
						if m.keys != "" {
							m.keys += SEP + key
						} else {
							m.keys = key
						}

						// m.lastPage = m.paginator.Page
						m.lastCursor = append(m.lastCursor, m.cursor)
						m.cursor = 0
					}
				}
			}

		case "q", "esc", "ctrl+c":
			return m, tea.Quit
			// default:
			// m.paginator, cmd = m.paginator.Update(msg)

		}
	}
	if m.paginator.TotalPages != 0 {
		m.paginator, cmd = m.paginator.Update(msg)
	}
	return m, cmd
}

func (m *Notes) View() string {
	var b strings.Builder
	b.WriteString("\n  Notes " + m.keys + "\n\n")
	ranges := []string{}
	if m.paginator.TotalPages != 0 {
		start, end := m.paginator.GetSliceBounds(len(m.Index))
		ranges = m.Index[start:end]
	} else {
		ranges = m.Index
	}
	for i, item := range ranges {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}
		checked := "•"
		if itemValue, ok := m.Items[item]; ok {
			switch itemValue.(type) {
			case map[string]interface{}:
				if v, ok := itemValue.(map[string]interface{})["status"]; ok && strings.HasPrefix(item, "[]") {
					if v.(bool) {
						checked = "◉" // selected!
					} else {
						checked = "੦" // not selected
					}
					item = item[2:]
					checked = utils.Red(checked)

				} else {
					checked = utils.Blue("¶")
				}

			default:
			}
		}
		b.WriteString(cursor + checked + " " + item + "\n")
	}
	if m.paginator.TotalPages != 0 {
		b.WriteString("  " + m.paginator.View())
	}
	b.WriteString("\n\nj/k move > |  h/l ←/→ page • q: quit\n")
	return b.String()
}

func MainNote(root string) {
	p := tea.NewProgram(NewNotes(root))
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
