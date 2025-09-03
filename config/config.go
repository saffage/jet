package config

import (
	"sync"

	"github.com/saffage/jet/text"
)

var (
	filesMutex sync.RWMutex
	files      []*text.File
)

var (
	CompilerFilepath string // Path to the compiler executable.

	TraceParser      bool // Trace parser calls (used for debugging).
	Debug            bool // Enable debug information.
	NoHints          bool // Disable compiler hints.
	NoBuiltinPackage bool // Disable checking of 'builtin' package

	// Path to the 'builtin' package directory,
	// relative to the compiler executable.
	BuiltinPackagePath string = "lib/builtin/"

	// Compiler cache directory.
	CacheDirName string = ".jet-cache"

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
