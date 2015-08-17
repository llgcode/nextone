package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/peterh/liner"
)

type CommandFunc func(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string)

type CommandHandler struct {
	Name string
	Desc string
	Exec CommandFunc
}

var commands = []CommandHandler{}

func init() {
	commands = []CommandHandler{
		{"help", "help <cmd> show help of a command", help},
		{"show", "Show specific task by id", showTask},
		{"list", "List task by status task and search criteria", listTasks},
		{"json", "Print all tasks in json", printJSON},
		{"save", "Save the database", save},
		{"recomputeIds", "Recompute id for all tasks. Warning! this will change all ids.", recomputeIds},
		{"setstatus", "Set task status to done", setStatus},
	}
}

func execCommand(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string) {
	for _, cmd := range commands {
		if strings.HasPrefix(cmdLine, cmd.Name) {
			cmd.Exec(stdout, db, line, cmdLine)
			return
		}
	}
	fmt.Println("type help. Unknown command ", cmdLine)
}

func tokenize(cmdLine string) []string {
	return strings.Split(cmdLine, " ")
}

func help(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string) {
	cmdArgs := tokenize(cmdLine)
	for _, cmd := range commands {
		if len(cmdArgs) == 1 || cmdArgs[1] == cmd.Name {
			fmt.Printf("%s:\n\t%s\n", cmd.Name, cmd.Desc)
		}
	}
}

func showTask(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string) {
	cmdArgs := tokenize(cmdLine)
	id, err := strconv.Atoi(cmdArgs[1])
	if err != nil {
		fmt.Println("First arg need to be an integer")
		return
	}
	task := findByID(db.Tasks, id)
	// Print result
	fmt.Fprintln(stdout, task.AnsiString())

}

func listTasks(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string) {
	cmdArgs := tokenize(cmdLine)
	// Filter
	filteredTasks := db.Tasks
	if len(cmdArgs) > 1 {
		status := strings.Split(cmdArgs[1], ",")
		filteredTasks = FilterByStatus(filteredTasks, status)
	}
	if len(cmdArgs) > 2 {
		tags := strings.Split(cmdArgs[2], ",")
		filteredTasks = FilterByTags(filteredTasks, tags)
	}
	if len(cmdArgs) > 3 {
		search := cmdArgs[3]
		filteredTasks = FilterByText(filteredTasks, search)
	}

	// Print result
	for _, task := range filteredTasks {
		fmt.Fprintln(stdout, task.AnsiString())
	}
	fmt.Printf("%d tasks.\n", len(filteredTasks))

}

func setStatus(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string) {
	cmdArgs := tokenize(cmdLine)
	if len(cmdArgs) == 3 {
		id, err := strconv.Atoi(cmdArgs[1])
		if err != nil {
			fmt.Println("First arg need to be an integer")
			return
		}
		task := findByID(db.Tasks, id)
		task.Status = cmdArgs[2]
		// Print result
		fmt.Fprintln(stdout, task.AnsiString())
	} else {
		fmt.Println("You need to specify task id and new status")
	}
}

func save(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string) {
	backupDatabase(dbPath)
	saveDatabase(dbPath, db)
}

func printJSON(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string) {
	result, err := json.MarshalIndent(db, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprintln(stdout, string(result))
}

func recomputeIds(stdout io.Writer, db *JSONDb, line *liner.State, cmdLine string) {
	count := 1
	for i := range db.Tasks {
		db.Tasks[i].ID = count
		count++
	}
}
