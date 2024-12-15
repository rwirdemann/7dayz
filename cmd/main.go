package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rwirdemann/weekplanner"
	"os"
	"strings"
)

var ColorBlue = lipgloss.Color("12")
var ColorWhite = lipgloss.Color("255")

type model struct {
	boxes      []weekplanner.Box
	focus      int
	fullWidth  int
	fullHeight int
}

func initialModel() model {
	titles := []string{"Inbox", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	var boxes []weekplanner.Box
	for _, t := range titles {
		var items []list.Item
		if t == "Inbox" {
			items = []list.Item{
				weekplanner.Item("XMas-Karten unterschreiben"),
				weekplanner.Item("Register Manager weiter bauen"),
				weekplanner.Item("Stunden aufschreiben"),
			}
		}
		b := weekplanner.NewBox(t, items, 20, 14)
		boxes = append(boxes, b)
	}

	return model{
		focus: 0,
		boxes: boxes,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.boxes[m.focus], cmd = m.boxes[m.focus].Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.fullHeight = msg.Height
		m.fullWidth = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
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

	return m, tea.Batch(cmds...)
}

func (m model) boxSize() (int, int) {
	width := 0
	if m.fullWidth%4 != 0 {
		width = footerWidth(m.fullWidth + m.fullWidth%4 + 2)
	} else {
		width = footerWidth(m.fullWidth)
	}

	height := 0
	if m.fullHeight%2 != 0 {
		height = footerHeight(m.fullHeight + m.fullHeight%2 + 2)
	} else {
		height = footerHeight(m.fullHeight)
	}

	return width, height
}

func footerWidth(fullWidth int) int {
	return fullWidth/4 - 2
}

func footerHeight(fullHeight int) int {
	return fullHeight/2 - 2
}

func generateBorder(title string, width int) lipgloss.Border {
	if width < 0 {
		return lipgloss.RoundedBorder()
	}
	border := lipgloss.RoundedBorder()
	border.Top = border.Top + border.MiddleRight + " " + title + " " + border.MiddleLeft + strings.Repeat(border.Top, width)
	return border
}

func (m model) boxStyle(title string) lipgloss.Style {
	w, h := m.boxSize()
	return lipgloss.NewStyle().
		Border(generateBorder(title, w)).
		Width(w).
		Height(h)
}

func (m model) View() string {
	boxesPerRow := len(m.boxes) / 2
	return lipgloss.JoinVertical(lipgloss.Top,
		m.renderRow(0, boxesPerRow),
		m.renderRow(boxesPerRow, len(m.boxes)))
}

func (m model) renderRow(start, end int) string {
	r := ""
	for i := start; i < end; i++ {
		var style = m.boxStyle(m.boxes[i].Title)
		if m.focus == i {
			style = style.BorderForeground(ColorBlue)
		} else {
			style = style.BorderForeground(ColorWhite)
		}
		r = lipgloss.JoinHorizontal(lipgloss.Right, r, style.Render(m.boxes[i].View()))
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
