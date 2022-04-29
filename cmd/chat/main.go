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

func main() {
	ctx, cancel, subscribe, inbox := chat.StartChatRoom()
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

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// teaw.WithAltScreen) on a session by session basis
func teaHandler(subscribe func(username string) chan string, push chan<- string) bm.Handler {
	handler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		username := s.User()
		pty, _, active := s.Pty()
		pull := subscribe(username)
		if !active {
			fmt.Println("no active terminal, skipping")
			return nil, nil
		}
		m := chat.NewClient(username, pty, push, pull)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
	return handler
}
