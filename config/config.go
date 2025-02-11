package config

import (
	"sync"

	"github.com/saffage/jet/text"
)

var (
	filesMutex sync.RWMutex
	files      []*text.File

	CompilerFilepath string // Path to the compiler executable.

	Run              bool // Run a compiled executable.
	Debug            bool // Enable debug information.
	NoHints          bool // Disable compiler hints.
	DumpCheckerState bool // Dump the checker state after checking a specified module.
	ParseAst         bool // Display program AST of the specified module and exit.
	TraceParser      bool // Trace parser calls (used for debugging).

	// Disable checking of 'builtin' package
	NoBuiltinPackage bool

	// Path to the 'builtin' package directory,
	// relative to the compiler executable.
	BuiltinPackagePath string = "lib/builtin/"

	// Compiler cache directory.
	CacheDirName string = ".jet"

	CC      string // Path to a C compiler executable.
	CCFlags string // Flags that must be passed to a C compiler.
	LDFlags string // Flags that must be passed to a linker.
)

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

	if len(files) <= int(id) {
		return files[id-1]
	}

	return nil
}
