# gomon - nodemon but for Go files

A simple Go program that watches Go files in a directory and restarts the program when a change is detected. It runs the `go run` command to start the program.

It only works on macOS (and maybe even BSD but I haven't tested it) because it uses macOS's built-in `kqueue()` and `kevent()` system calls. You can read [Apple's docs](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/kqueue.2.html) about them.

This is not a production-ready program. It's just a fun project to learn about kqueue and kevent.

## Usage

```
Usage: gomon <file/dir>

Watches the directory for changes and restarts the program when a change is detected.

- file: When a file is specified, the program will watch all .go files in that file's parent directory.
- dir: When a directory is specified, the program will watch all .go files in the directory.
```

## Contributing

Contributions are welcome!

### File structure

- _main.go_ is the entry point of the program. It
  1. instantiates a `ProcessManager` struct.
  2. calls `RunProcess()` to start the Go program.
  3. calls `Watch()` to watch the directory for changes.
- _process_manager.go_ containers a `ProcessManager` struct that is used to start and stop the Go program.
  - `RunProcess()` starts the Go program with `go run` command.
  - `StopProcess()` stops the Go program.
- _watcher.go_ contains a `Watch()` function that watches all the Go files in the directory and calls `onChange` when a change is detected. It is actually a wrapper around platform specific `watch()` function. Check _kqueue.go_ and _inotify.go_ for the implementations.

### Guidelines

- Use `log.Println()` instead of `fmt.Println()`.
- Use [conventional commit messages](https://www.conventionalcommits.org/).
