//go:build darwin || freebsd

package main

import (
	"log"
	"os"
	"syscall"
)

func watch(dir string, onChange func()) {
	files, err := getAllFiles(dir, ".go")
	if err != nil {
		log.Fatalf("error getting all Go files in directory: %v", err)
	}

	// create a new kernel event queue & get a descriptor
	// The kqueue() system call provides a generic method of notifying the user
	// when an kernel event (kevent) happens or a condition holds
	// https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/kqueue.2.html
	kq, err := syscall.Kqueue()
	if err != nil {
		log.Fatalf("error creating kqueue: %v", err)
	}

	changes := make([]syscall.Kevent_t, len(files))
	for i, filename := range files {
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
			// NOTE_WRITE: A write occurred on the file referenced by the descriptor.
			// NOTE_RENAME: The file referenced by the descriptor was renamed.
			// NOTE_DELETE: The unlink() system call was called on the file referenced by the descriptor
			Fflags: syscall.NOTE_WRITE | syscall.NOTE_RENAME | syscall.NOTE_DELETE,

			Udata: nil,
		}
	}

	events := make([]syscall.Kevent_t, 1)
	for {
		_, err := syscall.Kevent(kq, changes, events, nil) // wait indefinitely for events since timeout is nil
		if err != nil {
			log.Printf("error waiting for events: %v", err)
			continue
		}

		onChange()
	}
}
