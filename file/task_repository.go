package file

import (
	"encoding/json"
	"github.com/rwirdemann/perpetask"
	"io"
	"log"
	"os"
	"path"
	"time"
)

var base = "tasks.json"

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	base = home + "/.perpetask/"
}

type TaskRepository struct {
}

func (t TaskRepository) Load() []perpetask.Task {
	file, err := os.Open(path.Join(base, "tasks.json"))
	if err != nil {
		log.Printf("Failed to open tasks.json: %v", err)
		return []perpetask.Task{}
	}
	defer file.Close()

	var tasks struct {
		Tasks []perpetask.Task `json:"tasks"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		log.Printf("Failed to decode tasks.json: %v", err)
		return []perpetask.Task{}
	}

	return tasks.Tasks
}

func (t TaskRepository) Save(tasks []perpetask.Task) {

	// Archive existing tasks.json
	archive()

	file, err := os.Create(path.Join(base, "tasks.json"))
	if err != nil {
		log.Fatalf("Failed to create tasks.json: %v", err)
	}
	defer file.Close()

	data := struct {
		Tasks []perpetask.Task `json:"tasks"`
	}{
		Tasks: tasks,
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to encode tasks to tasks.json: %v", err)
	}
}

func archive() {
	currentTime := time.Now().Format("20060102T150405")
	backupFileName := path.Join(base, "tasks_"+currentTime+".json")

	inputFile, err := os.Open(path.Join(base, "tasks.json"))
	if err != nil {
		log.Printf("Failed to open tasks.json for backup: %v", err)
		return
	}
	defer inputFile.Close()

	outputFile, err := os.Create(backupFileName)
	if err != nil {
		log.Printf("Failed to create backup file: %v", err)
		return
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		log.Printf("Failed to copy data to backup file: %v", err)
		return
	}
	log.Printf("Backup of tasks.json created at %s", backupFileName)
}
