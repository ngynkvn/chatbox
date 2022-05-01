package chat

import (
	"context"
	"log"
	"math/rand"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerow/go-color"
	"github.com/gliderlabs/ssh"
)

func NewClient(username string, pty ssh.Pty, push chan<- string, pull chan string) Client {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 128
	ti.Prompt = username + ": "
	ti.PromptStyle = ti.PromptStyle.Foreground(randomColor())
	ti.Width = 80
	ti.TextStyle = ti.TextStyle.Foreground(randomColor())
	return Client{
		width:    pty.Window.Width,
		height:   pty.Window.Height,
		username: username,
		input:    ti,
		pull:     pull,
		push:     push,
	}
}

func StartChatRoom() (context.Context, context.CancelFunc, func(username string) chan string, chan<- string) {
	inbox := make(chan string, 1024)
	chatRoom := ChatRoom{
		lines: []string{"Welcome!"},
	}
	go func() {
		for msg := range inbox {
			log.Printf("RECV: %s, %d chars long", msg, len(msg))
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
	subscribe := func(username string) chan string {
		ch := make(chan string, 1024)
		go func() {
			chatRoom.mutex.Lock()
			log.Println("[Subscribe] LOCK")
			chatRoom.subscriptions = append(chatRoom.subscriptions, ch)
			s := strings.Join(chatRoom.lines, "\n")
			inbox <- username + " has joined"
			ch <- s
			chatRoom.mutex.Unlock()
			log.Println("[Subscribe] UNLOCK")
		}()
		return ch
	}
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, cancel, subscribe, inbox
}

/**
 * Private Functions
 */
func randomColor() lipgloss.Color {
	hsl := color.HSL{H: rand.Float64(), S: 0.7, L: 0.7}
	return lipgloss.Color("#" + hsl.ToHTML())
}
