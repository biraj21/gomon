# gomon - nodemon but for Go files

A simple Go program that watches Go files in a directory and restarts the program when a change is detected. It runs the `go run` command to start the program.

It only works on macOS (and maybe even BSD but I haven't tested it) because it uses macOS's built-in `kqueue()` and `kevent()` system calls. You can read [Apple's docs](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/kqueue.2.html) about them.

## Usage

```
Usage: gomon <file/dir>

Watches the directory for changes and restarts the program when a change is detected.

- file: When a file is specified, the program will watch all .go files in that file's parent directory.
- dir: When a directory is specified, the program will watch all .go files in the directory.
```
