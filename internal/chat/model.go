package chat

import (
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
)

type ChatUpdate struct {
	chat string
}

type ChatRoom struct {
	lines         []string
	mutex         sync.Mutex
	subscriptions []chan<- string
}

type Client struct {
	input    textinput.Model
	pull     chan string
	push     chan<- string
	username string
	polls    int
	chat     string
	width    int
	height   int
}
