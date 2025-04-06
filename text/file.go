package text

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	. "github.com/saffage/jet/internal/debug"
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

func (file *File) LineStart(line int) Pos {
	Assert(line > 0)
	Assert(line <= len(file.lines))

	return PosFrom(file.ID, file.lines[line-1])
}

func (file *File) LineContent(pos Pos) string {
	line := file.searchForOffset(file.fix(pos.Offset())) + 1

	if line == 0 {
		return ""
	}

	startIndex := file.LineStart(line).Offset()
	endIndex := len(file.Content)

	if line+1 < len(file.lines) {
		endIndex = file.LineStart(line+1).Offset() - 1
	}

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

func (file *File) fix(offset int) int {
	switch {
	case offset < 0:
		if !Debug {
			return 0
		}

	case offset > len(file.Content):
		if !Debug {
			return len(file.Content)
		}

	default:
		return offset
	}

	if Debug {
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
