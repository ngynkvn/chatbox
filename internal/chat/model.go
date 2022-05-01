package chat

import (
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
)

type RecvMsg struct {
	Msg
}

type ChatRoom struct {
	lines []string
	mutex sync.Mutex
	users map[string]chan<- Msg
	Inbox chan string
}

type Client struct {
	input    textinput.Model
	recv     chan Msg
	send     chan<- string
	username string
	users    []string
	polls    int
	chat     string
	width    int
	height   int
}

type Msg interface {
	Tag() string
}

type MsgChat struct {
	chat string
}

func (m MsgChat) Tag() string {
	return "CHAT"
}

type MsgUserList struct {
	users []string
}

func (m MsgUserList) Tag() string {
	return "USERLIST"
}
