package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mgutz/ansi"
)

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

func findByID(tasks []Task, id int) *Task {
	for i, task := range tasks {
		if task.ID == id {
			return &tasks[i]
		}
	}
	return nil
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
