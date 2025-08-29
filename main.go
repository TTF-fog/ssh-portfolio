package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
	"sync"

	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	HOST = "0.0.0.0"
	PORT = "23849"
)

var uptime time.Time
var vCount int
var authToken string
var logger fileLogger
var stats UserStats
var dailyStats dailyUserStats
var hash []byte

func main() {
	f, _ := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	var err error
	hash, err = os.ReadFile("sha.txt")
	if err != nil {
		hash = []byte("3werwegwaq124414")
	}
	logger = fileLogger{
		file: f,
		lock: &sync.RWMutex{},
	}

	var port string
	var host string
	port = PORT
	host = HOST
	authToken, _ = os.LookupEnv("HACKATIME_API_KEY")
	cacheData(&stats, &dailyStats)
	go func() {
		for {
			cacheData(&stats, &dailyStats)
			time.Sleep(120 * time.Second)
		}
	}()
	key, found := os.LookupEnv("PORT")

	if found {
		port = key
	}
	key, found = os.LookupEnv("HOST")
	if found {
		host = key
	}
	uptime = time.Now().Truncate(time.Second)
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			portfolioInit(),
			trackUser(),
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
func trackUser() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(session ssh.Session) {
			fmt.Println(session.User(), session.RemoteAddr().String(), session.Environ())
			go incrementVisitsCounter(&vCount)
			go logger.log(fmt.Sprintf("visits: %d, user %s, ip %s", vCount, session.User(), session.RemoteAddr().String()))
			//hehehehhe
			next(session)

		}
	}
}

func portfolioInit() wish.Middleware {
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

		descs := make(map[string]string)
		items, _ := os.ReadDir("descs")
		for _, item := range items {
			data, _ := os.ReadFile("descs/" + item.Name())
			descs[item.Name()] = string(data)
		}
		listItems := loadFrameworks()
		blogItems := loadBlogs()
		ti := textinput.New()
		ti.Placeholder = "Your Name"
		ti.CharLimit = 156
		ti.Width = 20
		t2 := textinput.New()
		t2.Placeholder = "Your Email"
		t2.CharLimit = 156
		t2.Width = 20

		contContent := textarea.New()
		contContent.Placeholder = "Your Message"
		contContent.CharLimit = 156

		m := model{
			mainPage: mainPage{
				description: viewport.New(0, 0),
			},
			tabs: tabInterface{
				tabs: []string{"About Me", "My Skills", "Contact Me", "Blog", "Stats"},
				idx:  0,
			},
			mySkills: mySkills{
				frameworks:          list.New(listItems, itemDelegate{}, 0, 0),
				expandedDescription: viewport.New(0, 0),
			},
			contactMe: contactMe{name: ti, email: t2, content: contContent},
			noLifeStats: noLifeStats{
				allTimeStats: stats,
				dailyStats:   dailyStats,
			},
			blogPage: blog{
				Blogs:               list.New(blogItems, blogDelegate{}, 0, 0),
				expandedDescription: viewport.New(0, 0),
				contentFocused:      false,
			},
		}

		data, _ := os.ReadFile("about_me.md")

		m.content, m.aboutImages = parseMarkdownForImages(string(data))
		m.mySkills.frameworks.SetShowTitle(false)
		m.blogPage.Blogs.Title = "Blogs"
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
	content     string
	ready       bool
	width       int
	height      int
	time        time.Time
	tabs        tabInterface
	mainPage    mainPage
	mySkills    mySkills
	contactMe   contactMe
	noLifeStats noLifeStats
	blogPage    blog
	aboutImages []string
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
		availableHeight := msg.Height - 8
		listWidth := msg.Width / 3

		m.blogPage.Blogs.SetSize(listWidth, availableHeight)
		m.blogPage.expandedDescription.Width = msg.Width - listWidth - 6
		m.blogPage.expandedDescription.Height = availableHeight - 2
		if m.blogPage.contentFocused {
			m.blogPage.expandedDescription.Width = msg.Width
			m.blogPage.expandedDescription.Height = msg.Height - 1
		} else {
			m.blogPage.Blogs.SetSize(listWidth, availableHeight)
			m.blogPage.expandedDescription.Width = msg.Width - listWidth - 6
			m.blogPage.expandedDescription.Height = availableHeight - 2
		}

		m.contactMe.content.SetWidth(msg.Width - 30)
		m.contactMe.content.SetHeight(msg.Height / 2)

		if !m.ready {
			m.mainPage.description.SetContent(parseMarkdownAgainForImages(m.content, m.aboutImages, lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder()), m.mainPage.description.Width))
			m.mySkills.expandedDescription.SetContent("Try pressing enter :)")
			m.ready = true
		}

		m.mainPage.description, cmd = m.mainPage.description.Update(msg)
		cmds = append(cmds, cmd)
		m.mySkills.frameworks, cmd = m.mySkills.frameworks.Update(msg)
		cmds = append(cmds, cmd)
		m.blogPage.expandedDescription, cmd = m.blogPage.expandedDescription.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
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
		case "down":
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
		case "tab":
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
		case "f":
			if m.tabs.tabs[m.tabs.idx] == "My Skills" {
				m.mySkills.contentFocused = !m.mySkills.contentFocused
				return m, nil
			}
			if m.tabs.tabs[m.tabs.idx] == "Blog" {
				m.blogPage.contentFocused = !m.blogPage.contentFocused
				return m, nil
			}
		}

	}

	switch m.tabs.tabs[m.tabs.idx] {
	case "About Me":
		m.mainPage.description, cmd = m.mainPage.description.Update(msg)
		cmds = append(cmds, cmd)
	case "My Skills":
		if m.mySkills.contentFocused {
			m.mySkills.expandedDescription, cmd = m.mySkills.expandedDescription.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
				switch item := m.mySkills.frameworks.SelectedItem().(type) {
				case *Framework:
					md, images := parseMarkdownForImages(item.ExpandedDescriptionMD)
					m.mySkills.expandedDescription.SetContent(parseMarkdownAgainForImages(md, images, lipgloss.NewStyle(), m.mySkills.expandedDescription.Width))
				}
			} else {
				m.mySkills.frameworks, cmd = m.mySkills.frameworks.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	case "Blog":
		if m.blogPage.contentFocused {
			m.blogPage.expandedDescription, cmd = m.blogPage.expandedDescription.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			ey, _ := msg.(tea.KeyMsg)
			if ey.String() == "enter" {

				switch item := m.blogPage.Blogs.SelectedItem().(type) {
				case *Article:
					m.blogPage.expandedDescription.SetContent(parseMarkdownAgainForImages(item.Body, item.Images, lipgloss.NewStyle().Padding(1, 1), m.blogPage.expandedDescription.Width))
				}
			} else {
				m.blogPage.Blogs, cmd = m.blogPage.Blogs.Update(msg)
				cmds = append(cmds, cmd)
			}
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

	docStyle := lipgloss.NewStyle().Padding(1, 1).BorderStyle(lipgloss.NormalBorder())

	if m.width < 105 || m.height < 25 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Window too small! Please resize.")
	}

	tabs := m.tabs.View()
	stats := docStyle.Padding(0, 0).Render(fmt.Sprintf("Uptime: %s "+
		"Visits: %d "+" Git Hash: %s", time.Since(uptime).Truncate(time.Second).String(), vCount, hash))

	remainingWidth := m.width - lipgloss.Width(tabs)
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	statsPlaced := lipgloss.PlaceHorizontal(remainingWidth, lipgloss.Right, stats)
	tabView := lipgloss.JoinHorizontal(lipgloss.Top, tabs, statsPlaced)

	switch m.tabs.tabs[m.tabs.idx] {
	case "My Skills":
		return m.mySkills.View(tabView, m.height)
	case "Contact Me":
		return m.contactMe.View(tabView, m.width, m.height)
	case "Blog":
		if m.blogPage.contentFocused {
			indicator := lipgloss.PlaceHorizontal(m.width, lipgloss.Bottom, lipgloss.NewStyle().Reverse(true).Render(" Press 'f' to exit fullscreen "))
			return lipgloss.JoinVertical(lipgloss.Left, m.blogPage.expandedDescription.View(), indicator)
		}
		return m.blogPage.View(tabView, m.height, m.width)
	case "Stats":
		return m.noLifeStats.View(tabView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabView, "Tab / Shift+Tab to navigate", docStyle.AlignHorizontal(lipgloss.Left).Render(m.mainPage.description.View()))
}
