package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rwirdemann/7dayz"
	"github.com/rwirdemann/7dayz/file"
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
var ColorGrey = lipgloss.Color("248")
var ColorGreen = lipgloss.Color("28")

var helpBlockStyle = lipgloss.NewStyle().Foreground(ColorGrey)
var helpHeaderStyle = lipgloss.NewStyle().Foreground(ColorGreen)

type model struct {
	boxModel   _dayz.TabModel
	fullWidth  int
	fullHeight int
	textinput  textinput.Model
	mode       int
	showHelp   bool
}

func initialModel() model {
	return model{
		boxModel:  _dayz.NewTabModel(file.TaskRepository{}),
		textinput: textinput.New(),
		mode:      none,
		showHelp:  true,
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
						m.boxModel.Tabs[m.boxModel.Focus].InsertItem(0, _dayz.Task{Name: value, Day: m.boxModel.Focus})
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
	m.boxModel.Tabs[m.boxModel.Focus], cmd = m.boxModel.Tabs[m.boxModel.Focus].Update(msg)
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
			m.boxModel = m.boxModel.NextTab()

		// Move focus to prev box
		case "shift+tab":
			m.boxModel = m.boxModel.PreviousTab()

		case "i":
			m.boxModel = m.boxModel.SelectTab(0)

		case "I":
			m.boxModel.MoveItem(0)

		// Move item to next box
		case "shift+right":
			m.boxModel.MoveItem(m.boxModel.Focus + 1)

			// Move item to prev box
		case "shift+left":
			m.boxModel.MoveItem(m.boxModel.Focus - 1)

		// Move task to today
		case "T":
			today := time.Now().Weekday()
			m.boxModel.MoveItem(int(today))

		// Edit selected item in footer
		case "enter":
			selected := m.boxModel.Tabs[m.boxModel.Focus].SelectedItem()
			m.textinput.SetValue(selected.(_dayz.Task).Name)
			m.textinput.Focus()
			m.mode = edit
			m.showHelp = false

		// Add new item
		case "n":
			m.textinput.Focus()
			m.mode = add
			m.showHelp = false

		// Cross item off or on
		case " ":
			selected := m.boxModel.Tabs[m.boxModel.Focus].SelectedItem().(_dayz.Task)
			selected.Done = !selected.Done
			m.boxModel.Tabs[m.boxModel.Focus].RemoveItem(m.boxModel.Tabs[m.boxModel.Focus].Index())
			m.boxModel.Tabs[m.boxModel.Focus].InsertItem(m.boxModel.Tabs[m.boxModel.Focus].Index(), selected)

		case "backspace":
			m.boxModel.Tabs[m.boxModel.Focus].RemoveItem(m.boxModel.Tabs[m.boxModel.Focus].Index())
		case "t":
			today := time.Now().Weekday()
			m.boxModel.Focus = int(today)
			_dayz.ActiveTab = m.boxModel.Tabs[m.boxModel.Focus].Title
		case "?":
			m.showHelp = !m.showHelp
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

	// Reduce height of second row to fit textinput or help underneath
	gap := 0
	if m.showHelp {
		gap = 6
	}

	if m.mode != none {
		gap = 2
	}
	style = style.Height(h - gap)
	row2 := m.renderRow(4, 8, style, wDelta, hDelta)

	if m.mode != none {
		footer := " " + m.textinput.View()
		return lipgloss.JoinVertical(lipgloss.Top, row1, row2, footer)
	}

	if m.showHelp {
		footer := " " + m.helpView()
		return lipgloss.JoinVertical(lipgloss.Top, row1, row2, footer)
	}

	return lipgloss.JoinVertical(lipgloss.Top, row1, row2)
}

type key struct {
	k    string
	desc string
}

func (m model) helpView() string {
	general := []key{
		{"tab", "focus next day"},
		{"shift+tab", "focus prev day"},
		{"t", "focus today"},
		{"i", "focus inbox"},

		{"?", "toggle help"},
	}

	management := []key{
		{"backspace", "delete task"},
		{"space", "complete task"},
		{"enter", "edit task"},
		{"n", "new task"},
	}

	movement := []key{
		{"shift+right", "move task right"},
		{"shift+left", "move task left"},
		{"shift+t", "move task to today"},
		{"shift+i", "move task to inbox"},
	}

	sorting := []key{
		{"shift+up", "move task up"},
		{"shift+down", "move task down"},
	}

	return lipgloss.JoinHorizontal(0,
		m.helpBlock("General", 11, 16, general), "  ",
		m.helpBlock("Task Editing", 11, 15, management), "  ",
		m.helpBlock("Task Movement", 13, 20, movement), "  ",
		m.helpBlock("Sorting", 12, 20, sorting))
}

func (m model) helpBlock(title string, padding int, space int, keys []key) string {
	block := helpHeaderStyle.Render(title)
	for _, k := range keys {
		block = lipgloss.JoinVertical(0, block, helpBlockStyle.Render(fmt.Sprintf("%*s%*s", padding, k.k+" ", space, k.desc)))
	}

	return block
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
		m.boxModel.Tabs[i].SetHeight(style.GetHeight())
		m.boxModel.Tabs[i].SetWidth(style.GetWidth())

		style = style.Border(generateBorder(m.boxModel.Tabs[i].Title, style.GetWidth()))
		r = lipgloss.JoinHorizontal(lipgloss.Top, r, style.Render(m.boxModel.Tabs[i].View()))
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
