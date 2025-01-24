package _dayz

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"log/slog"
	"strings"
)

// ActiveTab represents the currently active box, identified by its title. The selection pointer is only rendered for
// the currently active box.
var ActiveTab = "Inbox"

type Tab struct {
	list.Model
}

func NewTab(title string, items []list.Item) Tab {
	l := list.New(items, itemDelegate{boxTitle: title}, 0, 0)
	l.Title = title
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
	return Tab{Model: l}
}

func (b Tab) Update(msg tea.Msg) (Tab, tea.Cmd) {
	var cmd tea.Cmd
	b.Model, cmd = b.Model.Update(msg)
	return b, cmd
}

type TabModel struct {
	Tabs       []Tab
	repository TaskRepository
	Focus      int
}

func NewTabModel(repository TaskRepository) TabModel {
	var tasksByDay = make(map[int][]list.Item)
	tasks := repository.Load()
	for _, task := range tasks {
		tasksByDay[task.Day] = append(tasksByDay[task.Day], task)
	}

	m := TabModel{repository: repository, Focus: 0}
	titles := []string{"Inbox", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for i, t := range titles {
		b := NewTab(t, tasksByDay[i])
		m.Tabs = append(m.Tabs, b)
	}
	return m
}

func (m TabModel) Save() {
	var tasks []Task
	for _, box := range m.Tabs {
		for _, item := range box.Items() {
			tasks = append(tasks, item.(Task))
		}
	}
	m.repository.Save(tasks)
}

func (m TabModel) Update(s string, box int) {
	task := m.Tabs[box].SelectedItem().(Task)
	task.Name = s
	m.Tabs[box].RemoveItem(m.Tabs[box].Index())
	m.Tabs[box].InsertItem(m.Tabs[box].Index(), task)
}

func (m TabModel) NextTab() TabModel {
	if m.Focus == 7 {
		m.Focus = 0
	} else {
		m.Focus++
	}
	ActiveTab = m.Tabs[m.Focus].Title
	return m
}

func (m TabModel) PreviousTab() TabModel {
	if m.Focus == 0 {
		m.Focus = 7
	} else {
		m.Focus--
	}
	ActiveTab = m.Tabs[m.Focus].Title
	return m
}
func (m TabModel) MoveItem(to int) {
	if item := m.Tabs[m.Focus].SelectedItem(); item != nil && to < 7 && to >= 0 {
		t := item.(Task)
		m.Tabs[m.Focus].RemoveItem(m.Tabs[m.Focus].Index())
		t.Day = to
		m.Tabs[to].InsertItem(0, t)
	}
}

func (m TabModel) SelectTab(i int) TabModel {
	m.Focus = i
	ActiveTab = m.Tabs[i].Title
	return m
}

type itemDelegate struct {
	boxTitle string
}

func (d itemDelegate) Height() int {
	return 1
}

func (d itemDelegate) Spacing() int {
	return 0
}

func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	task, ok := listItem.(Task)
	if !ok {
		return
	}

	fn := lipgloss.NewStyle().Strikethrough(task.Done).PaddingLeft(2).Render

	// Render selection pointer only for the currently active box.
	if ActiveTab == d.boxTitle && index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().Strikethrough(task.Done).PaddingLeft(0).Foreground(lipgloss.Color("170")).Render("> " + strings.Join(s, " "))
		}
	}

	name := task.Name
	if task.Done {
		name = fmt.Sprintf("X %s", name)
	}
	_, err := fmt.Fprint(w, fn(name))
	if err != nil {
		slog.Error(err.Error())
	}
}
