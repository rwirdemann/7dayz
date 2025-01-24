package _dayz

const (
	Inbox     = 0
	Monday    = 1
	Tuesday   = 2
	Wednesday = 3
	Thursday  = 4
	Friday    = 5
	Saturday  = 6
	Sunday    = 7
)

type Task struct {
	Name string `json:"name"`
	Day  int    `json:"day"`
	Done bool   `json:"done"`
	Pos  int    `json:"pos"`
}

func (t Task) FilterValue() string {
	return ""
}

type TaskRepository interface {
	Load() []Task
	Save(tasks []Task)
}
