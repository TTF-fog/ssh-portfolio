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
}
type contactMe struct {
	name    textinput.Model
	email   textinput.Model
	content textarea.Model
}

func (c *contactMe) View(TabView string, width int, height int) string {
	docStyle := lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("250"))
	render := docStyle.Render(lipgloss.JoinVertical(lipgloss.Center, c.name.View(), c.email.View(), c.content.View()))
	return lipgloss.JoinVertical(lipgloss.Center, TabView, "write me a message here and i'll (probably) get back to you \n press esc to escape and enter to submit, use arrow key to navigate", lipgloss.Place(width, height-40, lipgloss.Center, lipgloss.Center, render))
}
func (m *mySkills) View(TabView string) string {
	docStyle := lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("250"))
	return lipgloss.JoinVertical(lipgloss.Center, TabView, m.frameworks.View(), docStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, m.expandedDescription.View())))
}

type noLifeStats struct {
	allTimeStats      UserStats
	dailyStats        UserStats
	languageBreakdown []*progress.Model
	projects          table.Model
}

func (m *noLifeStats) View(TabView string) string {
	docStyle := lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("250"))
	var languageStack string
	sort.Slice(m.allTimeStats.Languages, func(i, j int) bool {
		return m.allTimeStats.Languages[i].TotalSeconds > m.allTimeStats.Languages[j].TotalSeconds
	})
	for ind, l := range m.allTimeStats.Languages {
		if l.Name == "JSON" {
			l.Name = "C++"
		}
		languageStack += fmt.Sprintf("%d. %s for %s \n", ind+1, l.Name, l.Text)
		if ind == 4 {
			break
		}
	}
	return lipgloss.JoinVertical(lipgloss.Center, TabView, docStyle.Align(lipgloss.Center).Render(fmt.Sprintf("I have worked for %s, since the First of June, 2025\n %s", m.allTimeStats.HumanReadableTotal, languageStack)))
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
	frameworks          list.Model
	expandedDescription viewport.Model
}
