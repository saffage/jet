package text

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	ErrInvalidExt      = errors.New("the file is expected to have a '.jet' or '.jem' extension")
	ErrInvalidFilepath = errors.New("argument is not a valid filepath")
	ErrPathNotFile     = errors.New("path is not a file")
)

// Represents an ID of the file, processed by a config.
//
// Zero value its an invalid ID.
type FileID uint16

type File struct {
	Name    string // File name without extension.
	Path    string // Path to the file.
	Content []byte // File content.
	lines   []int  // Line start offsets.

	ID      FileID
	Flags   FileFlags
	Options FileOptions
}

// Reserved
type FileFlags struct{}

// Reserved
type FileOptions struct{}

func NewFile(id FileID, path string, content []byte) (*File, error) {
	base, ext := filepath.Base(path), filepath.Ext(path)
	name := base[:len(base)-len(ext)]

	switch ext {
	case ".jet", ".jem":
		// OK
	default:
		return nil, ErrInvalidExt
	}

	linesCountGuess := max(16, len(content)/30)
	lines := make([]int, 1, linesCountGuess)
	wasLF := false

	for i := 0; i < len(content) && content[i] != '\000'; i++ {
		// Fixes a case when LF is the last character in the input.
		// For example "…\n\0" will not add the line start index
		// but "…\n…" will.
		if wasLF {
			lines = append(lines, i)
		}

		wasLF = content[i] == '\n'
	}

	return &File{
		ID:      id,
		Name:    name,
		Path:    path,
		Content: content,
		lines:   lines[:len(lines):len(lines)],
	}, nil
}

func ReadFile(id FileID, path string) (*File, error) {
	path = filepath.Clean(path)

	if !fs.ValidPath(path) {
		return nil, ErrInvalidFilepath
	}

	fileInfo, err := os.Stat(path)

	if err != nil {
		err := err.(*os.PathError)
		return nil, fmt.Errorf("failed to read file: '%s', %w", err.Path, err.Err)
	}

	if !fileInfo.Mode().IsRegular() {
		return nil, ErrPathNotFile
	}

	content, readError := os.ReadFile(path)

	if readError != nil {
		return nil, fmt.Errorf("failed to read the file: '%s', %w", path, readError)
	}

	return NewFile(id, path, content)
}

func lineOffsetOfPos(file *File, pos Pos) (offset, index int, found bool) {
	if pos.ID() != file.ID {
		return 0, 0, false
	}

	for i, lineOffset := range file.lines {
		if lineOffset > pos.Offset() {
			found = true
			break
		}
		offset = lineOffset
		index = i
	}
	return
}

func getLineEndOffset(file *File, line int) (offset int) {
	switch {
	case len(file.lines) < line:
		return -1

	case len(file.lines) > line:
		return file.lines[line+1] - 1

	default:
		return len(file.Content)
	}
}

func (file *File) Line(pos Pos) string {
	if lineOffset, lineIndex, found := lineOffsetOfPos(file, pos); found {
		endOffset := getLineEndOffset(file, lineIndex)
		return string(file.Content[lineOffset:endOffset])
	}

	return ""
}

func (file *File) GetPosition(pos Pos) (position Position, valid bool) {
	if lineOffset, lineIndex, found := lineOffsetOfPos(file, pos); found {
		offset := pos.Offset()
		columnOffset := offset - lineOffset

		valid = true
		position = Position{
			Pos:  pos,
			Path: file.Path,
			Line: lineIndex + 1,
		}

		for i := range file.Content {
			if i >= columnOffset {
				position.Char = i + 1
				break
			}
		}
	}

	return
}
