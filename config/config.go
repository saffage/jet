package config

import "bytes"

var Global *Config

// Must be created via the [New] function.
type Config struct {
	// Contains filenames indexed by their IDs.
	//
	// NOTE: the file on which the compiler was called always has the key [config.MainFileID].
	Files     map[FileID]FileInfo
	MaxErrors int
}

func New() *Config {
	return &Config{
		Files:     map[FileID]FileInfo{},
		MaxErrors: 3,
	}
}

type FileID uint16

const MainFileID FileID = 1

type FileInfo struct {
	Name string        // Module name (file name without extension).
	Path string        // Path to the file.
	Buf  *bytes.Buffer // File content.
}

func init() {
	Global = New()
}
