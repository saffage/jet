package main_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/saffage/jet/cmd"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

const stopAfterFirstError = true

func TestExamples(t *testing.T) {
	t.Skip("local test")

	err := filepath.WalkDir("examples/", dirWalker(t))

	if err != nil {
		t.Error(err)
	}
}

func isJetFile(filename string) bool {
	return filepath.Ext(filename) == ".jet"
}

func dirWalker(t *testing.T) fs.WalkDirFunc {
	rootSkipped := false

	return func(path string, entry fs.DirEntry, err error) error {
		if !rootSkipped {
			rootSkipped = true
			return err
		}
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return fs.SkipDir
		}
		if !isJetFile(entry.Name()) {
			return nil
		}

		file, err := config.ReadFile(path)
		if err != nil {
			return err
		}

		// TODO replace it with the 'check' command
		t.Logf("building: '%s'", path)
		if buildErr := cmd.Build(file); buildErr != nil {
			t.Error("error!")
			report.Report(buildErr)

			if stopAfterFirstError {
				return buildErr
			}
		}

		return nil
	}
}
