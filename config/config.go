// Package config provides global configuration settings for the compiler.
// These settings are initialized through program arguments when the compiler
// starts and remain unchanged throughout its execution.
//
// The configuration includes options to control the compiler's behavior,
// such as enabling or disabling debug and trace information, path to the
// compiler cache directory, etc.
//
// The package also defines functionality for managing files used by the
// compiler, including creating new files, reading files, and retrieving files
// by their ID.
package config

import (
	"sync"

	"github.com/saffage/jet/text"
)

var (
	filesMutex sync.RWMutex
	files      []*text.File
)

// Mutex is not needed due to one-time initialization.

var (
	// Path to the compiler executable (i.e. the very first program argument).
	CompilerFilepath string

	TraceParser     bool // Trace parser calls (used for debugging).
	Debug           bool // Enable debug information.
	NoHints         bool // Disable compiler hints.
	NoCoreLib       bool // Disable core library.
	NoBuiltinLib    bool // Disable built-in library.
	ShowTimings     bool // Show timings for every compiler stage.
	ShowMemoryUsage bool // Show memory usage for every compiler stage.

	// Compiler cache directory.
	CacheDirName string

	Target  BuildTarget = TargetC
	CC      string      // Path to a C compiler executable.
	CCFlags string      // Flags that must be passed to a C compiler.
	LDFlags string      // Flags that must be passed to a linker.
)

//go:generate go tool stringer -type=BuildTarget -linecomment -output=build_target_string.go
type BuildTarget byte

const (
	TargetC BuildTarget = iota // c
)

func BuildTargetFromString(s string) (BuildTarget, bool) {
	switch s {
	case "c":
		return TargetC, true
	default:
		return 0, false
	}
}

func NewFile(path string, content []byte) (*text.File, error) {
	filesMutex.Lock()
	defer filesMutex.Unlock()

	fileID := text.FileID(len(files) + 1)
	file, err := text.NewFile(fileID, path, content)

	if err != nil {
		return nil, err
	}

	files = append(files, file)
	return file, nil
}

func ReadFile(path string) (*text.File, error) {
	filesMutex.Lock()
	defer filesMutex.Unlock()

	fileID := text.FileID(len(files) + 1)
	file, err := text.ReadFile(fileID, path)

	if err != nil {
		return nil, err
	}

	files = append(files, file)
	return file, nil
}

func File(id text.FileID) *text.File {
	filesMutex.RLock()
	defer filesMutex.RUnlock()

	if len(files) <= int(id) && id != 0 {
		return files[id-1]
	}

	return nil
}
