package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"os"
	"sort"
	"time"
)

type mainPage struct {
	description viewport.Model
}
type mySkills struct {
	frameworks          list.Model
	expandedDescription viewport.Model
	contentFocused      bool
}
type contactMe struct {
	name    textinput.Model
	email   textinput.Model
	content textarea.Model
}

const (
	FONT_WIDTH  = 36
	FONT_HEIGHT = 72
	N_CHANNELS  = 4
)

func (c *contactMe) View(TabView string, width int, height int) string {
	docStyle := lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "250"})
	render := docStyle.Render(lipgloss.JoinVertical(lipgloss.Center, c.name.View(), c.email.View(), c.content.View()))
	return lipgloss.JoinVertical(lipgloss.Center, TabView, "write me a message here and i'll (probably) get back to you \n press esc to escape and enter to submit, use arrow key to navigate", lipgloss.Place(width, height-40, lipgloss.Center, lipgloss.Center, render))
}
func (m *mySkills) View(TabView string, height int) string {
	docStyle := lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "250"})
	if m.contentFocused {
		m.expandedDescription.Height = height - 10
		return lipgloss.JoinVertical(lipgloss.Center, TabView, "press f to unfocus", docStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, m.expandedDescription.View())))
	}
	return lipgloss.JoinVertical(lipgloss.Center, TabView, "press f to focus text view (allows scrolling)", m.frameworks.View(), docStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, m.expandedDescription.View())))
}

type noLifeStats struct {
	allTimeStats      UserStats
	dailyStats        dailyUserStats
	languageBreakdown []*progress.Model
	projects          table.Model
}
type dailyUserStats struct {
	Text string `json:"text"`
}

func (m *noLifeStats) View(TabView string) string {
	docStyle := lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "250"})
	var languageStack string
	sort.Slice(m.allTimeStats.Languages, func(i, j int) bool {
		return m.allTimeStats.Languages[i].TotalSeconds > m.allTimeStats.Languages[j].TotalSeconds
	})
	for ind, l := range m.allTimeStats.Languages {
		if l.Name == "JSON" {
			l.Name = "C++"
		}

		languageStack += fmt.Sprintf("\n %d. %s for %s \n", ind+1, l.Name, l.Text)
		languageStack += docStyle.Padding(1).Render(progress.New().ViewAs(l.Percent / 100))
		if ind == 4 {
			break
		}
	}
	if m.dailyStats.Text == "Start coding to track your time" {
		return lipgloss.JoinVertical(lipgloss.Center, TabView, docStyle.Align(lipgloss.Center).Render(fmt.Sprintf("I have worked for %s, since the First of June, 2025\n %s", m.allTimeStats.HumanReadableTotal, languageStack)))
	}
	return lipgloss.JoinVertical(lipgloss.Center, TabView, docStyle.Align(lipgloss.Center).Render(fmt.Sprintf("I have worked for %s, since the First of June, 2025\n Today, i have coded for %s \n %s", m.allTimeStats.HumanReadableTotal, m.dailyStats.Text, languageStack)))
}

type UserStats struct {
	Username              string     `json:"username"`
	UserID                string     `json:"user_id"`
	Start                 time.Time  `json:"start"`
	End                   time.Time  `json:"end"`
	Range                 string     `json:"range"`
	HumanReadableRange    string     `json:"human_readable_range"`
	TotalSeconds          int64      `json:"total_seconds"`
	DailyAverageSeconds   int64      `json:"daily_average"`
	HumanReadableTotal    string     `json:"human_readable_total"`
	HumanReadableDailyAvg string     `json:"human_readable_daily_average"`
	Languages             []Language `json:"languages"`
}

type Language struct {
	Name         string  `json:"name"`
	TotalSeconds int64   `json:"total_seconds"`
	Text         string  `json:"text"`
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Percent      float64 `json:"percent"`
	Digital      string  `json:"digital"`
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

type blog struct {
	Blogs               list.Model
	expandedDescription viewport.Model
	contentFocused      bool
}

func (b *blog) View(TabView string, height int, width int) string {
	docStyle := lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "250"})
	if b.contentFocused {
		return lipgloss.JoinVertical(lipgloss.Center, docStyle.Render(b.expandedDescription.View()))
	}
	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		b.Blogs.View(),
		docStyle.Render(b.expandedDescription.View()),
	)

	return lipgloss.JoinVertical(lipgloss.Center,
		TabView,
		"preview - press f to open",
		mainView,
	)
}
