package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rwirdemann/7dayz"
	"github.com/rwirdemann/7dayz/file"
	"log/slog"
	"os"
	"strings"
	"time"
)

const (
	none = iota
	edit = iota
	add
)

var ColorBlue = lipgloss.Color("12")
var ColorWhite = lipgloss.Color("255")
var ColorGrey = lipgloss.Color("240")

type model struct {
	boxModel   _dayz.BoxModel
	fullWidth  int
	fullHeight int
	textinput  textinput.Model
	mode       int
}

func initialModel() model {
	return model{
		boxModel:  _dayz.NewBoxModel(file.TaskRepository{}),
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
						m.boxModel.Boxes[m.boxModel.Focus].InsertItem(0, _dayz.Task{Name: value, Day: m.boxModel.Focus})
					}

					if m.mode == edit {
						m.boxModel.Update(value, m.boxModel.Focus)
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
	m.boxModel.Boxes[m.boxModel.Focus], cmd = m.boxModel.Boxes[m.boxModel.Focus].Update(msg)
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
			m.boxModel.Save()
			return m, tea.Quit

		// Move focus to next box
		case "tab":
			m.boxModel = m.boxModel.NextDay()

		// Move focus to prev box
		case "shift+tab":
			m.boxModel = m.boxModel.PreviousDay()

		// Move item to next box
		case "m":
			m.boxModel.MoveItem(m.boxModel.Focus + 1)

		// Move item to next box
		case "b":
			if item := m.boxModel.Boxes[m.boxModel.Focus].SelectedItem(); item != nil && m.boxModel.Focus > 0 {
				t := item.(_dayz.Task)
				m.boxModel.Boxes[m.boxModel.Focus].RemoveItem(m.boxModel.Boxes[m.boxModel.Focus].Index())
				t.Day = m.boxModel.Focus - 1
				m.boxModel.Boxes[m.boxModel.Focus-1].InsertItem(0, t)
			}

		// Move task to today
		case "ctrl+t":
			slog.Info("ctrl+t")
			today := time.Now().Weekday()
			m.boxModel.MoveItem(int(today))

		// Edit selected item in footer
		case "enter":
			selected := m.boxModel.Boxes[m.boxModel.Focus].SelectedItem()
			m.textinput.SetValue(selected.(_dayz.Task).Name)
			m.textinput.Focus()
			m.mode = edit

		// Add new item
		case "n":
			m.textinput.Focus()
			m.mode = add

		// Cross item off or on
		case " ":
			selected := m.boxModel.Boxes[m.boxModel.Focus].SelectedItem().(_dayz.Task)
			selected.Done = !selected.Done
			m.boxModel.Boxes[m.boxModel.Focus].RemoveItem(m.boxModel.Boxes[m.boxModel.Focus].Index())
			m.boxModel.Boxes[m.boxModel.Focus].InsertItem(m.boxModel.Boxes[m.boxModel.Focus].Index(), selected)

		case "backspace":
			m.boxModel.Boxes[m.boxModel.Focus].RemoveItem(m.boxModel.Boxes[m.boxModel.Focus].Index())
		case "t":
			today := time.Now().Weekday()
			m.boxModel.Focus = int(today)
			_dayz.ActiveBox = m.boxModel.Boxes[m.boxModel.Focus].Title
		}

	}
	return m, tea.Batch(cmds...)
}

func (m model) size() (int, int, int, int) {
	var w, h, wDelta, hDelta int
	if (m.fullWidth - 8%4) == 0 {
		w = (m.fullWidth - 8) / 4
	} else {
		wDelta = (m.fullWidth - 8) % 4
		w = (m.fullWidth - 8) / 4
	}

	const extraHeight = 8
	if (m.fullHeight - extraHeight%2) == 0 {
		h = (m.fullHeight - extraHeight) / 2
	} else {
		hDelta = (m.fullHeight - extraHeight) % 2
		h = (m.fullHeight - extraHeight) / 2
	}

	h = m.fullHeight/2 - 2

	return w, h, wDelta, hDelta
}

func (m model) View() string {
	w, h, wDelta, hDelta := m.size()

	// Skip first rendering when we don't know the terminal size yet
	if w <= 0 || h <= 0 {
		return ""
	}
	style := lipgloss.NewStyle().Width(w).Height(h)
	row1 := m.renderRow(0, 4, style, wDelta, hDelta)

	// Reduce height of second row to fit textinput underneath
	style = style.Height(h - 2)
	row2 := m.renderRow(4, 8, style, wDelta, hDelta)

	var footer string
	switch m.mode {
	case add, edit:
		footer = " " + m.textinput.View()
	case none:
		footer = m.helpView()
	}
	return lipgloss.JoinVertical(lipgloss.Top, row1, row2, footer)
}

func (m model) helpView() string {
	helpStyle := lipgloss.NewStyle().Foreground(ColorGrey)
	return helpStyle.Render(" tab: next day • shift+tab: prev day • enter: edit task • n: new task • t: today")
}

func generateBorder(title string, width int) lipgloss.Border {
	if width < 0 {
		return lipgloss.RoundedBorder()
	}
	border := lipgloss.RoundedBorder()
	border.Top = border.Top + border.MiddleRight + " " + title + " " + border.MiddleLeft + strings.Repeat(border.Top, width)
	return border
}

func (m model) renderRow(start, end int, style lipgloss.Style, wDelta int, hDelta int) string {
	r := ""
	for i := start; i < end; i++ {
		if m.boxModel.Focus == i {
			style = style.BorderForeground(ColorBlue)
		} else {
			style = style.BorderForeground(ColorWhite)
		}

		if i == 3 || i == 7 {
			style = style.Width(style.GetWidth() + wDelta)
		}

		if i == 4 {
			style = style.Height(style.GetHeight() + hDelta)
		}

		// Adapt the box list model to current box dimensions
		m.boxModel.Boxes[i].SetHeight(style.GetHeight())
		m.boxModel.Boxes[i].SetWidth(style.GetWidth())

		style = style.Border(generateBorder(m.boxModel.Boxes[i].Title, style.GetWidth()))
		r = lipgloss.JoinHorizontal(lipgloss.Top, r, style.Render(m.boxModel.Boxes[i].View()))
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
