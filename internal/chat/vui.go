package chat

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Client) Init() tea.Cmd {
	return tea.Batch(m.pollChat, textinput.Blink)
}

func (m Client) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case RecvMsg:
		return m.handleRecvMsg(msg)
	case tea.KeyMsg:
		switch msg.Type {
		// quit on ctrl+c
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			i := m.input
			m.send <- i.PromptStyle.Render(i.Prompt) + i.TextStyle.Render(i.Value())
			m.input.Reset()
		}
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

const cat = "猫咪"

// Style definitions.
var (

	// General.

	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
)

func (m Client) View() string {
	// Given the terminal size, print the chat and user list
	w, h := m.width, m.height

	var chatWindowStyle = lipgloss.NewStyle().
		MaxWidth(w).
		MaxHeight(h).
		PaddingLeft(1).
		Foreground(subtle).
		Border(lipgloss.RoundedBorder())
	fw, fh := chatWindowStyle.GetFrameSize()
	w, h = w-fw, h-fh

	// User list render
	// var userHeaderStyle = lipgloss.NewStyle().
	// 	Bold(true).
	// 	BorderStyle(lipgloss.ThickBorder()).
	// 	BorderBottom(true)
	var userListStyle = lipgloss.NewStyle().
		Width(12).
		BorderForeground(subtle).
		Border(lipgloss.RoundedBorder())
	var header = lipgloss.NewStyle().
		Bold(true).
		Foreground(special).
		SetString("Users\n=====")
	userList := userListStyle.Render(header.String() + "\nme\nfriendo\nstevenn\ngato")
	// Render the chatbox container
	var chatContentStyle = lipgloss.NewStyle().
		Width(w - fw - userListStyle.GetWidth()).
		PaddingTop(1).
		Height(h - fh)
	chat := chatContentStyle.Render(lipgloss.JoinVertical(0, m.chat, m.input.View()))

	return chatWindowStyle.Render(lipgloss.JoinHorizontal(0, chat, userList))

}

/**
 * Private Functions
 */

func (m Client) pollChat() tea.Msg {
	chat := <-m.recv
	return RecvMsg{Msg: chat}
}

func (m *Client) handleRecvMsg(msg RecvMsg) (tea.Model, tea.Cmd) {
	switch msg := msg.Msg.(type) {
	case MsgChat:
		m.chat = msg.chat
	case MsgUserList:
		m.users = msg.users
	}
	m.polls = m.polls + 1
	return m, m.pollChat
}
