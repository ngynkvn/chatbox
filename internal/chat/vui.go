package chat

import (
	"fmt"

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

/**
 * Private Functions
 */

func (m Client) pollChat() tea.Msg {
	chat := <-m.pull
	return ChatUpdate{chat: chat}
}

func (m *Client) updateChat(chat string) (tea.Model, tea.Cmd) {
	m.chat = chat
	m.polls = m.polls + 1
	return m, m.pollChat
}
