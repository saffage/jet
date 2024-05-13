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
		&report.IsDebug,
		"debug",
		false,
		"specifies whether to output debug messages",
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

	path := filepath.Clean(args[0])
	fileExt := filepath.Ext(path)

	if fileExt != ".jet" {
		report.Errorf("expected file extension '.jet', not '%s'", fileExt)
		return
	}

	moduleName := filepath.Base(path[:len(path)-len(fileExt)])

	if _, err := token.IsValidIdent(moduleName); err != nil {
		err = errors.Join(fmt.Errorf("invalid module name (file name used as module identifier)"), err)
		report.Errorf(err.Error())
		return
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		report.InternalErrorf("while reading file '%s': %s", err.Error())
		panic(err)
	}

	config.Global.Files[config.MainFileID] = config.FileInfo{
		Name: moduleName,
		Path: path,
		Buf:  bytes.NewBuffer(buf),
	}
	report.Debugf("set file '%s' as main module", path)
	process(config.Global, config.MainFileID)
}
