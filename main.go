package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"os"

	"github.com/charmbracelet/bubbletea"
)

var ColorBlue = lipgloss.Color("12")
var ColorWhite = lipgloss.Color("255")

type model struct {
	boxes []string
	focus int
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

	// Is it a key press?
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

func boxSize() (int, int) {
	w, h, _ := term.GetSize(os.Stdout.Fd())
	return (w / 4) - 2, (h / 2) - 2
}

func (m model) View() string {
	w, h := boxSize()
	var style = lipgloss.NewStyle().
		Width(w).
		Height(h).
		BorderStyle(lipgloss.NormalBorder())
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
