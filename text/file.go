package text

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/saffage/jet/internal/debug"
)

var (
	ErrInvalidExt      = errors.New("the file is expected to have a '.jet' or '.jem' extension")
	ErrInvalidFilepath = errors.New("argument is not a valid filepath")
	ErrPathNotFile     = errors.New("path is not a file")
)

type File struct {
	Name    string // File name without extension.
	Path    string // Path to the file.
	Content []byte // File content.
	lines   []int  // Line start offsets.

	ID      FileID
	Flags   FileFlags
	Options FileOptions
}

// Represents an ID of the file, processed by a config.
//
// Zero value its an invalid ID.
type FileID uint16

// Reserved
type FileFlags struct{}

// Reserved
type FileOptions struct{}

func NewFile(id FileID, path string, content []byte) (*File, error) {
	name := ""

	if path != "" {
		base, ext := filepath.Base(path), filepath.Ext(path)
		name = base[:len(base)-len(ext)]

		switch ext {
		case ".jet", ".jem":
			// OK
		default:
			return nil, ErrInvalidExt
		}
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

func (file *File) LineStart(line int) Pos {
	debug.Assert(line > 0 && line <= len(file.lines), fmt.Sprintf(
		"line %d is not in valid range [1, %d]",
		line,
		len(file.lines),
	))

	return PosFrom(file.ID, file.lines[line-1])
}

func (file *File) LineEnd(line int) Pos {
	debug.Assert(line > 0 && line <= len(file.lines), fmt.Sprintf(
		"line %d is not in valid range [1, %d]",
		line,
		len(file.lines),
	))

	if line == len(file.lines) {
		if file.Content[len(file.Content)-1] == '\n' {
			return PosFrom(file.ID, len(file.Content)-1)
		}
		return PosFrom(file.ID, len(file.Content))
	}

	return PosFrom(file.ID, file.lines[line]-1)
}

func (file *File) LineContent(line int) string {
	startIndex := file.LineStart(line).Offset()
	endIndex := file.LineEnd(line).Offset()

	return string(file.Content[startIndex:endIndex])
}

func (file *File) PositionOf(pos Pos) Position {
	if pos.ID() != file.ID {
		return Position{}
	}

	offset := file.fix(pos.Offset())
	line, char := file.position(offset)

	return Position{
		Filepath: file.Path,
		ID:       file.ID,
		Offset:   offset,
		Line:     line,
		Char:     char,
	}
}

func (file *File) ByteAt(offset int) byte {
	debug.Assert(offset >= 0 && offset <= len(file.Content), fmt.Sprintf(
		"offset %d is not in valid range [0, %d]",
		offset,
		len(file.Content),
	))

	if offset == len(file.Content) {
		return '\000'
	}

	return file.Content[offset]
}

func (file *File) ByteOf(pos Pos) byte {
	if pos.ID() != file.ID {
		// NOTE: not sure that is a correct behavior.
		return '\000'
	}

	return file.ByteAt(file.fix(pos.Offset()))
}

func (file *File) fix(offset int) int {
	switch {
	case offset < 0:
		if !debug.Enabled {
			return 0
		}

	case offset > len(file.Content):
		if !debug.Enabled {
			return len(file.Content)
		}

	default:
		return offset
	}

	if debug.Enabled {
		panic(fmt.Sprintf("offset %d out of bounds [%d, %d)",
			offset,
			0,
			len(file.Content),
		))
	}

	return 0
}

func (file *File) position(offset int) (line, char int) {
	if offset >= len(file.Content) {
		return len(file.lines), len(file.Content) - file.lines[len(file.lines)-1] + 1
	}
	if i := file.searchForOffset(offset); i >= 0 {
		line = i + 1
		char = offset - file.lines[i] + 1
	}
	return
}

func (file *File) searchForOffset(offset int) (lineIndex int) {
	if offset > file.lines[len(file.lines)-1] && offset < len(file.Content) {
		return len(file.lines) - 1
	}

	lineIndex, found := slices.BinarySearch(file.lines, offset)
	if !found {
		return -1
	}
	return lineIndex
}
