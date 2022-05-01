package chat

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/gerow/go-color"
	"github.com/gliderlabs/ssh"
)

func NewClient(username string, pty ssh.Pty, send chan<- string, recv chan Msg) Client {
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
		recv:     recv,
		send:     send,
	}
}

func (chatRoom *ChatRoom) withLock(tag string, f func()) {
	log.Printf("[🔒 %s] LOCK", tag)
	chatRoom.mutex.Lock()
	f()
	log.Printf("[🔒 %s] UNLOCK", tag)
	chatRoom.mutex.Unlock()
}

func (chatRoom *ChatRoom) Subscribe(username string) chan Msg {
	ch := make(chan Msg, 1024)
	go chatRoom.withLock("SUBSCRIBE", func() {
		_, ok := chatRoom.users[username]
		if ok {
			log.Println("[Subscribe] Already subscribed")
			close(ch)
			return
		} else {
			chatRoom.users[username] = ch
		}
		chatRoom.Inbox <- username + " has joined"
		ch <- MsgChat{chatRoom.history()}
	})
	return ch
}

func (chatRoom *ChatRoom) Unsubscribe(username string) {
	chatRoom.withLock("UNSUBSCRIBE", func() {
		delete(chatRoom.users, username)
	})
}
func (chatRoom *ChatRoom) history() string {
	return strings.Join(chatRoom.lines, "\n")
}

func (chatRoom *ChatRoom) SendAll(m Msg) {
	chatRoom.withLock("SendAll:"+m.Tag(), func() {
		log.Printf("--- 📤️ %d to send\n", len(chatRoom.users))
		for _, ch := range chatRoom.users {
			ch <- m
		}
	})
}

func logTime(tag string, f func()) {
	now := time.Now()
	f()
	after := time.Now()
	log.Printf("[⏱ %s] took %s", tag, after.Sub(now))
}

func StartChatRoom() (context.Context, context.CancelFunc, *ChatRoom) {
	chatRoom := ChatRoom{
		lines: []string{"Welcome!"},
		users: make(map[string]chan<- Msg),
		Inbox: make(chan string, 1024),
	}
	// Entry point for new messages from subscriptions.
	go func() {
		for msg := range chatRoom.Inbox {
			logTime("SendAll", func() {
				log.Printf("RECV: %s, %d chars long", msg, len(msg))
				chatRoom.lines = append(chatRoom.lines, msg)
				chat := MsgChat{chat: strings.Join(chatRoom.lines, "\n")}
				chatRoom.SendAll(chat)
			})
		}
	}()

	// Function to subscribe to the chat room.

	ctx, cancel := context.WithCancel(context.Background())
	return ctx, cancel, &chatRoom
}

/**
 * Private Functions
 */
func randomColor() lipgloss.Color {
	hsl := color.HSL{H: rand.Float64(), S: 0.7, L: 0.7}
	return lipgloss.Color("#" + hsl.ToHTML())
}
