package file

import (
	"encoding/json"
	"github.com/rwirdemann/weekplanner"
	"log"
	"os"
	"path"
)

var base = "tasks.json"

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	base = home + "/.7d/"
}

type TaskRepository struct {
}

func (t TaskRepository) Load() []weekplanner.Task {
	file, err := os.Open(path.Join(base, "tasks.json"))
	if err != nil {
		log.Fatalf("Failed to open tasks.json: %v", err)
	}
	defer file.Close()

	var tasks struct {
		Tasks []weekplanner.Task `json:"tasks"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		log.Fatalf("Failed to decode tasks.json: %v", err)
	}

	return tasks.Tasks
}

func (t TaskRepository) Save(tasks []weekplanner.Task) {
	file, err := os.Create(path.Join(base, "tasks.json"))
	if err != nil {
		log.Fatalf("Failed to create tasks.json: %v", err)
	}
	defer file.Close()

	data := struct {
		Tasks []weekplanner.Task `json:"tasks"`
	}{
		Tasks: tasks,
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to encode tasks to tasks.json: %v", err)
	}
}
