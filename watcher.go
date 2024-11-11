package main

import "log"

// watches all the Go files in the directory and calls onChange when a change is detected
// this is a wrapper around platform specific watch() function
func Watch(dir string, onChange func()) {
	log.Printf("watching %s for changes...\n", dir)

	files, err := getAllFiles(dir, ".go")
	if err != nil {
		log.Fatalf("error getting all Go files in directory: %v", err)
	}

	watch(files, onChange)
}
