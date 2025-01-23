package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
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
	boxes      []weekplanner.Box
	focus      int
	fullWidth  int
	fullHeight int
	textinput  textinput.Model
	mode       int
}

var taskRepository = file.TaskRepository{}

func initialModel() model {
	titles := []string{"Inbox", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	var boxes []weekplanner.Box
	for _, t := range titles {
		var items []list.Item
		b := weekplanner.NewBox(t, items)
		boxes = append(boxes, b)
	}

	tasks := taskRepository.Load()
	for _, task := range tasks {
		boxes[task.Day].InsertItem(0, task)
	}

	return model{
		focus:     0,
		boxes:     boxes,
		textinput: textinput.New(),
		mode:      none,
	}
}

func (m model) save() {
	var tasks []weekplanner.Task
	for _, box := range m.boxes {
		for _, item := range box.Items() {
			tasks = append(tasks, item.(weekplanner.Task))
		}
	}
	taskRepository.Save(tasks)
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
						m.boxes[m.focus].InsertItem(0, weekplanner.Task{Name: value, Day: m.focus})
					}

					if m.mode == edit {
						m.boxes[m.focus].RemoveItem(m.boxes[m.focus].Index())
						m.boxes[m.focus].InsertItem(m.boxes[m.focus].Index(), weekplanner.Task{Name: value})
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
			m.save()
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
			if item := m.boxes[m.focus].SelectedItem(); item != nil && m.focus < 7 {
				t := item.(weekplanner.Task)
				m.boxes[m.focus].RemoveItem(m.boxes[m.focus].Index())
				t.Day = m.focus + 1
				m.boxes[m.focus+1].InsertItem(0, t)
			}

		// Move item to next box
		case "b":
			if item := m.boxes[m.focus].SelectedItem(); item != nil && m.focus > 0 {
				t := item.(weekplanner.Task)
				m.boxes[m.focus].RemoveItem(m.boxes[m.focus].Index())
				t.Day = m.focus - 1
				m.boxes[m.focus-1].InsertItem(0, t)
			}

		// Edit selected item in footer
		case "enter":
			selected := m.boxes[m.focus].SelectedItem()
			m.textinput.SetValue(selected.(weekplanner.Task).Name)
			m.textinput.Focus()
			m.mode = edit

		// Add new item
		case "n":
			m.textinput.Focus()
			m.mode = add

		// Cross item off or on
		case "x":
			selected := m.boxes[m.focus].SelectedItem().(weekplanner.Task)
			selected.Done = !selected.Done
			m.boxes[m.focus].RemoveItem(m.boxes[m.focus].Index())
			m.boxes[m.focus].InsertItem(m.boxes[m.focus].Index(), selected)

		case "backspace":
			m.boxes[m.focus].RemoveItem(m.boxes[m.focus].Index())
		case "t":
			today := time.Now().Weekday()
			m.focus = int(today)
			weekplanner.ActiveBox = m.boxes[m.focus].Title
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
