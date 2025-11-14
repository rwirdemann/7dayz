package perpetask

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"
)

var ActiveTab = 0

type Tab struct {
	list.Model
	number int
	Week   int
	Date   time.Time
}

func (b Tab) TitleString() string {
	if b.number == 0 {
		return fmt.Sprintf("Inbox (Week %d)", b.Week)
	}
	return fmt.Sprintf("%s (%s)", b.Title, b.Date.Format("02.01.2006"))
}

func NewTab(title string, number int, week int) Tab {
	l := list.New(nil, itemDelegate{panel: number}, 0, 0)
	l.Title = title
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
	return Tab{Model: l, number: number, Week: week}
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
	Week       int
}

func NewTabModel(repository TaskRepository, week int) TabModel {
	m := TabModel{repository: repository, Focus: 0, Week: week}
	titles := []string{"Inbox", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	weekday := getMondayOfWeek(week)
	for i, t := range titles {
		tab := NewTab(t, i, week)
		if i > 0 {
			tab.Date = weekday
			weekday = weekday.AddDate(0, 0, 1)
		}
		m.Tabs = append(m.Tabs, tab)
	}
	return m
}

// Get the date of Monday for the given week
func getMondayOfWeek(week int) time.Time {
	now := time.Now()

	// Get January 1 of the current year
	startOfYear := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.Local)

	// Calculate the first Monday on or after January 1
	firstMonday := startOfYear
	for firstMonday.Weekday() != time.Monday {
		firstMonday = firstMonday.AddDate(0, 0, 1)
	}

	// Add the offset for the given week (weeks start from 0)
	daysToAdd := (week - 2) * 7
	mondayOfWeek := firstMonday.AddDate(0, 0, daysToAdd)

	return mondayOfWeek
}

func (m TabModel) NextWeek() TabModel {
	if m.Week == 52 {
		return m
	}
	return m.updateDates(1)
}

func (m TabModel) PrevWeek() TabModel {
	if m.Week == 1 {
		return m
	}
	return m.updateDates(-1)
}

func (m TabModel) updateDates(delta int) TabModel {
	m.Week += delta
	weekday := getMondayOfWeek(m.Week)
	for i := range m.Tabs {
		m.Tabs[i].Week = m.Week
		if i == 0 {
			continue
		}
		m.Tabs[i].Date = weekday
		weekday = weekday.AddDate(0, 0, 1)
	}
	return m
}

func (m TabModel) Load(week int) {
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
	// Find index of last task in list that starts with a time of format hh:mm
	insertAt := indexOfLastScheduledTask(m.Tabs[m.Focus])

	// Insert task behind first task with time or a position 0
	m.Tabs[m.Focus].InsertItem(insertAt+1, Task{Name: s, Day: m.Focus})
}

// indexOfLastScheduledTask returns the index of the last task in the given tab whose name starts with a time in the
// format hh:mm.
func indexOfLastScheduledTask(tab Tab) int {
	insertAt := -1
	startsWithTime := regexp.MustCompile(`^\d{2}:\d{2}`)
	for i, item := range tab.Items() {
		if startsWithTime.MatchString(item.(Task).Name) {
			insertAt = i
		}
	}
	return insertAt
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
	ActiveTab = m.Focus
	return m
}

func (m TabModel) PreviousTab() TabModel {
	if m.Focus == 0 {
		m.Focus = 7
	} else {
		m.Focus--
	}
	ActiveTab = m.Focus
	return m
}
func (m TabModel) MoveItem(to int) {
	if item := m.Tabs[m.Focus].SelectedItem(); item != nil && to < 7 && to >= 0 {
		t := item.(Task)
		m.Tabs[m.Focus].RemoveItem(m.Tabs[m.Focus].Index())
		t.Day = to
		insertAt := indexOfLastScheduledTask(m.Tabs[to])
		m.Tabs[to].InsertItem(insertAt+1, t)
	}
}

func (m TabModel) SelectTab(i int) TabModel {
	m.Focus = i
	ActiveTab = m.Focus
	return m
}

type itemDelegate struct {
	panel int
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
	if ActiveTab == d.panel && index == m.Index() {
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
