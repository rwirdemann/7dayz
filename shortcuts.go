package perpetask

var (
	KeyNextPanel = "tab"
	KeyPrefPanel = "shift+tab"
)

type Shortcut struct {
	Key  string
	Desc string
}

var General = []Shortcut{
	{KeyNextPanel, "focus next day"},
	{KeyPrefPanel, "focus prev day"},
	{"alt+0", "focus inbox"},
	{"alt+{i}", "focus weekday"},
	{"s", "save"},
	{"?", "toggle help"},
}

var Management = []Shortcut{
	{"backspace", "delete task"},
	{"space", "complete task"},
	{"enter", "edit task"},
	{"n", "new task"},
}

var Movement = []Shortcut{
	{"shift+right", "move task right"},
	{"shift+left", "move task left"},
	{"shift+t", "move task to today"},
	{"shift+i", "move task to inbox"},
}

var Sorting = []Shortcut{
	{"shift+up", "move task up"},
	{"shift+down", "move task down"},
}
