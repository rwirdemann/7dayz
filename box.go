package weekplanner

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"log/slog"
	"strings"
)

// ActiveBox represents the currently active box, identified by its title. The selection pointer is only rendered for
// the currently active box.
var ActiveBox = "Inbox"

type itemDelegate struct {
	boxTitle string
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	task, ok := listItem.(Task)
	if !ok {
		return
	}

	fn := lipgloss.NewStyle().Strikethrough(task.Done).PaddingLeft(2).Render

	// Render selection pointer only for the currently active box.
	if ActiveBox == d.boxTitle && index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().Strikethrough(task.Done).PaddingLeft(0).Foreground(lipgloss.Color("170")).Render("> " + strings.Join(s, " "))
		}
	}

	_, err := fmt.Fprint(w, fn(task.Name))
	if err != nil {
		slog.Error(err.Error())
	}
}

type Box struct {
	list.Model
}

func NewBox(title string, items []list.Item) Box {
	l := list.New(items, itemDelegate{boxTitle: title}, 0, 0)
	l.Title = title
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
	return Box{Model: l}
}

func (b Box) Update(msg tea.Msg) (Box, tea.Cmd) {
	var cmd tea.Cmd
	b.Model, cmd = b.Model.Update(msg)
	return b, cmd
}
