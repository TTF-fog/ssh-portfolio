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
	DatePublished time.Time
	Name          string
	Desc          string
	Body          string
	Categories    []string
	Images        []string
}

func (i *Article) Title() string       { return i.Name }
func (i *Article) Description() string { return i.Desc }
func (i *Article) FilterValue() string { return i.Name }
func (*Article) Init() tea.Cmd         { return nil }
func (i *Article) getFormattedData() string {
	categoryStyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#5F00AF", Dark: "127"}).Padding(0, 1)
	var drawn_categories string
	if len(i.Categories) > 0 {
		drawn_categories = categoryStyle.Render(strings.Join(i.Categories, " | "))
	}
	if time.Since(i.DatePublished).Hours() > 336 {

		return fmt.Sprintf("%s ⏲ %s \n %s \n %s", i.Name, i.DatePublished.String(), i.Desc, drawn_categories)
	} else {
		return fmt.Sprintf("%s  ⏲ %s Ago \n %s \n %s", i.Name, time.Since(i.DatePublished).Truncate(time.Second), i.Desc, drawn_categories)
	}

}

type itemDelegate struct{}
type blogDelegate struct{}

func (b blogDelegate) Height() int {
	return 6
}
func (b blogDelegate) Spacing() int {
	return 0
}
func (b blogDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (b blogDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	fn := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "201"}).Padding(PADDING).Render
	switch item := listItem.(type) {
	case *Article:
		if index == m.Index() {
			fn = func(s ...string) string {
				return lipgloss.NewStyle().Margin(0, PADDING).
					BorderStyle(lipgloss.NormalBorder()).
					Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "201"}).
					Background(lipgloss.AdaptiveColor{Light: "#E5E5E5", Dark: "235"}).
					Render("> " + strings.Join(s, "\n "))
			}
		}
		str := item.getFormattedData()
		fmt.Fprint(w, fn(str))
	}
}
func (d itemDelegate) Height() int { return 6 }

func (d itemDelegate) Spacing() int { return 0 }

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	switch item := listItem.(type) {
	case *Framework:
		fn := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "201"}).Padding(PADDING).Render
		if index == m.Index() {
			fn = func(s ...string) string {
				return lipgloss.NewStyle().Margin(0, PADDING).
					BorderStyle(lipgloss.NormalBorder()).
					Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "201"}).
					Background(lipgloss.AdaptiveColor{Light: "#E5E5E5", Dark: "235"}).
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
			tabStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#008700", Dark: "120"}).Padding(0, 1)
			tabView = append(tabView, tabStyle.Render(tab))
		} else {
			tabStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#5F5F5F", Dark: "240"}).Padding(0, 1)
			tabView = append(tabView, tabStyle.Render(tab))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, tabView...)
}

type Framework struct {
	Name                      string `json:"Name,omitempty"`
	Desc                      string `json:"Desc,omitempty"`
	progress                  progress.Model
	ExpandedDescriptionMD     string  `json:"ExpandedDescriptionMD,omitempty"`
	ExpandedDescriptionMDFile string  `json:"ExpandedDescriptionMDFile,omitempty"`
	Percent                   float64 `json:"Percent,omitempty"`
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
