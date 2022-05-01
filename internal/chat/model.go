package chat

import (
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
)

type ChatUpdate struct {
	chat string
}

type ChatRoom struct {
	lines []string
	mutex sync.Mutex
	users map[string]chan<- string
	Inbox chan string
}

type Client struct {
	input    textinput.Model
	recv     chan string
	send     chan<- string
	username string
	polls    int
	chat     string
	width    int
	height   int
}
