package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	"github.com/peterh/liner"
)

// DbFileName is the default file name find in user home
const DbFileName = "db.task"

// HistoryFileName is the name of the nextone cmd line history
const HistoryFileName = ".nextone_history"

// DbEnvPath is an environment variable that helps configure db path
const DbEnvPath = "NEXTONE_DB_PATH"

var dbFlag = flag.String("db", "", "Db path.")

var dbPath string

func main() {
	// Ensure that we have an ansi enabled terminal
	ansi.DisableColors(false)
	stdout := colorable.NewColorableStdout()

	flag.Parse()

	u, err := user.Current()
	if *dbFlag == "" {
		// open db.task file in user home
		dbPath = os.Getenv(DbEnvPath)
		if dbPath == "" {
			if err != nil {
				fmt.Print(err)
				return
			}
			dbPath = filepath.Join(u.HomeDir, DbFileName)
		}

	} else {
		dbPath = *dbFlag
	}

	// Open file database
	db, err := openDatabase(dbPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	// sort task
	sort.Sort(TaskByTime(db.Tasks))

	historyPath := filepath.Join(u.HomeDir, HistoryFileName)
	interactive(stdout, db, historyPath)
}

func openDatabase(dbPath string) (*JSONDb, error) {
	var db JSONDb
	// open db file
	f, err := os.Open(dbPath)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(f)
	defer f.Close()

	err = decoder.Decode(&db)
	if err != nil {
		return nil, err
	}

	for _, task := range db.Tasks {
		if task.ID >= db.IDGen {
			db.IDGen = task.ID
		}
	}
	return &db, nil
}

func backupDatabase(dbPath string, suffix string) (err error) {
	// open db file
	in, err := os.Open(dbPath)
	if err != nil {
		return err
	}
	defer in.Close()
	// open db bakup file
	out, err := os.Create(dbPath + suffix)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func saveDatabase(dbPath string, db *JSONDb) (err error) {
	// open db file
	f, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	result, err := json.MarshalIndent(db, "", " ")
	if err != nil {
		return err
	}
	fmt.Fprintln(f, string(result))
	return nil
}

func interactive(stdout io.Writer, db *JSONDb, historyPath string) {
	line := liner.NewLiner()
	defer line.Close()

	f, err := os.Open(historyPath)
	if err == nil {
		defer f.Close()
		line.ReadHistory(f)
	}

	line.SetCtrlCAborts(true)

	line.SetCompleter(func(line string) (c []string) {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd.Name, strings.ToLower(line)) {
				c = append(c, cmd.Name)
			}
		}
		return
	})

	saveHistory := func() {
		fmt.Println("save history")
		f, err := os.Create(historyPath)
		if err != nil {
			log.Print("Error writing history file: ", err)
		} else {
			defer f.Close()
			line.WriteHistory(f)
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// handle ctrl+c event here
		saveHistory()
		os.Exit(0)
	}()

	mainPrompt(line, stdout, db)
	saveHistory()
}

func mainPrompt(line *liner.State, stdout io.Writer, db *JSONDb) {
	if cmdLine, err := line.Prompt("nextone> "); err == nil {
		if cmdLine == "quit" {
			return
		}
		execCommand(stdout, db, line, cmdLine)
		line.AppendHistory(cmdLine)
		mainPrompt(line, stdout, db)
	} else if err == liner.ErrPromptAborted {
		log.Print("Aborted")
	} else {
		log.Print("Error reading line: ", err)
	}
}
