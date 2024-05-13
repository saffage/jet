package config

import "bytes"

const MainFileID FileID = 1

type FileID uint16

type FileInfo struct {
	Name string        // File name without extension.
	Path string        // Path to the file.
	Buf  *bytes.Buffer // File content.
}

func NextFileID() FileID {
	id := fileID
	fileID++
	return id
}

var fileID = MainFileID + 1
