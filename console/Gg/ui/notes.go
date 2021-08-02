package ui

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
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
	Msg        string
	target     string
	keys       string
	Choose     string
	cursor     int // which to-do list item our cursor is pointing at
	active     int
	Page       int
	Index      []string
	lastCursor []int
	Exit       bool
	textInput  textinput.Model
	paginator  paginator.Model
	// pres      chan string
	Items map[string]interface{}
	Now   map[string]interface{}

	After map[string]func(s string) bool
}

func (m *Notes) Init() tea.Cmd {
	if m.active == 0 {
		return nil
	} else {
		return textinput.Blink
	}
}

func NewNotes(filepath string) (n *Notes) {
	n = new(Notes)
	n.Now = make(map[string]interface{})
	n.Items = make(map[string]interface{})
	n.textInput = textinput.NewModel()
	n.Load(filepath)
	n.After = make(map[string]func(s string) bool)
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

func (m *Notes) UpdateKeys(now ...string) {
	if m.active != 0 {
		if m.textInput.Prompt == "" {
			m.textInput = textinput.NewModel()
		}
		m.textInput.SetValue("")
		if now != nil {
			m.textInput.Prompt = now[0]
		}
		m.textInput.Focus()
		m.textInput.CharLimit = 156
		m.textInput.Width = 20

		return
	}
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
	if m.active == 0 {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}

				start, _ := m.paginator.GetSliceBounds(len(m.Index))

				if m.cursor < start {
					m.paginator.PrevPage()
				}

			// The "down" and "j" keys move the cursor down
			case "down", "j":
				if m.cursor < len(m.Index)-1 {
					m.cursor++
				}
				_, end := m.paginator.GetSliceBounds(len(m.Index))

				if m.cursor >= end {
					m.paginator.NextPage()
				}
			case "+":
				m.active = 1
				m.UpdateKeys("create item >")
			case "-":
				m.WithEdit("Delete it ?(yes/other) ", func(s string) bool {
					if s == "yes" {
						m.Del(m.GetCursor())
						m.UpdateKeys()
						if m.cursor > 0 {
							m.cursor -= 1
						}
					}
					return true
				})
			case "h":
				if m.paginator.TotalPages != 0 {
					start, _ := m.paginator.GetSliceBounds(len(m.Index))
					left := m.cursor - start
					m.paginator, cmd = m.paginator.Update(msg)
					start, _ = m.paginator.GetSliceBounds(len(m.Index))
					m.cursor = start + left
					m.Page = m.paginator.Page

				}

			case "l":
				if m.paginator.TotalPages != 0 {

					start, end := m.paginator.GetSliceBounds(len(m.Index))
					left := m.cursor - start

					m.paginator, cmd = m.paginator.Update(msg)
					start, end = m.paginator.GetSliceBounds(len(m.Index))
					m.cursor = start + left
					if m.cursor >= end {
						m.cursor = end - 1
					}
					m.Page = m.paginator.Page
				}
			case "backspace":
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
			case "u":
				m.Msg = "Updateing"
				// m.WithEdit("update..")
				m.WithEdit("git password:", func(s string) bool {
					os.RemoveAll(PWDDIR)
					GetRepo(s)
					Update(s)
					m.Msg = ""
					// m.Reload()
					m.Exit = true
					return true
				}, true)

			case "ctrl+s":
				m.WithEdit("save ? ", func(w string) bool {
					if w == "yes" {
						m.WithEdit("encryped notes :", func(s string) bool {
							m.Save(s)
							m.WithEdit("git password :", func(ss string) bool {
								Upload(ss)
								return true
							}, true)
							return false
						}, true)
						return false
					}

					return true
				})

				// fmt.Println("ssss\n\nfsaf")

				// if m.paginator.
				// m.paginator.NextPage()
			case "enter":
				key := m.Index[m.cursor]
				if strings.HasPrefix(key, "[]") {
					val := m.Now[key]
					switch val.(type) {
					case map[string]interface{}:
						n := val.(map[string]interface{})["status"].(bool)
						if n {
							val.(map[string]interface{})["status"] = false
						} else {
							val.(map[string]interface{})["status"] = true
						}
					case bool:
						if val.(bool) {
							m.Now[key] = false
						} else {
							m.Now[key] = true
						}
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
						case bool:
							if subv.(bool) {
								m.Now[key] = false
							} else {
								m.Now[key] = true
							}
						case string:
							m.Choose = subv.(string)
							return m, tea.Quit
						}
					}
				}
			case "e":
				// case string:
				if subv, ok := m.Now[m.GetCursor()]; ok {
					switch subv.(type) {
					case string:

						m.Edit(subv.(string))

					}
				}

			case "q", "esc", "ctrl+c":
				if m.active == 0 {

					return m, tea.Quit
				} else {
					m.active = 0
				}
				// default:
				// m.paginator, cmd = m.paginator.Update(msg)

			}
		}
		if m.paginator.TotalPages != 0 {
			m.paginator, cmd = m.paginator.Update(msg)
		}

	} else {
		// fmt.Println(msg)
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "esc":
				m.active = 0
				m.Msg = ""
			case "enter":
				if m.active == 2 {
					m.SetItem(m.textInput.Value())
					m.active = 0
				} else if m.active == 1 {
					m.Add(m.textInput.Value())
					m.UpdateKeys()
					m.active = 0
				} else {
					if m.target != "" {
						f := m.After[m.target]
						delete(m.After, m.target)
						if f(m.textInput.Value()) {
							m.active = 0
						} else {
							m.textInput.Prompt = m.target
						}
					}

				}

			}
		}
		m.textInput, cmd = m.textInput.Update(msg)

		// fmt.Println(m.textInput.View())
	}
	if m.Exit {
		return m, tea.Quit
	}
	return m, cmd
}

func (m *Notes) GetCursor() string {
	return m.Index[m.cursor]
}

func (m *Notes) Del(key string) {
	delete(m.Now, key)
	tmp := []string{}
	for _, v := range m.Index {
		if v == key {
			continue
		}
		tmp = append(tmp, v)
	}
	m.Index = tmp
}

func (m *Notes) Save(pass string) {
	buf, err := json.Marshal(m.Items)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Generate new notes json...")
	fmt.Print("AES Crypt Password: ", pass)
	// bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("get passwd err:", err)
	}

	log.Println("Generate crypted file to git ....")
	en, err := AesEncrypt(buf, []byte(pass))
	if err != nil {
		log.Fatal("en err:", err)
	}
	ioutil.WriteFile(filepath.Join(PWDDIR, "note.en"), en, os.ModePerm)
}

func (m *Notes) WithEdit(now string, do func(s string) bool, secure ...bool) {
	m.active = 3
	if secure != nil && secure[0] {
		m.active = 4
	}
	m.target = now
	m.After[now] = do

	m.UpdateKeys(now)
}

func (m *Notes) Edit(now string) {
	m.textInput.Placeholder = now
	m.textInput.SetValue(now)
	m.active = 2
	m.UpdateKeys()
	// m.textInput.Focus()
	// m.textInput.CharLimit = 156
	// m.textInput.Width = 20
}
func (m *Notes) Add(res string) {
	if strings.Contains(res, "=") {
		fs := strings.SplitN(res, "=", 2)
		nkey := strings.TrimSpace(fs[0])
		nval := strings.TrimSpace(fs[1])
		if strings.HasPrefix(nkey, "[]") {
			m.Index = append(m.Index, nkey)
			m.Now[nkey] = false
		} else if nval == "" {
			tmp := []string{}
			for _, v := range m.Index {
				if v == nkey {
					continue
				}
				tmp = append(tmp, v)
			}
			m.Index = tmp
			delete(m.Now, nkey)
		} else {
			m.Index = append(m.Index, nkey)
			m.Now[nkey] = nval
		}
		// return dict
	} else {
		m.Index = append(m.Index, strings.TrimSpace(res))
		m.Now[strings.TrimSpace(res)] = map[string]interface{}{}
	}
}

func (m *Notes) GetInput() string {
	if m.active == 4 {
		msg := m.textInput.Value()
		m := m.textInput.Prompt
		for i := 0; i < len(msg); i++ {
			m += "*"
		}
		return m + fmt.Sprintf("(%d)", len(m))
	}
	return m.textInput.View()
}

func (m *Notes) View() string {
	var b strings.Builder
	b.WriteString("j/k move lines h/l ←/→ page • +: add new item • -: remove item \nq: quit  • ctrl+s: save • u: update • \n")
	if len(m.keys) > 0 {
		b.WriteString(utils.Yellow(m.keys) + "\n")
	}

	ranges := []string{}
	start := 0
	end := 0
	if m.paginator.TotalPages != 0 {
		start, end = m.paginator.GetSliceBounds(len(m.Index))
		ranges = m.Index[start:end]
	} else {
		start = 0
		ranges = m.Index
	}

	for i, item := range ranges {

		cursor := " " // no cursor
		if m.cursor == i+start {
			cursor = ">" // cursor!
		}
		checked := "•"
		if itemValue, ok := m.Now[item]; ok {
			switch itemValue.(type) {
			case map[string]interface{}:
				if v, ok := itemValue.(map[string]interface{})["status"]; ok && strings.HasPrefix(item, "[]") {

					item = item[2:]
					if v.(bool) {
						checked = "- ◉" // selected!
						item = utils.Green(item) + "\n"
						checked = utils.Red(checked)
					} else {
						checked = "- ੦" // not selected
						item = utils.Yellow(item) + "\n"
					}
					checked = utils.Red(checked)

				} else {
					checked = utils.Blue("+ ¶")
				}
			case bool:
				item = strings.TrimPrefix(item, "[]")
				if itemValue.(bool) {
					checked = "- ◉" // selected!
					item = utils.Green(item) + "\n"
					checked = utils.Red(checked)
				} else {
					checked = "- ੦" // not selected
					item = utils.Yellow(item) + "\n"
				}
			default:
				if cursor == ">" {
					item += "\n\t" + utils.UnderLine(utils.Blue(itemValue.(string)))
				} else {
					item += "\n"
				}
			}
		}
		b.WriteString(cursor + checked + " " + item + "\n")
	}
	if m.paginator.TotalPages != 0 {
		b.WriteString("  " + m.paginator.View() + "\n")
	}

	if m.active != 0 {
		// b.WriteString("------------------------------------------------\n")
		b.WriteString("( " + utils.UnderLine(m.GetInput()) + " )\n")
		// b.WriteString("-----------------------------------------------\n")
	}
	return b.String()
}

func (m *Notes) Reload() {
	// passwd := ""
	// m.keys = ""
	// m.cursor = 0
	// m.Msg = ""
	// m.textInput = textinput.NewModel()
	// m.After = make(chan func(s string) error, 2)
	// m.WithEdit("git pass:", func(passwd string) error {
	// 	filepath.Walk(PWDDIR, func(p string, f os.FileInfo, e error) error {
	// 		if strings.HasSuffix(p, ".en") {
	// 			if decry, err := ioutil.ReadFile(p); err != nil {
	// 				log.Println(p, "broken !")
	// 			} else {
	// 				if en, err := AesDecrypt(decry, []byte(passwd)); err != nil {
	// 					log.Println("Err pass or other:", err)
	// 				} else {
	// 					dict := make(map[string]interface{})
	// 					if err := json.Unmarshal(en, &dict); err != nil {
	// 						log.Println("Err buf !!", err)
	// 					}
	// 					for k, v := range dict {
	// 						m.Items[k] = v
	// 					}
	// 				}
	// 			}

	// 		}
	// 		return nil
	// 	})
	// 	return nil
	// })

	// m.Now = m.Items

	// m.active = 0
	// m.UpdateKeys()
}

func MainNote() string {
	GetRepo()
	passwd := ""
	pn := new(Notes)
	pn.Items = make(map[string]interface{})
	pn.Now = make(map[string]interface{})
	pn.After = make(map[string]func(s string) bool)

	filepath.Walk(PWDDIR, func(p string, f os.FileInfo, e error) error {
		if strings.HasSuffix(p, ".en") {
			if decry, err := ioutil.ReadFile(p); err != nil {
				log.Println(p, "broken !")
			} else {
				if passwd == "" {
					passwd = GetPass("cry :")
				}
				if en, err := AesDecrypt(decry, []byte(passwd)); err != nil {
					log.Println("Err pass or other:", err)
				} else {
					m := make(map[string]interface{})
					if err := json.Unmarshal(en, &m); err != nil {
						log.Println("Err buf !!", err)
					}
					// fmt.Println(m)
					for k, v := range m {
						pn.Items[k] = v
					}
				}
			}

		}
		return nil
	})
	pn.Now = pn.Items
	pn.UpdateKeys()
	p := tea.NewProgram(pn)
	// for {
	// for {
	if err := p.Start(); err != nil {
		// log.Pri(err)
		fmt.Println(pn.Choose)
		return pn.Choose
	}

	return pn.Choose
	// }

	// }

	// return pn
}
