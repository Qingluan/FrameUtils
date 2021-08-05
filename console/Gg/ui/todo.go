package ui

// A simple program that opens the alternate screen buffer then counts down
// from 5 and then exits.

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
	tea "github.com/charmbracelet/bubbletea"
)

type Val struct {
	num int
}

type model struct {
	choices  []string    // items on the to-do list
	cursor   int         // which to-do list item our cursor is pointing at
	selected map[int]Val // which to-do items are selected
	choiced  []string
	Path     string
	msg      string
	action   time.Time
	Do       func(choiced []string)
}

type tickMsg time.Time

func NewModel(items ...string) (m *model) {
	m = new(model)

	// Our to-do list is just a grocery list

	// A map which indicates which choices are selected. We're using
	// the  map like a mathematical set. The keys refer to the indexes
	// of the `choices` slice, above.
	m.selected = make(map[int]Val)
	m.choices = items
	return
}

func (m *model) Add(item string) {
	m.choices = append(m.choices, item)
}

func (m *model) SetItems(items ...string) {
	m.choices = items
	m.selected = make(map[int]Val)
	m.choiced = []string{}
}

func (m *model) Remove(ix int) {
	m.choices = append(m.choices[:ix], m.choices[ix+1:]...)
}

func (m *model) Vim() *model {
	if m.Path == "" {
		return m
	}
	cmd := exec.Command("vim", m.Path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	return ReadToDo(m.Path)

}

func StartModel(m *model) {
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		// os.Exit(1)
	}
}

func Main(file string) {
	m := ReadToDo(file)
	if m != nil {
		p := tea.NewProgram(m)
		if err := p.Start(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	} else {
		m = new(model)
		if strings.HasSuffix(file, ".todo") {
			m.Path = file
		} else {
			m.Path = filepath.Join(file, ".todo")
		}
		m = m.Vim()
		p := tea.NewProgram(m)
		if err := p.Start(); err != nil {
			// fmt.Printf("Alas, there's been an error: %v", err)
			// os.Exit(1)
		}
	}

}

func (m *model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

			if m.choices[m.cursor] == "---" {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
			if m.choices[m.cursor] == "---" {
				m.cursor++
			}
		case "v":
			m = m.Vim()
		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case " ", "l":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = Val{}
			}
			pre := strings.Count(m.choices[m.cursor], "\t")
			i := m.cursor
			for i < len(m.choices)-1 {
				nextchoo := m.choices[i+1]
				if thispre := strings.Count(nextchoo, "\t"); thispre > pre {
					if _, ok := m.selected[i+1]; ok {
						delete(m.selected, i+1)
					} else {
						// w, _ := m.selected[m.cursor]
						// w.num = i + 1 - pre
						m.selected[i+1] = Val{}
					}
				} else {
					break
				}
				i += 1
			}
			// m.selected[m.cursor] = Val{num: i + 1 - pre}

		case "enter":
			for i, v := range m.choices {
				if _, ok := m.selected[i]; ok {
					m.choiced = append(m.choiced, v)
				}
			}
			if m.Do != nil {
				m.Do(m.choiced)

				m.choiced = []string{}
			}
			m.Save()

		}

	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m *model) Save() {
	if m.Path == "" {
		return
	}
	m.action = time.Now()
	m.msg = time.Now().String()
	if _, err := os.Stat(m.Path); err == nil {
		os.Remove(m.Path)
	}
	fp, err := os.Create(m.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	for i, v := range m.choices {
		if strings.Contains(v, "\t") {
			v = strings.ReplaceAll(v, "\t", "- ")
		}
		if _, ok := m.selected[i]; ok {
			fp.WriteString(" x " + v + "\n")
		} else {
			fp.WriteString(" - " + v + "\n")
		}
	}
	m.msg = "Save in " + m.Path + "\ntime:" + time.Now().String()
}

func (m *model) View() string {
	// The header
	s := fmt.Sprintf("Todo ?%s\n\n", m.msg)
	if time.Since(m.action) > 4*time.Second {
		m.msg = ""
		// m.action = time
	}
	// Iterate over our choice
	// groupContains := false
	// groupFinish := false
	for i, choice := range m.choices {
		if choice == "---" {
			s += "—————————————————————————\n"
			continue
		}
		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if c, ok := m.selected[i]; ok {
			if c.num > 0 {
				checked = utils.Green(c.num) // selected!
			} else {
				checked = utils.Green("■") // selected!

			}
			// checked = utils.BBlue("✔")
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress 'v' to edit // enter to save //  'q'  to quit.\n"

	// Send the UI for rendering
	return s
}
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
func ReadToDo(root string) *model {
	if strings.HasSuffix(root, ".todo") {
		buf, err := ioutil.ReadFile(root)
		if err != nil {
			return nil
		}
		items := []string{}
		finished := []int{}

		for c, line := range strings.Split(string(buf), "\n") {
			l := strings.TrimSpace(line)
			if strings.HasPrefix(l, "- ") {
				// fmt.Println("ddd")
				items = append(items, strings.ReplaceAll(l[2:], "- ", "\t"))
			} else if strings.HasPrefix(l, "x ") {
				finished = append(finished, c)
				// fmt.Println("ccc")
				items = append(items, strings.ReplaceAll(l[2:], "- ", "\t"))
			} else if strings.HasPrefix(l, "---") {
				// fmt.Println("ddd")
				items = append(items, "---")
			}
		}

		m := NewModel(items...)

		m.Path = root
		for _, c := range finished {

			m.selected[c] = Val{}
		}
		return m
	} else {
		buf, err := ioutil.ReadFile(filepath.Join(root, ".todo"))
		if err != nil {
			return nil
		}
		items := []string{}
		finished := []int{}
		for c, line := range strings.Split(string(buf), "\n") {
			l := strings.TrimSpace(line)
			if strings.HasPrefix(l, "- ") {
				items = append(items, strings.ReplaceAll(l[2:], "- ", "\t"))
			} else if strings.HasPrefix(l, "x ") {
				finished = append(finished, c)
				items = append(items, strings.ReplaceAll(l[2:], "- ", "\t"))

			}
		}

		m := NewModel(items...)
		m.Path = filepath.Join(root, ".todo")
		for _, c := range finished {

			m.selected[c] = Val{}
		}
		return m
	}
}
