package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const programName = "gomon"

var usage = fmt.Sprintf(`Usage: %s <file/dir>

Watches the directory for changes and restarts the program when a change is detected.

- file: When a file is specified, the program will watch all .go files in that file's parent directory.
- dir: When a directory is specified, the program will watch all .go files in the directory.
`, programName)

type ProcessManager struct {
	mutex    sync.Mutex
	cmd      *exec.Cmd
	pgid     int           // process group ID. will be used to kill the entire process group
	waitDone chan struct{} // channel to signal when Wait() completes
}

func (pm *ProcessManager) runProcess(filename string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// create a new channel for this process
	pm.waitDone = make(chan struct{})

	command := []string{"go", "run", filename}

	log.Printf("Running `%s`", strings.Join(command, " "))

	pm.cmd = exec.Command(command[0], command[1:]...)
	pm.cmd.Stdin = os.Stdin
	pm.cmd.Stdout = os.Stdout
	pm.cmd.Stderr = os.Stderr

	pm.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // create a new process group
	}

	if err := pm.cmd.Start(); err != nil {
		log.Printf("error starting process: %v", err)
		return
	}

	// store the process group ID. will be used to kill the entire process group
	pm.pgid = pm.cmd.Process.Pid

	go func() {
		err := pm.cmd.Wait()
		if err == nil {
			log.Println("clean exit - waiting for changes before restart")
		} else {
			log.Printf("app crashed: %v", err)
			log.Println("waiting for changes before restart")
		}

		// signal that Wait() has completed by closing this channel
		close(pm.waitDone)
	}()
}

func (pm *ProcessManager) stopProcess() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.cmd == nil || pm.cmd.Process == nil {
		return
	}

	// kill the entire process group
	pgidKill := -pm.pgid // negative pgid means kill process group
	syscall.Kill(pgidKill, syscall.SIGTERM)

	// give processes a moment to terminate gracefully
	time.Sleep(100 * time.Millisecond)

	// force kill if still running
	syscall.Kill(pgidKill, syscall.SIGKILL)

	// wait for the anon Wait() goroutine to finish
	<-pm.waitDone

	// if we were to set pm.cmd before reading from pm.waitDone, then it might
	// cause a race condition because the Wait() goroutine uses pm.cmd.Wait()
	// to wait for the process to exit. this is why we read from pm.waitDone
	// first and then set pm.cmd to nil
	pm.cmd = nil
}

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

	// create a new kernel event queue & get a descriptor
	// The kqueue() system call provides a generic method of notifying the user
	// when an kernel event (kevent) happens or a condition holds
	// https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/kqueue.2.html
	kq, err := syscall.Kqueue()
	if err != nil {
		log.Fatalf("error creating kqueue: %v", err)
	}

	allGoFiles, err := getAllFiles(dir, ".go")
	if err != nil {
		log.Fatalf("error getting all Go files in directory: %v", err)
	}

	changes := make([]syscall.Kevent_t, len(allGoFiles))
	for i, filename := range allGoFiles {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("error opening file '%s': %v", filename, err)
		}

		defer file.Close()

		// create a new kevent for the filex
		changes[i] = syscall.Kevent_t{
			// value used to identify this event
			Ident: uint64(file.Fd()), // file descriptor

			// Identifies the kernel filter used to process this event.
			// EVFILT_VNODE: Takes a file descriptor as the identifier and the events to watch for
			// in fflags, and returns when one or more of the requested events
			// occurs on the descriptor (from the above apple doc)
			Filter: syscall.EVFILT_VNODE,

			// Actions to perform on the event.
			// EV_ADD: Add a kevent to the kqueue.
			// EV_CLEAR: After the event is retrieved by the user, its state is reset.
			Flags: syscall.EV_ADD | syscall.EV_CLEAR,

			// Filter-specific flags. - notify on write
			Fflags: syscall.NOTE_WRITE, // NOTE_WRITE: A write occurred on the file referenced by the descriptor.

			Udata: nil,
		}
	}

	// buffered channel because we don't want the sender (the signal package) to block
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		log.Printf("received signal: %v\n", sig)
		manager.stopProcess()
		os.Exit(0)
	}()

	log.Printf("watching %s for changes...\n", dir)
	manager.runProcess(filename)

	events := make([]syscall.Kevent_t, 1)
	for {
		_, err := syscall.Kevent(kq, changes, events, nil) // wait indefinitely for events since timeout is nil
		if err != nil {
			log.Printf("error waiting for events: %v", err)
			continue
		}

		log.Println()
		log.Println("restarting due to changes...")
		manager.stopProcess()
		time.Sleep(100 * time.Millisecond) // give processes some time to cleanup
		manager.runProcess(filename)
	}
}

func getAllFiles(dir string, ext string) ([]string, error) {
	ext = strings.ToLower(strings.Trim(ext, " "))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ext) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
