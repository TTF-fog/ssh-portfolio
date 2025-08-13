package main

import (
	"context"
	"errors"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
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
	"syscall"
	"time"
)

const (
	host = "localhost"
	port = "23841"
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
		cpp_desc, _ := os.ReadFile("descs/cpp_desc.md")
		go_desc, _ := os.ReadFile("descs/go_desc.md")
		list_items := []list.Item{
			&Framework{
				name:                  "Go",
				description:           "Language i learnt recently, which i rely heavily on for app development (what you see is all Go!)",
				expandedDescriptionMD: string(go_desc),
				progress:              progress.New(),
				percent:               70,
			}, &Framework{
				name:                  "C++",
				description:           "One of the languages i have more experience with, used by me primarily for the Arduino platform",
				expandedDescriptionMD: string(cpp_desc),
				progress:              progress.New(),
				percent:               78,
			},
		}
		m := model{mainPage: mainPage{
			description: viewport.New(0, 0),
		}, tabs: tabInterface{
			tabs: []string{"About Me", "My Skills"},
			idx:  0,
		}, mySkills: mySkills{frameworks: list.New(list_items, itemDelegate{}, 0, 0), expandedDescription: viewport.New(0, 0)},
		}
		data, _ := os.ReadFile("artichoke.md")
		m.content = string(data)
		m.mySkills.frameworks.SetShowTitle(false)
		m.mySkills.frameworks.SetShowHelp(true)

		return newProg(m, append(bubbletea.MakeOptions(s), tea.WithAltScreen())...)
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

type mainPage struct {
	description viewport.Model
}
type mySkills struct {
	frameworks          list.Model
	expandedDescription viewport.Model
}
type model struct {
	content  string
	ready    bool
	time     time.Time
	tabs     tabInterface
	mainPage mainPage
	mySkills mySkills
}

type timeMsg time.Time

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timeMsg:
		m.time = time.Time(msg)
	case tea.WindowSizeMsg:
		if !m.ready {
			m.mainPage.description.Width = msg.Width / 2
			m.mainPage.description.Height = msg.Height - 4
			m.mySkills.expandedDescription.Height = msg.Height - 4
			m.mySkills.expandedDescription.Width = msg.Width / 6
			m.ready = true
		} else {
			m.mainPage.description.Width = msg.Width - 6
			m.mainPage.description.Height = msg.Height - 4
			m.mySkills.expandedDescription.Width = msg.Width - 6
			m.mySkills.expandedDescription.Height = msg.Height - 30
		}

		m.mainPage.description.Height = msg.Height - 8
		m.mySkills.expandedDescription.Height = msg.Height/2 - 8
		m.mySkills.frameworks.SetSize(msg.Width-6, msg.Height/2-4)
		renderer, _ := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(m.mainPage.description.Width))
		str, _ := renderer.Render(m.content)
		m.mainPage.description.SetContent(str)
		m.mySkills.expandedDescription.SetContent("Try pressing enter :)")
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.tabs.idx == len(m.tabs.tabs)-1 {
				break
			}
			m.tabs.idx += 1
		case "enter":
			if m.tabs.tabs[m.tabs.idx] == "My Skills" {
				renderer, _ := glamour.NewTermRenderer(glamour.WithStandardStyle("dark"), glamour.WithWordWrap(m.mainPage.description.Width))
				switch item := m.mySkills.frameworks.SelectedItem().(type) {
				case *Framework:
					str, _ := renderer.Render(item.expandedDescriptionMD)
					m.mySkills.expandedDescription.SetContent(str)
				}

			}
		case "shift+tab":
			if m.tabs.idx == 0 {
				break
			}
			m.tabs.idx -= 1
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.mainPage.description, cmd = m.mainPage.description.Update(msg)
	m.mySkills.frameworks, cmd = m.mySkills.frameworks.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
func (m model) View() string {
	docStyle := lipgloss.NewStyle().Padding(1, 2).BorderStyle(lipgloss.NormalBorder())
	if m.tabs.tabs[m.tabs.idx] == "My Skills" {
		return lipgloss.JoinVertical(lipgloss.Center, m.tabs.View(), m.mySkills.frameworks.View(), docStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, m.mySkills.expandedDescription.View())))
	}
	return lipgloss.JoinVertical(lipgloss.Center, m.tabs.View(), docStyle.Copy().AlignHorizontal(lipgloss.Center).Render(m.mainPage.description.View()))
}
