package config

var Global = &Config{}

type Config struct {
	// Contains filenames indexed by their IDs.
	//
	// The file on which the compiler was called always has the key [MainFileID].
	Files     map[FileID]FileInfo
	MaxErrors int
	Exe       string // Path to the compiler executable.
	Flags     Flags
	Options   Options
}

type Flags struct {
	Run       bool // Run a compiled executable.
	Debug     bool // Enable debug information.
	NoHints   bool // Disable compiler hints.
	NoCoreLib bool // Disable the language core library.
}

type Options struct {
	CacheDir    string // Compiler cache directory.
	CoreLibPath string // Path to the language core library.
	CC          string // Path to a C compiler executable.
	CCFlags     string // Flags that must be passed to a C compiler.
	LDFlags     string // Flags that must be passed to a linker.
}
