package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rwirdemann/weekplanner"
	"os"
	"strings"
)

const (
	none = iota
	edit = iota
	add
)

var ColorBlue = lipgloss.Color("12")
var ColorWhite = lipgloss.Color("255")

type model struct {
	boxes      []weekplanner.Box
	focus      int
	fullWidth  int
	fullHeight int
	textinput  textinput.Model
	mode       int
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
		b := weekplanner.NewBox(t, items)
		boxes = append(boxes, b)
	}

	return model{
		focus:     0,
		boxes:     boxes,
		textinput: textinput.New(),
		mode:      none,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle text input in the footer
	if m.textinput.Focused() {
		m.textinput, cmd = m.textinput.Update(msg)
		cmds = append(cmds, cmd)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				value := m.textinput.Value()
				if len(strings.TrimSpace(value)) > 0 {
					if m.mode == add {
						m.boxes[m.focus].InsertItem(0, weekplanner.Item(value))
					}

					if m.mode == edit {
						m.boxes[m.focus].RemoveItem(m.boxes[m.focus].Index())
						m.boxes[m.focus].InsertItem(m.boxes[m.focus].Index(), weekplanner.Item(value))
					}
				}

				m.textinput.Reset()
				m.textinput.Blur()
				m.mode = none
			case "esc":
				m.textinput.Reset()
				m.textinput.Blur()
				m.mode = none
			}
		}
		return m, tea.Batch(cmds...)
	}

	// Update active box
	m.boxes[m.focus], cmd = m.boxes[m.focus].Update(msg)
	cmds = append(cmds, cmd)

	// Handle global key strokes
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.fullHeight = msg.Height
		m.fullWidth = msg.Width

	case tea.KeyMsg:
		switch msg.String() {

		// Exit
		case "ctrl+c", "q":
			return m, tea.Quit

		// Move focus to next box
		case "tab":
			if m.focus == 7 {
				m.focus = 0
			} else {
				m.focus++
			}

			// Tell box logic which box is active to enable proper rendering of selected item
			weekplanner.ActiveBox = m.boxes[m.focus].Title

		// Move focus to prev box
		case "shift+tab":
			if m.focus == 0 {
				m.focus = 7
			} else {
				m.focus--
			}
			// Tell box logic which box is active to enable proper rendering of selected item
			weekplanner.ActiveBox = m.boxes[m.focus].Title

		// Move item to next box
		case "m":
			if m.focus < 7 {
				selected := m.boxes[m.focus].SelectedItem()
				m.boxes[m.focus].RemoveItem(m.boxes[m.focus].Index())
				m.boxes[m.focus+1].InsertItem(0, selected)
			}

		// Move item to next box
		case "b":
			if m.focus > 0 {
				selected := m.boxes[m.focus].SelectedItem()
				m.boxes[m.focus].RemoveItem(m.boxes[m.focus].Index())
				m.boxes[m.focus-1].InsertItem(0, selected)
			}

		// Edit selected item in footer
		case "enter":
			selected := m.boxes[m.focus].SelectedItem()
			m.textinput.SetValue(string(selected.(weekplanner.Item)))
			m.textinput.Focus()
			m.mode = edit

		// Add new item
		case "n":
			m.textinput.Focus()
			m.mode = add
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	w, h := m.fullWidth/4-2, m.fullHeight/2-2

	// Skip first rendering when we don't know the terminal size yet
	if w <= 0 || h <= 0 {
		return ""
	}
	style := lipgloss.NewStyle().Width(w).Height(h)
	row1 := m.renderRow(0, 4, style)

	// Reduce height of second row to fit textinput underneath
	style = style.Height(h - 1)
	row2 := m.renderRow(4, 8, style)
	row3 := " " + m.textinput.View()
	return lipgloss.JoinVertical(lipgloss.Top, row1, row2, row3)
}

func generateBorder(title string, width int) lipgloss.Border {
	if width < 0 {
		return lipgloss.RoundedBorder()
	}
	border := lipgloss.RoundedBorder()
	border.Top = border.Top + border.MiddleRight + " " + title + " " + border.MiddleLeft + strings.Repeat(border.Top, width)
	return border
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
