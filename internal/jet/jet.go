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
	checker.CheckBuiltInPkgs()

	m, errs := checker.CheckFile(cfg, fileID)
	if len(errs) != 0 {
		report.Error(errs...)
		return
	}

	if config.FlagGenC {
		finfo := cfg.Files[fileID]
		dir := filepath.Dir(finfo.Path)

		err := os.Mkdir(filepath.Join(dir, ".jet"), os.ModePerm)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		dir = filepath.Join(dir, ".jet")
		report.Hintf("emit C...")

		for _, mImported := range m.Imports {
			genCFile(mImported, dir)
		}

		genCFile(m, dir)
	}
}

func genCFile(m *checker.Module, dir string) {
	f, err := os.Create(filepath.Join(dir, m.Name()+"__jet.c"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if errs := cgen.Generate(f, m); len(errs) != 0 {
		report.Error(errs...)
		return
	}
}
