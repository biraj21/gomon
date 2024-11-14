//go:build linux

package main

import (
	"strings"
	"syscall"
)

func watch(dir string, onChange func()) {
	// man page: https://man7.org/linux/man-pages/man7/inotify.7.html

	// create an inotify instance
	ifd, err := syscall.InotifyInit()
	if err != nil {
		panic(err)
	}
	defer syscall.Close(ifd)

	// add an item to watch on a set of events that kernel should monitor
	_, err = syscall.InotifyAddWatch(ifd, dir, syscall.IN_MODIFY|syscall.IN_DELETE|syscall.IN_CREATE|syscall.IN_MOVE)
	if err != nil {
		panic(err)
	}

	// byteOrder := getByteOrder()
	buffer := make([]byte, syscall.SizeofInotifyEvent+syscall.NAME_MAX+1)
	for {
		_, err = syscall.Read(ifd, buffer)
		if err != nil {
			panic(err)
		}

		// var event syscall.InotifyEvent

		// err := binary.Read(bytes.NewReader(buffer), byteOrder, &event)
		// if err != nil {
		// 	fmt.Println("Error reading binary data:", err)
		// 	return
		// }

		// for some reason, event.Len returns the multiples of 16 (16, 32, 64, etc)
		// it depends on the actual length of the name. if the name is let's say 10 byes
		// then event.Len will be 16. it's probably adding the padding for some reason
		// that's why i am ignoring it. and since i'm only interested in the name, i'm
		// not even converting the buffer to InotifyEvent struct

		var nameLength int
		for i, byt := range buffer[syscall.SizeofInotifyEvent:] {
			if byt == 0 {
				nameLength = i
				break
			}

		}

		name := string(buffer[syscall.SizeofInotifyEvent : syscall.SizeofInotifyEvent+nameLength])
		if strings.HasSuffix(name, ".go") {
			onChange()
		}

		// zero out the buffer
		for i := range buffer {
			buffer[i] = 0
		}
	}
}
