//go:build darwin || freebsd || linux

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const programName = "gomon"

var usage = fmt.Sprintf(`Usage: %s <file/dir>

Watches the directory for changes and restarts the program when a change is detected.

- file: When a file is specified, the program will watch all .go files in that file's parent directory.
- dir: When a directory is specified, the program will watch all .go files in the directory.
`, programName)

// pointer because the receivers in methods will modify it
var manager = &ProcessManager{}

func init() {
	log.SetPrefix(fmt.Sprintf("[%s] ", programName)) // prefix log lines with program name
	log.SetFlags(log.Ltime)                          // don't show date in logs

	flag.Usage = func() {
		fmt.Print(usage)
		os.Exit(0)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Print(usage)
		os.Exit(1)
	}

	filename := flag.Arg(0)
	filename, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("error getting absolute path: %v", err)
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		log.Fatalf("error getting file info: %v", err)
	}

	var dir string
	if fileInfo.IsDir() {
		dir = filename
	} else {
		dir = filepath.Dir(filename)
	}

	// buffered channel because we don't want the sender (the signal package) to block
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		log.Printf("received signal: %v\n", sig)
		manager.StopProcess()
		os.Exit(0)
	}()

	programArgs := flag.Args()[1:] // pass rest of args to program
	manager.RunProcess(filename, programArgs)

	Watch(dir, func() {
		log.Println()
		log.Println("restarting due to changes...")
		manager.StopProcess()
		manager.RunProcess(filename, programArgs)
	})
}
