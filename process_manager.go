package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ProcessManager struct {
	mutex    sync.Mutex
	cmd      *exec.Cmd
	pgid     int           // process group ID. will be used to kill the entire process group
	waitDone chan struct{} // channel to signal when Wait() completes
}

// runs a go run command
func (pm *ProcessManager) RunProcess(filename string, programArgs []string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// create a new channel for this process
	pm.waitDone = make(chan struct{})

	command := []string{"go", "run", filename}
	command = append(command, programArgs...)

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

// stops the process
func (pm *ProcessManager) StopProcess() {
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
