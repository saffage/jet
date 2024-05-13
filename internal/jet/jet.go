package jet

import (
	"os"
	"path/filepath"

	"github.com/saffage/jet/cgen"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/report"
)

func process(cfg *config.Config, fileID config.FileID) {
	// if WriteAstFileHandle != nil && nodeList != nil {
	// 	if bytes, err := json.MarshalIndent(nodeList.Nodes, "", "  "); err == nil {
	// 		_, err := WriteAstFileHandle.Write(bytes)
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		fmt.Printf("AST is writed to %s\n", WriteAstFileHandle.Name())
	// 	} else {
	// 		panic(err)
	// 	}
	// }

	m, errs := checker.CheckFile(cfg, fileID)
	if len(errs) != 0 {
		report.Report(errs...)
		return
	}

	if GenC {
		finfo := cfg.Files[fileID]
		dir := filepath.Dir(finfo.Path)
		f, err := os.Create(filepath.Join(dir, "out.c"))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		errs = cgen.Generate(f, m)

		if len(errs) != 0 {
			report.Report(errs...)
			return
		}
	}
}
