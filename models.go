package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	_ "net/http/pprof"
	"strings"
	"time"
)

const (
	PADDING = 0
)

type Article struct {
	DatePublished time.Time `json:"DatePublished"`
	Name          string    `json:"Title,omitempty"`
	Desc          string    `json:"Description,omitempty"`
	Body          string    `json:"Body,omitempty"`
	Categories    []string  `json:"Categories,omitempty"`
}

func (i *Article) Title() string       { return i.Name }
func (i *Article) Description() string { return i.Desc }
func (i *Article) FilterValue() string { return i.Name }
func (*Article) Init() tea.Cmd         { return nil }

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

type visits struct {
	Visits int `json:"visits"`
}
type Theme struct {
	tabSelectColor     lipgloss.Color
	globalBorderColor  lipgloss.Color
	HighlightColor     lipgloss.Color
	textEmphasisColor  lipgloss.Color
	textEmphasisColor2 lipgloss.Color
}
