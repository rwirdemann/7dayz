package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

var ColorBlue = lipgloss.Color("12")
var ColorWhite = lipgloss.Color("255")

type item string

func (i item) FilterValue() string {
	return ""
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	fn := lipgloss.NewStyle().PaddingLeft(2).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170")).Render("> " + strings.Join(s, " "))
		}
	}

	_, err := fmt.Fprint(w, fn(fmt.Sprintf("%s", i)))
	if err != nil {
		slog.Error(err.Error())
	}
}

type model struct {
	boxes      []string
	inbox      list.Model
	focus      int
	fullWidth  int
	fullHeight int
}

func initialModel() model {
	var items = []list.Item{
		item("XMas-Karten unterschreiben"),
		item("Register Manager weiter bauen"),
		item("Stunden aufschreiben"),
	}

	l := list.New(items, itemDelegate{}, 20, 14)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2)

	return model{
		focus: 0,
		inbox: l,
		boxes: []string{"Inbox", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
	}
}

func (m model) Init() tea.Cmd {
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
		if i == 0 {
			m.inbox.SetHeight(style.GetHeight() - 6)
			s := lipgloss.JoinVertical(lipgloss.Top, m.boxes[i], m.inbox.View())
			r = lipgloss.JoinHorizontal(lipgloss.Top, r, style.Render(s))
		} else {
			r = lipgloss.JoinHorizontal(lipgloss.Top, r, style.Render(m.boxes[i]))
		}

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
