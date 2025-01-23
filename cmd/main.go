package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rwirdemann/weekplanner"
	"github.com/rwirdemann/weekplanner/file"
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
	boxModel   weekplanner.BoxModel
	fullWidth  int
	fullHeight int
	textinput  textinput.Model
	mode       int
}

func initialModel() model {
	return model{
		boxModel:  weekplanner.NewBoxModel(file.TaskRepository{}),
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
						m.boxModel.Boxes[m.boxModel.Focus].InsertItem(0, weekplanner.Task{Name: value, Day: m.boxModel.Focus})
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

			// Tell box logic which box is active to enable proper rendering of selected item
			weekplanner.ActiveBox = m.boxModel.Boxes[m.boxModel.Focus].Title

		// Move focus to prev box
		case "shift+tab":
			if m.boxModel.Focus == 0 {
				m.boxModel.Focus = 7
			} else {
				m.boxModel.Focus--
			}
			// Tell box logic which box is active to enable proper rendering of selected item
			weekplanner.ActiveBox = m.boxModel.Boxes[m.boxModel.Focus].Title

		// Move item to next box
		case "m":
			if item := m.boxModel.Boxes[m.boxModel.Focus].SelectedItem(); item != nil && m.boxModel.Focus < 7 {
				t := item.(weekplanner.Task)
				m.boxModel.Boxes[m.boxModel.Focus].RemoveItem(m.boxModel.Boxes[m.boxModel.Focus].Index())
				t.Day = m.boxModel.Focus + 1
				m.boxModel.Boxes[m.boxModel.Focus+1].InsertItem(0, t)
			}

		// Move item to next box
		case "b":
			if item := m.boxModel.Boxes[m.boxModel.Focus].SelectedItem(); item != nil && m.boxModel.Focus > 0 {
				t := item.(weekplanner.Task)
				m.boxModel.Boxes[m.boxModel.Focus].RemoveItem(m.boxModel.Boxes[m.boxModel.Focus].Index())
				t.Day = m.boxModel.Focus - 1
				m.boxModel.Boxes[m.boxModel.Focus-1].InsertItem(0, t)
			}

		// Edit selected item in footer
		case "enter":
			selected := m.boxModel.Boxes[m.boxModel.Focus].SelectedItem()
			m.textinput.SetValue(selected.(weekplanner.Task).Name)
			m.textinput.Focus()
			m.mode = edit

		// Add new item
		case "n":
			m.textinput.Focus()
			m.mode = add

		// Cross item off or on
		case " ":
			selected := m.boxModel.Boxes[m.boxModel.Focus].SelectedItem().(weekplanner.Task)
			selected.Done = !selected.Done
			m.boxModel.Boxes[m.boxModel.Focus].RemoveItem(m.boxModel.Boxes[m.boxModel.Focus].Index())
			m.boxModel.Boxes[m.boxModel.Focus].InsertItem(m.boxModel.Boxes[m.boxModel.Focus].Index(), selected)

		case "backspace":
			m.boxModel.Boxes[m.boxModel.Focus].RemoveItem(m.boxModel.Boxes[m.boxModel.Focus].Index())
		case "t":
			today := time.Now().Weekday()
			m.boxModel.Focus = int(today)
			weekplanner.ActiveBox = m.boxModel.Boxes[m.boxModel.Focus].Title
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

func (m model) renderRow(start, end int, style lipgloss.Style) string {
	r := ""
	for i := start; i < end; i++ {
		if m.boxModel.Focus == i {
			style = style.BorderForeground(ColorBlue)
		} else {
			style = style.BorderForeground(ColorWhite)
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
