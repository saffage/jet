package config

import (
	"bytes"
	"slices"
)

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

func (fi *FileInfo) Line(n int) string {
	if n >= 0 {
		for lineNum, b := 1, fi.Buf.Bytes(); ; lineNum++ {
			idx := slices.IndexFunc(b, func(char byte) bool { return char == '\r' || char == '\n' })
			if idx < 0 {
				break
			}
			if lineNum >= n {
				return string(b[:idx])
			}
			if idx+1 >= len(b) {
				break
			}
			if b[idx] == '\r' {
				idx++
			}
			if b[idx] == '\n' {
				idx++
			}
			if idx >= len(b) {
				break
			}
			b = b[idx:]
		}
	}
	return ""
}

func NextFileID() FileID {
	fileID++
	return fileID - 1
}

var fileID = firstFreeFileID
