package _dayz

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"log/slog"
	"sort"
	"strings"
)

// ActiveTab represents the currently active box, identified by its title. The selection pointer is only rendered for
// the currently active box.
var ActiveTab = "Inbox"

type Tab struct {
	list.Model
}

func NewTab(title string) Tab {
	l := list.New(nil, itemDelegate{boxTitle: title}, 0, 0)
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "shift+up":
			selected := b.SelectedItem()
			if selected != nil {
				t := selected.(Task)
				if b.Index() > 0 {
					b.RemoveItem(b.Index())
					b.InsertItem(b.Index()-1, t)
					b.Select(b.Index() - 1)
				}
			}
		case "shift+down":
			selected := b.SelectedItem()
			if selected != nil {
				t := selected.(Task)
				if b.Index() < len(b.Items())-1 {
					b.RemoveItem(b.Index())
					b.InsertItem(b.Index()+1, t)
					b.Select(b.Index() + 1)
				}
			}
		}
	}
	return b, cmd
}

type TabModel struct {
	Tabs       []Tab
	repository TaskRepository
	Focus      int
}

func NewTabModel(repository TaskRepository) TabModel {
	m := TabModel{repository: repository, Focus: 0}
	titles := []string{"Inbox", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, t := range titles {
		m.Tabs = append(m.Tabs, NewTab(t))
	}
	return m
}

func (m TabModel) Load() {
	var tasksByDay = make(map[int][]list.Item)
	tasks := m.repository.Load()
	for _, task := range tasks {
		tasksByDay[task.Day] = append(tasksByDay[task.Day], task)
	}

	// Sort tasks by their Pos field
	for _, items := range tasksByDay {
		sort.Slice(items, func(i, j int) bool {
			return items[i].(Task).Pos < items[j].(Task).Pos
		})
	}

	for day := range m.Tabs {
		for i, item := range tasksByDay[day] {
			m.Tabs[day].InsertItem(i, item)
		}
	}
}

func (m TabModel) Save() {
	var tasks []Task
	for _, box := range m.Tabs {
		for i, item := range box.Items() {
			t := item.(Task)
			t.Pos = i
			tasks = append(tasks, t)
		}
	}
	m.repository.Save(tasks)
}

func (m TabModel) Add(s string) {
	m.Tabs[m.Focus].InsertItem(0, Task{Name: s, Day: m.Focus})
}

func (m TabModel) Update(s string) {
	task := m.Tabs[m.Focus].SelectedItem().(Task)
	task.Name = s
	m.Tabs[m.Focus].RemoveItem(m.Tabs[m.Focus].Index())
	m.Tabs[m.Focus].InsertItem(m.Tabs[m.Focus].Index(), task)
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

	width := m.Width() - 6

	// Some extra space for the done mark
	if task.Done {
		width -= 2
	}

	name := task.Name
	if len(name) > width {
		name = name[:width] + "..."
	}

	if task.Done {
		name = fmt.Sprintf("X %s", name)
	}
	_, err := fmt.Fprint(w, fn(name))
	if err != nil {
		slog.Error(err.Error())
	}
}
