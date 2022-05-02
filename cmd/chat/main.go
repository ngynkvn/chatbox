package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	chat "github.com/ngynkvn/chatbox/internal/chat"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
)

const host = "0.0.0.0"
const port = 2222

type subscribeFn func(string) chan string

// Middleware for connecting to the chatroom.
func ChatSession(chat *chat.ChatRoom) wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			user := s.User()
			log.Printf("%s connected", user)
			sh(s)
			log.Println("Closing chatroom subscription for", user)
			chat.Unsubscribe(user)
			log.Printf("%#v", chat.GetUsers())
		}
	}
}

func main() {
	ctx, cancel, chatRoom := chat.StartChatRoom()
	defer cancel()
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			ChatSession(chatRoom),
			bm.Middleware(teaHandler(chatRoom)),
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

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// teaw.WithAltScreen) on a session by session basis
func teaHandler(chatRoom *chat.ChatRoom) bm.Handler {
	handler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		username := s.User()
		pty, _, active := s.Pty()
		if !active {
			fmt.Println("no active terminal, skipping")
			return nil, nil
		}
		recv := chatRoom.Subscribe(username)
		m := chat.NewClient(username, pty, chatRoom.Inbox, recv)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
	return handler
}
