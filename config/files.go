package config

import "bytes"

const (
	_ FileID = iota
	BuiltInModuleFileID
	CModuleFileID
	MainFileID

	firstFreeFileID
)

type FileID uint16

type FileInfo struct {
	Name string        // File name without extension.
	Path string        // Path to the file.
	Buf  *bytes.Buffer // File content.
}

func NextFileID() FileID {
	fileID++
	return fileID - 1
}

var fileID = firstFreeFileID
