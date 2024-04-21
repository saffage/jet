package jet

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/config"
)

func runREPL() {
	println("Welcome to the Jet REPL!\nPress 'CTRL + C' to exit")
	cfg, reader := config.New(), bufio.NewScanner(os.Stdin)

	cfg.Files[config.MainFileID] = config.FileInfo{
		Name: "unnamed",
		Path: "",
		Buf:  new(bytes.Buffer),
	}

	for {
		if promt(reader) {
			fileinfo := cfg.Files[config.MainFileID]
			fileinfo.Buf.WriteByte('\n')
			fileinfo.Buf.Write(reader.Bytes())
			cfg.Files[config.MainFileID] = fileinfo

			process(cfg, fileinfo.Buf.Bytes(), config.MainFileID, true /*isREPL*/)
		}

		if reader.Err() != nil {
			panic(reader.Err())
		}
	}
}

func promt(reader *bufio.Scanner) bool {
	print(color.HiCyanString("> "))
	return reader.Scan() && reader.Err() == nil && len(strings.TrimSpace(reader.Text())) != 0
}
