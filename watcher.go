package main

import "log"

// watches all the Go files in the directory and calls onChange when a change is detected
// this is a wrapper around platform specific watch() function
func Watch(dir string, onChange func()) {
	log.Printf("watching %s for changes...\n", dir)
	watch(dir, onChange)
}
