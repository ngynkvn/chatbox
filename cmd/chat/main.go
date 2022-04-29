package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gerow/go-color"
	"github.com/gliderlabs/ssh"
)

const host = "0.0.0.0"
const port = 2222

func main() {
	ctx, cancel, subscribe, inbox := StartChatRoom()
	defer cancel()
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			bm.Middleware(teaHandler(subscribe, inbox)),
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

type ChatRoom struct {
	lines         []string
	mutex         sync.Mutex
	subscriptions []chan<- string
}

func StartChatRoom() (context.Context, context.CancelFunc, func() chan string, chan<- string) {
	inbox := make(chan string, 1024)
	chatRoom := ChatRoom{
		lines: []string{"Welcome!"},
	}
	go func() {
		for msg := range inbox {
			log.Printf("RECV: %s\n", msg)
			chatRoom.lines = append(chatRoom.lines, msg)
			chatRoom.mutex.Lock()
			log.Printf("LOCK: %d to send\n", len(chatRoom.subscriptions))
			for _, ch := range chatRoom.subscriptions {
				s := strings.Join(chatRoom.lines, "\n")
				ch <- s
			}
			chatRoom.mutex.Unlock()
			log.Printf("UNLOCK\n")
		}
	}()
	subscribe := func() chan string {
		ch := make(chan string, 1024)
		go func() {
			chatRoom.mutex.Lock()
			log.Println("[Subscribe] LOCK")
			chatRoom.subscriptions = append(chatRoom.subscriptions, ch)
			s := strings.Join(chatRoom.lines, "\n")
			ch <- s
			chatRoom.mutex.Unlock()
			log.Println("[Subscribe] UNLOCK")
		}()
		return ch
	}
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, cancel, subscribe, inbox
}

type Client struct {
	input    textinput.Model
	pull     chan string
	push     chan<- string
	username string
	polls    int
	buffer   string
	chat     string
	width    int
	height   int
}

func randomColor() lipgloss.Color {
	hsl := color.HSL{H: rand.Float64(), S: 0.7, L: 0.7}
	return lipgloss.Color("#" + hsl.ToHTML())
}

func NewClient(username string, pty ssh.Pty, push chan<- string, pull chan string) Client {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 128
	ti.Prompt = username + ": "
	ti.PromptStyle = ti.PromptStyle.Foreground(randomColor())
	ti.TextStyle = ti.TextStyle.Foreground(randomColor())
	return Client{
		width:    pty.Window.Width,
		height:   pty.Window.Height,
		username: username,
		input:    ti,
		pull:     pull,
		push:     push,
		buffer:   username + ": ",
	}
}

type ChatUpdate struct {
	chat string
}

func (m Client) Init() tea.Cmd {
	return tea.Batch(m.pollChat, textinput.Blink)
}

func (m Client) pollChat() tea.Msg {
	chat := <-m.pull
	return ChatUpdate{chat: chat}
}

func (m *Client) updateChat(chat string) (tea.Model, tea.Cmd) {
	m.chat = chat
	m.polls = m.polls + 1
	return m, m.pollChat
}

func (m Client) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ChatUpdate:
		return m.updateChat(msg.chat)
	case tea.KeyMsg:
		switch msg.Type {
		// quit on ctrl+c
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			m.input.SetCursorMode(textinput.CursorHide)
			m.push <- m.input.View()
			m.input.SetCursorMode(textinput.CursorBlink)
			m.input.Reset()
		}
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Client) View() string {
	var promptStyle = lipgloss.NewStyle().Width(m.width).Height(3)
	return fmt.Sprintf("==Secret Club %d==\n%s\n%s", m.polls, m.chat, promptStyle.Render(m.input.View()))
}

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// teaw.WithAltScreen) on a session by session basis
func teaHandler(subscribe func() chan string, push chan<- string) bm.Handler {
	handler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		username := s.User()
		pty, _, active := s.Pty()
		pull := subscribe()
		if !active {
			fmt.Println("no active terminal, skipping")
			return nil, nil
		}
		m := NewClient(username, pty, push, pull)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
	return handler
}
