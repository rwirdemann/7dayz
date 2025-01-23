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

type BoxModel struct {
	Boxes      []Box
	repository TaskRepository
	Focus      int
}

func NewBoxModel(repository TaskRepository) BoxModel {
	var tasksByDay = make(map[int][]list.Item)
	tasks := repository.Load()
	for _, task := range tasks {
		tasksByDay[task.Day] = append(tasksByDay[task.Day], task)
	}

	m := BoxModel{repository: repository, Focus: 0}
	titles := []string{"Inbox", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for i, t := range titles {
		b := NewBox(t, tasksByDay[i])
		m.Boxes = append(m.Boxes, b)
	}
	return m
}

func (m BoxModel) Save() {
	var tasks []Task
	for _, box := range m.Boxes {
		for _, item := range box.Items() {
			tasks = append(tasks, item.(Task))
		}
	}
	m.repository.Save(tasks)
}

func (m BoxModel) Update(s string, box int) {
	task := m.Boxes[box].SelectedItem().(Task)
	task.Name = s
	m.Boxes[box].RemoveItem(m.Boxes[box].Index())
	m.Boxes[box].InsertItem(m.Boxes[box].Index(), task)
}

func (m BoxModel) NextDay() BoxModel {
	if m.Focus == 7 {
		m.Focus = 0
	} else {
		m.Focus++
	}
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
	if ActiveBox == d.boxTitle && index == m.Index() {
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
