package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
)

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

func getByteOrder() binary.ByteOrder {
	var x uint32 = 1 // Define a 32-bit unsigned integer with a value of 1

	// Create a byte slice to hold the binary representation of the number
	buf := make([]byte, 4)

	// Use binary.LittleEndian.PutUint32 to write the value of x to buf in little-endian order
	// This means that the least significant byte (LSB) will be stored first
	binary.LittleEndian.PutUint32(buf, x)

	/**
	 * The number '1' in binary is: 0x00000001 (in hexadecimal, 4 bytes).
	 * In Little Endian, the least significant byte (0x01) is stored first:
	 *
	 *    memory: 0x00 0x01 0x02 0x03
	 *    buf   : 0x01 0x00 0x00 0x00 (in little-endian order)
	 */

	// Now check the first byte (buf[0]) to determine endianness:
	// If the first byte is 0x01, it means the system is little-endian (LSB first)
	if buf[0] == 1 {
		// If buf[0] is 1, it's little endian
		return binary.LittleEndian
	} else {
		// Otherwise, the system is big-endian, where the most significant byte is stored first
		// In big-endian systems, the most significant byte is stored at the lowest address.
		return binary.BigEndian

	}
}
