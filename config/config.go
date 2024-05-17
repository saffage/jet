package config

var Global = New()

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
