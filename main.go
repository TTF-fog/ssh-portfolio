package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	host = "0.0.0.0"
	port = "23849"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			myCustomBubbleteaMiddleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}

}
func myCustomBubbleteaMiddleware() wish.Middleware {
	newProg := func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
		p := tea.NewProgram(m, opts...)
		go func() {
			for {
				<-time.After(1 * time.Second)
				p.Send(timeMsg(time.Now()))
			}
		}()
		return p
	}

	teaHandler := func(s ssh.Session) *tea.Program {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}
		lipgloss.SetHasDarkBackground(true)
		_ = os.Setenv("COLORTERM", "truecolor")
		_ = os.Setenv("TERM", "xterm-256color")
		descs := make(map[string]string)
		items, _ := os.ReadDir("descs")
		for _, item := range items {
			data, _ := os.ReadFile("descs/" + item.Name())
			descs[item.Name()] = string(data)
		}
		list_items := loadFrameworks()

		ti := textinput.New()
		ti.Placeholder = "Your Name"
		ti.CharLimit = 156
		ti.Width = 20
		t2 := textinput.New()
		t2.Placeholder = "Your Email"
		t2.CharLimit = 156
		t2.Width = 20

		cont_content := textarea.New()
		cont_content.Placeholder = "Your Message"
		cont_content.CharLimit = 156
		m := model{
			mainPage: mainPage{
				description: viewport.New(0, 0),
			},
			tabs: tabInterface{
				tabs: []string{"About Me", "My Skills", "Contact Me"},
				idx:  0,
			},
			mySkills: mySkills{
				frameworks:          list.New(list_items, itemDelegate{}, 0, 0),
				expandedDescription: viewport.New(0, 0),
			},
			contactMe: contactMe{name: ti, email: t2, content: cont_content},
		}

		data, _ := os.ReadFile("artichoke.md")
		m.content = string(data)
		m.mySkills.frameworks.SetShowTitle(false)
		m.mySkills.frameworks.SetShowHelp(true)
		opts := append(
			bubbletea.MakeOptions(s),
			tea.WithAltScreen(),
		)
		lipgloss.SetColorProfile(termenv.TrueColor)
		return newProg(m, opts...)
	}

	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.TrueColor)
}

type model struct {
	content   string
	ready     bool
	width     int
	height    int
	time      time.Time
	tabs      tabInterface
	mainPage  mainPage
	mySkills  mySkills
	contactMe contactMe
}

type timeMsg time.Time

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case timeMsg:
		m.time = time.Time(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.mainPage.description.Width = msg.Width - 6
		m.mainPage.description.Height = msg.Height - 8

		m.mySkills.frameworks.SetSize(msg.Width-6, msg.Height/2-4)
		m.mySkills.expandedDescription.Width = msg.Width - 6
		m.mySkills.expandedDescription.Height = msg.Height/2 - 8

		m.contactMe.content.SetWidth(msg.Width - 30)
		m.contactMe.content.SetHeight(msg.Height / 2)

		if !m.ready {
			renderer, _ := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(m.mainPage.description.Width))
			str, _ := renderer.Render(m.content)
			m.mainPage.description.SetContent(str)
			m.mySkills.expandedDescription.SetContent("Try pressing enter :)")
			m.ready = true
		}

		m.mainPage.description, cmd = m.mainPage.description.Update(msg)
		cmds = append(cmds, cmd)
		m.mySkills.frameworks, cmd = m.mySkills.frameworks.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.tabs.tabs[m.tabs.idx] == "Contact Me" {
				if m.contactMe.name.Focused() {
					m.contactMe.name.Blur()
					m.contactMe.email.Focus()
					return m, nil
				} else if m.contactMe.email.Focused() {
					m.contactMe.email.Blur()
					m.contactMe.content.Focus()
					return m, nil
				} else if m.contactMe.content.Focused() {
					m.contactMe.name.Focus()
					m.contactMe.content.Blur()
					return m, nil
				}
			}
			if m.tabs.idx < len(m.tabs.tabs)-1 {
				m.tabs.idx++
				if m.tabs.tabs[m.tabs.idx] == "Contact Me" {
					m.contactMe.name.Focus()
				} else {
					m.contactMe.name.Blur()
				}
			}
			return m, nil
		case "shift+tab":
			if m.tabs.tabs[m.tabs.idx] == "Contact Me" {
				if m.contactMe.name.Focused() {
					m.contactMe.name.Blur()
					m.contactMe.content.Focus()
					return m, nil
				} else if m.contactMe.email.Focused() {
					m.contactMe.email.Blur()
					m.contactMe.name.Focus()
					return m, nil
				} else if m.contactMe.content.Focused() {
					m.contactMe.email.Focus()
					m.contactMe.content.Blur()
					return m, nil
				}
			}
			if m.tabs.idx > 0 {
				m.tabs.idx--
				if m.tabs.tabs[m.tabs.idx] == "Contact Me" {
					m.contactMe.name.Focus()
				} else {
					m.contactMe.name.Blur()
				}
			}

			return m, nil
		case "enter":
			if m.tabs.tabs[m.tabs.idx] == "Contact Me" {
				m.contactMe.Dump()
				return m, nil
			}
		case "esc":
			m.tabs.idx = 0
		}

	}

	switch m.tabs.tabs[m.tabs.idx] {
	case "About Me":
		m.mainPage.description, cmd = m.mainPage.description.Update(msg)
		cmds = append(cmds, cmd)
	case "My Skills":
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			renderer, _ := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(m.mainPage.description.Width))
			switch item := m.mySkills.frameworks.SelectedItem().(type) {
			case *Framework:
				str, _ := renderer.Render(item.ExpandedDescriptionMD)
				m.mySkills.expandedDescription.SetContent(str)
			}
		} else {
			m.mySkills.frameworks, cmd = m.mySkills.frameworks.Update(msg)
			cmds = append(cmds, cmd)
		}
	case "Contact Me":
		m.contactMe.name, cmd = m.contactMe.name.Update(msg)
		cmds = append(cmds, cmd)
		m.contactMe.email, cmd = m.contactMe.email.Update(msg)
		cmds = append(cmds, cmd)
		m.contactMe.content, cmd = m.contactMe.content.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
func (m model) View() string {
	docStyle := lipgloss.NewStyle().Padding(1, 2).BorderStyle(lipgloss.NormalBorder()).Foreground(lipgloss.Color("250"))
	switch m.tabs.tabs[m.tabs.idx] {
	case "My Skills":
		return lipgloss.JoinVertical(lipgloss.Center, m.tabs.View(), m.mySkills.frameworks.View(), docStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, m.mySkills.expandedDescription.View())))
	case "Contact Me":
		render := docStyle.Render(lipgloss.JoinVertical(lipgloss.Center, m.contactMe.name.View(), m.contactMe.email.View(), m.contactMe.content.View()))
		return lipgloss.JoinVertical(lipgloss.Center, m.tabs.View(), "write me a message here and i'll (probably) get back to you \n press esc to escape and enter to submit", lipgloss.Place(m.width, m.height-40, lipgloss.Center, lipgloss.Center, render))
	}

	return lipgloss.JoinVertical(lipgloss.Center, m.tabs.View(), docStyle.Copy().AlignHorizontal(lipgloss.Center).Render(m.mainPage.description.View()))
}

func loadFrameworks() []list.Item {
	items, _ := os.ReadDir("descs")
	var frameworks []list.Item
	for _, item := range items {
		name := item.Name()
		dat, _ := os.ReadFile(filepath.Join("descs", name))
		var framework Framework
		json.Unmarshal(dat, &framework)
		framework.progress = progress.New()
		frameworks = append(frameworks, &framework)

	}
	fmt.Println(frameworks)
	return frameworks
}
