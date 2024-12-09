package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"os"

	"github.com/charmbracelet/bubbletea"
)

var ColorBlue = lipgloss.Color("12")
var ColorWhite = lipgloss.Color("255")

type model struct {
	boxes      []string
	focus      int
	fullWidth  int
	fullHeight int
}

func initialModel() model {
	return model{
		focus: 0,
		boxes: []string{"Inbox", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.fullHeight = msg.Height
		m.fullWidth = msg.Width

	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.focus == 7 {
				m.focus = 0
			} else {
				m.focus++
			}
		case "shift+tab":
			if m.focus == 0 {
				m.focus = 7
			} else {
				m.focus--
			}
		}
	}

	return m, nil
}

func (m model) boxSize() (int, int) {
	return m.fullWidth / 4, m.fullHeight / 2
}

func generateBorder() lipgloss.Border {
	return lipgloss.RoundedBorder()
}

func (m model) boxStyle() lipgloss.Style {
	w, h := m.boxSize()
	return lipgloss.NewStyle().
		Border(generateBorder()).
		Width(w - 2).
		Height(h - 2)
}

func (m model) View() string {
	var style = m.boxStyle()
	return lipgloss.JoinVertical(lipgloss.Top, m.renderRow(style, 0, 4), m.renderRow(style, 4, 8))
}

func (m model) renderRow(style lipgloss.Style, start, end int) string {
	r := ""
	for i := start; i < end; i++ {
		if m.focus == i {
			style = style.BorderForeground(ColorBlue)
		} else {
			style = style.BorderForeground(ColorWhite)
		}
		r = lipgloss.JoinHorizontal(lipgloss.Top, r, style.Render(m.boxes[i]))
	}
	return r
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
