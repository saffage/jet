package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

var (
	ErrInvalidExt = errors.New("the file is expected to have a '.jet' or '.jem' extension")
)

// Represents an ID of the file, processed by a config.
//
// Zero value its an invalid ID.
type FileID uint16

type File struct {
	ID   FileID
	Name string        // File name without extension.
	Path string        // Path to the file.
	Buf  *bytes.Buffer // File content.
}

func (file *File) Line(n int) string {
	if n >= 0 {
		for lineNum, b := 1, file.Buf.Bytes(); ; lineNum++ {
			idx := slices.IndexFunc(b, func(char byte) bool {
				return char == '\r' || char == '\n'
			})
			if idx < 0 {
				// Buffer contains single line.
				idx = len(b)
			}
			if lineNum >= n {
				return string(b[:idx])
			}
			if idx+1 >= len(b) {
				break
			}
			if b[idx] == '\r' {
				idx++
			}
			if b[idx] == '\n' {
				idx++
			}
			if idx >= len(b) {
				break
			}
			b = b[idx:]
		}
	}
	return ""
}

func (cfg *Config) File(id FileID) *File {
	if int(id) <= len(cfg.files) {
		return cfg.files[id-1]
	}

	return nil
}

func (cfg *Config) NewFile() *File {
	file := &File{ID: FileID(len(cfg.files) + 1)}
	cfg.files = append(cfg.files, file)
	return file
}

func (cfg *Config) ReadFile(path string) (*File, error) {
	path = filepath.Clean(path)

	if !fs.ValidPath(path) {
		return nil, fmt.Errorf("argument is not a valid path: %s", path)
	}

	name, data, err := readFile(path)

	if err != nil {
		return nil, err
	}

	file := cfg.NewFile()
	file.Name = name
	file.Path = path
	file.Buf = bytes.NewBuffer(data)
	return file, nil
}

func readFile(path string) (name string, data []byte, err error) {
	stat, err := os.Stat(path)

	if err != nil {
		return
	}

	if !stat.Mode().IsRegular() {
		err = fmt.Errorf("path is not a file: %s", path)
		return
	}

	fileExt := filepath.Ext(path)

	switch fileExt {
	case ".jet", ".jem":
		// OK

	default:
		err = ErrInvalidExt
		return
	}

	name = filepath.Base(path[:len(path)-len(fileExt)])
	data, err = os.ReadFile(path)

	if err != nil {
		err = fmt.Errorf("failed to read file: %s, %w", path, err)
		return
	}

	return
}
