package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"
)

const (
	PADDING = 2
)

type itemDelegate struct{}

func (d itemDelegate) Height() int { return 6 }

func (d itemDelegate) Spacing() int { return 3 }

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
		str := fmt.Sprintf("%s \n %s ... \n %s", item.Title(), item.description, item.progress.ViewAs(item.percent))
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
	name        string
	description string
	progress    progress.Model
	percent     float64
}

func (i *Framework) Title() string       { return i.name }
func (i *Framework) Description() string { return i.description }
func (i *Framework) FilterValue() string { return i.name }
func (*Framework) Init() tea.Cmd         { return nil }
