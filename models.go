package main

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
	"strings"
)

const (
	PADDING = 0
)

type mainPage struct {
	description viewport.Model
}
type mySkills struct {
	frameworks          list.Model
	expandedDescription viewport.Model
}
type contactMe struct {
	name    textinput.Model
	email   textinput.Model
	content textarea.Model
}

func (contact *contactMe) Dump() {
	var data string
	data = fmt.Sprintf("Contact Name: %s\nContact Email: %s\n Content: %s\n", contact.name.Value(), contact.email.Value(), insertNth(contact.content.Value(), 40))
	//TODO: unsanitized paths could cause security vuln, fix
	err := os.WriteFile("messages/message-"+contact.name.Value(), []byte(data), 0644)
	if err != nil {
		panic(err)
	}
	contact.name.Reset()
	contact.email.Reset()
	contact.content.Reset()
}

type itemDelegate struct{}

func (d itemDelegate) Height() int { return 6 }

func (d itemDelegate) Spacing() int { return 0 }

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	switch item := listItem.(type) {
	case *Framework:
		fn := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("201")).Padding(PADDING).Render
		if index == m.Index() {
			fn = func(s ...string) string {
				return lipgloss.NewStyle().Margin(0, PADDING).
					BorderStyle(lipgloss.NormalBorder()).
					Foreground(lipgloss.Color("201")).
					Background(lipgloss.Color("235")).
					Render("> " + strings.Join(s, "\n "))
			}
		}
		println(item.Percent)
		str := fmt.Sprintf("%s \n %s ... \n %s", item.Title(), item.Desc, item.progress.ViewAs(item.Percent))
		fmt.Fprint(w, fn(str))
	}

}

type tabInterface struct {
	tabs []string
	idx  int
}

func (d tabInterface) View() string {
	var tabView []string
	for index, tab := range d.tabs {
		if index == d.idx {
			tabStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("120")).Padding(0, 1)
			tabView = append(tabView, tabStyle.Render(tab))
		} else {
			tabStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("240")).Padding(0, 1)
			tabView = append(tabView, tabStyle.Render(tab))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, tabView...)
}

type Framework struct {
	Name                  string `json:"Name,omitempty"`
	Desc                  string `json:"Desc,omitempty"`
	progress              progress.Model
	ExpandedDescriptionMD string  `json:"ExpandedDescriptionMD,omitempty"`
	Percent               float64 `json:"Percent,omitempty"`
}

func (i *Framework) Title() string       { return i.Name }
func (i *Framework) Description() string { return i.Desc }
func (i *Framework) FilterValue() string { return i.Name }
func (*Framework) Init() tea.Cmd         { return nil }

func insertNth(s string, n int) string {
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune('\n')
		}
	}
	return buffer.String()
}
