package config

import "bytes"

const (
	_ FileID = iota
	MainFileID
	TypesModuleFileID
	CModuleFileID

	firstFreeFileID
)

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

var fileID = firstFreeFileID
