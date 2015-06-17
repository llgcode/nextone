package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
)

// DbFileName is the default file name find in user home
const DbFileName = "db.task"

// DbEnvPath is an environment variable that helps configure db path
const DbEnvPath = "NEXTONE_DB_PATH"

var dbFlag = flag.String("db", "", "Db path.")
var tagsFlag = flag.String("t", "", "Filter by tag. Tags list separated by ','.")
var statusFlag = flag.String("s", "", "Filter by status. Status list separated by ','.")
var findTextFlag = flag.String("f", "", "Find text in task.")
var recomputeIDFlag = flag.Bool("recomputeId", false, "Recompute id for all tasks. Warning! this will change all ids.")
var jsonFlag = flag.Bool("json", false, "Print tasks in json format.")

// Task represents what we have to do
type Task struct {
	ID      int      `json:"id"`      // Id of the task
	Created int64    `json:"created"` // timestamp when it has been created
	Text    string   `json:"text"`    // Text description of the Task
	Status  string   `json:"status"`  // Status of the task
	Tags    []string `json:"tags"`    // Tags of the task
}

// TaskByTime sort by timestamp
type TaskByTime []Task

func (t TaskByTime) Len() int           { return len(t) }
func (t TaskByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TaskByTime) Less(i, j int) bool { return t[i].Created < t[j].Created }

// AnsiString provide string with ansi color escapes
func (t Task) AnsiString() string {
	date := time.Unix(t.Created/1000, 0)
	var ansiStatus string
	if t.Status == "pending" || t.Status == "open" {
		ansiStatus = ansi.Color(t.Status, "94")
	} else if t.Status == "done" {
		ansiStatus = ansi.Color(t.Status, "90")
	}
	return fmt.Sprintf("%d %s %s\n  (%s) %s", t.ID, ansiStatus, ansi.Color(t.Text, "80"), ansi.Color(strings.Join(t.Tags, ", "), "90"), date.Format("2006-01-02"))
}

// JSONDb is a Task database in json
type JSONDb struct {
	Tags  []string // List of existing tags that can be used for a task
	Tasks []Task   // List of tasks
}

// FilterByTags return tasks list that have one the tags
func FilterByTags(tasks []Task, tags []string) []Task {
	var result []Task
	for _, task := range tasks {
		if tags[0] == "" || containsOne(tags, task.Tags) {
			result = append(result, task)
		}
	}
	return result
}

// FilterByStatus return tasks list that have one of the status
func FilterByStatus(tasks []Task, status []string) []Task {
	var result []Task
	for _, task := range tasks {
		if status[0] == "" || contains(status, task.Status) {
			result = append(result, task)
		}
	}
	return result
}

// FilterByText return tasks list that have contains text
func FilterByText(tasks []Task, text string) []Task {
	var result []Task
	for _, task := range tasks {
		if strings.Index(strings.ToLower(task.Text), strings.ToLower(text)) != -1 {
			result = append(result, task)
		}
	}
	return result
}

func containsOne(strs1 []string, strs2 []string) bool {
	for _, str1 := range strs1 {
		for _, str2 := range strs2 {
			if strings.ToLower(str1) == strings.ToLower(str2) {
				return true
			}
		}
	}
	return false
}

func contains(strs []string, s string) bool {
	for _, str := range strs {
		if strings.ToLower(str) == strings.ToLower(s) {
			return true
		}
	}
	return false
}

func main() {
	// Ensure that we have an ansi enabled terminal
	ansi.DisableColors(false)
	stdout := colorable.NewColorableStdout()

	flag.Parse()
	// Open file database
	var db JSONDb
	var decoder *json.Decoder
	var dbPath string
	if *dbFlag == "" {
		// open db.task file in user home
		dbPath = os.Getenv(DbEnvPath)
		if dbPath == "" {
			u, err := user.Current()
			if err != nil {
				fmt.Print(err)
				return
			}
			dbPath = filepath.Join(u.HomeDir, DbFileName)
		}

	} else {
		dbPath = *dbFlag
	}
	// open db file
	f, err := os.Open(dbPath)
	if err != nil {
		// Stop if the file opening failed
		fmt.Print(err)
		return
	}
	decoder = json.NewDecoder(f)
	defer f.Close()

	err = decoder.Decode(&db)
	if err != nil {
		fmt.Println(err)
		return
	}

	tags := strings.Split(*tagsFlag, ",")
	status := strings.Split(*statusFlag, ",")

	// sort task
	sort.Sort(TaskByTime(db.Tasks))

	// be sure we have an id
	if *recomputeIDFlag {
		count := 1
		for i := range db.Tasks {
			db.Tasks[i].ID = count
			count++
		}
	}

	// Filter
	db.Tasks = FilterByTags(db.Tasks, tags)
	db.Tasks = FilterByStatus(db.Tasks, status)
	db.Tasks = FilterByText(db.Tasks, *findTextFlag)

	// Print result
	if *jsonFlag {
		result, err := json.MarshalIndent(db, "", " ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintln(stdout, string(result))

	} else {
		for _, task := range db.Tasks {
			fmt.Fprintln(stdout, task.AnsiString())
		}
		fmt.Printf("%d tasks.\n", len(db.Tasks))
	}

}
