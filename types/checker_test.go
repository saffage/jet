package types

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

func TestFunctions(t *testing.T) {
	t.Run("ok_int_fn_return_int", checkCodeSnapshotOutput(
		`let foo() Int = 0`,
	))

	t.Run("error_int_fn_return_float", checkCodeSnapshotOutput(
		`let foo() Int = 0.0`,
	))

	t.Run("ok_int_fn_return_let_int", checkCodeSnapshotOutput(`
let foo() Int = {
	let a = 10
	a
}`,
	))

	t.Run("error_int_fn_return_let_int", checkCodeSnapshotOutput(`
let foo() Int = {
	let a = 10.0
	a
}`,
	))
}

func checkCodeSnapshotOutput(code string) func(t *testing.T) {
	return assertErrorConfig(code)
}

func assertErrorConfig(code string) func(t *testing.T) {
	return func(t *testing.T) {
		testName := t.Name()

		// Extract sub-test name
		if idx := strings.Index(testName, "/"); idx >= 0 {
			testName = testName[idx+1:]
		}

		defer func(outFile io.Writer, minDisplayLevel report.Level) {
			report.Output = outFile
		}(report.Output, report.MinDisplayLevel)

		buf := new(bytes.Buffer)
		report.Output = buf

		f, _ := config.NewFile(testName+".jet", []byte(code))
		// fileID := config.NextFileID()
		// cfg.Files[fileID] = config.FileInfo{
		// 	Name: testName,
		// 	Buf:  bytes.NewBuffer([]byte(code)),
		// }

		_, checkErr := CheckFile(f)
		snapshotPath := filepath.Join("./snapshots/", testName+".out")

		// Generate file report.
		report.Report(checkErr)

		if snapshot, err := os.ReadFile(snapshotPath); err != nil {
			// No error expected. Checker output must is empty.
			if os.IsNotExist(err) {
				if buf.Len() != 0 {
					t.Errorf("unexpected error:\n%s", buf.String())
				}
			} else {
				t.Errorf(
					"unexpected error while reading the file (%s): %s",
					snapshotPath,
					err,
				)
			}
		} else {
			// Checker output is not empty. Compare snapshots.
			if buf.Len() == 0 {
				t.Error("error expected, but stderr is empty")
			} else if idx := bytes.Compare(snapshot, buf.Bytes()); idx != 0 {
				t.Errorf("error mismatch:\n%s", buf.String())

				_ = os.WriteFile(
					filepath.Join("./snapshots/", testName+".actual.out"),
					buf.Bytes(),
					0,
				)
			}
		}
	}
}
