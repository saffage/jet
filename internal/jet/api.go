package jet

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/token"
)

var (
	WriteAstFileHandle *os.File
	ParseAst           = false
	TraceParser        = false
	GenC               = false
)

func ProcessArgs([]string) {
	writeAst := flag.String(
		"writeAst",
		"",
		"output program AST into a specified file is JSON format. Filename can be specified explicitly after the flag (default is 'ast.json')",
	)
	flag.BoolVar(
		&ParseAst,
		"parseAst",
		false,
		"output program AST into the console and stops processing",
	)
	flag.BoolVar(
		&TraceParser,
		"traceParser",
		false,
		"prints the parser function calls in stdout",
	)
	flag.BoolVar(
		&GenC,
		"genC",
		false,
		"generate C file",
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
		report.Errorf("REPL is not implemented")
		return
	}

	stat, err := os.Stat(args[0])
	if err != nil {
		report.Errorf(err.Error())
		return
	}

	if !stat.Mode().IsRegular() {
		report.Errorf("'%s' is not a file", args[0])
		return
	}

	filename := filepath.Clean(args[0])
	fileExt := filepath.Ext(filename)

	if fileExt != ".jet" {
		report.Errorf("expected file extension '.jet', not '%s'", fileExt)
		return
	}

	moduleName := filepath.Base(filename[:len(filename)-len(fileExt)])

	if _, err := token.IsValidIdent(moduleName); err != nil {
		err = errors.Join(fmt.Errorf("invalid module name (file name used as module identifier)"), err)
		report.Errorf(err.Error())
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
	process(config.Global, buf, config.MainFileID)
}
