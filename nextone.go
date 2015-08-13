package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	"github.com/peterh/liner"
)

// DbFileName is the default file name find in user home
const DbFileName = "db.task"

// DbEnvPath is an environment variable that helps configure db path
const DbEnvPath = "NEXTONE_DB_PATH"

var dbFlag = flag.String("db", "", "Db path.")

var dbPath string

func main() {
	// Ensure that we have an ansi enabled terminal
	ansi.DisableColors(false)
	stdout := colorable.NewColorableStdout()

	flag.Parse()

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

	// Open file database
	db, err := openDatabase(dbPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	// sort task
	sort.Sort(TaskByTime(db.Tasks))

	interactive(stdout, db)
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
	return &db, nil
}

func backupDatabase(dbPath string) (err error) {
	// open db file
	in, err := os.Open(dbPath)
	if err != nil {
		return err
	}
	defer in.Close()
	// open db bakup file
	out, err := os.Create(dbPath + "_bak")
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

func interactive(stdout io.Writer, db *JSONDb) {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	line.SetCompleter(func(line string) (c []string) {
		for _, cmd := range commands {
			if strings.HasPrefix(cmd.Name, strings.ToLower(line)) {
				c = append(c, cmd.Name)
			}
		}
		return
	})

	mainPrompt(line, stdout, db)
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
