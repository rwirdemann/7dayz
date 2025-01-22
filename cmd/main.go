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
				weekplanner.Item("Konzept erstellen"),
				weekplanner.Item("Miete anbieten"),
				weekplanner.Item("LTE App reviewen"),
				weekplanner.Item("Health Monitor reviewen"),
			}
		}
		b := weekplanner.NewBox(t, items, 0, 0)
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

func generateBorder(title string, width int) lipgloss.Border {
	if width < 0 {
		return lipgloss.RoundedBorder()
	}
	border := lipgloss.RoundedBorder()
	border.Top = border.Top + border.MiddleRight + " " + title + " " + border.MiddleLeft + strings.Repeat(border.Top, width)
	return border
}

func (m model) View() string {
	style := lipgloss.NewStyle().Width(m.fullWidth/4 - 2).Height(m.fullHeight/2 - 2)
	row1 := m.renderRow(0, 4, style)
	row2 := m.renderRow(4, 8, style)
	return lipgloss.JoinVertical(lipgloss.Top, row1, row2)
}

func (m model) renderRow(start, end int, style lipgloss.Style) string {
	r := ""
	for i := start; i < end; i++ {
		if m.focus == i {
			style = style.BorderForeground(ColorBlue)
		} else {
			style = style.BorderForeground(ColorWhite)
		}

		// Adapt the box list model to current box dimensions
		m.boxes[i].SetHeight(style.GetHeight())
		m.boxes[i].SetWidth(style.GetWidth())

		style = style.Border(generateBorder(m.boxes[i].Title, style.GetWidth()))
		r = lipgloss.JoinHorizontal(lipgloss.Top, r, style.Render(m.boxes[i].View()))
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
