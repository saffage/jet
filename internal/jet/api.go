package jet

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/log"
	"github.com/saffage/jet/token"
)

func ProcessArgs([]string) {
	writeAst := flag.String(
		"writeAst",
		"",
		"output program AST into a specified file is JSON format. Filename can be specified explicitly after the flag (default is 'ast.json')",
	)
	flag.Parse()

	if writeAst != nil && *writeAst != "" {
		filename := *writeAst

		err := error(nil)
		WriteAstFileHandle, err = os.Create(filename)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := WriteAstFileHandle.Close(); err != nil {
				panic(err)
			}
		}()
	}

	args := flag.Args()

	if len(args) == 0 {
		runREPL()
		return
	}

	stat, err := os.Stat(args[0])
	if err != nil {
		log.Error(err.Error())
		return
	}

	if !stat.Mode().IsRegular() {
		log.Error("'%s' is not a file", args[0])
		return
	}

	filename := filepath.Clean(args[0])
	fileExt := filepath.Ext(filename)

	if fileExt != ".jet" {
		log.Error("expected file extension '.jet', not '%s'", fileExt)
		return
	}

	moduleName := filepath.Base(filename[:len(filename)-len(fileExt)])

	if _, err := token.IsValidIdent(moduleName); err != nil {
		err = errors.Join(fmt.Errorf("invalid module name (file name used as module identifier)"), err)
		log.Error(err.Error())
		return
	}

	fmt.Printf("reading file '%s'\n", filename)

	buf, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	config.Global.Files[config.MainFileID] = config.FileInfo{
		Name: moduleName,
		Path: filename,
		Buf:  bytes.NewBuffer(buf),
	}
	process(config.Global, buf, config.MainFileID, false)
}
