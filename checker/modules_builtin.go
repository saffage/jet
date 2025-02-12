package checker

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

var (
	ErrWhileCheckingPackageCore = errors.New("while checking package 'builtin'")
	ErrPackageWasNotFound       = errors.New("package was not found")
	ErrInvalidPackagePath       = errors.New("invalid package path")
)

const (
	// This module contains the declaration of the Jet built-in types.
	builtinModuleName  = "builtin"
	builtinModuleIndex = 0

	// This module contains C type declarations and other tools for
	// interacting with the C backend.
	cModuleName  = "c"
	cModuleIndex = 1
)

const builtinFilesCount = len(builtinFiles)

var (
	builtinFiles = [...]string{
		builtinModuleIndex: builtinModuleName,
		cModuleIndex:       cModuleName,
	}

	builtinMutex      sync.RWMutex
	builtinModules    [builtinFilesCount]*Module
	builtinFileStatus [builtinFilesCount]FileStatus
)

var once sync.Once

func CheckBuiltInPackage() error {
	var err error
	once.Do(func() { err = checkBuiltInPackage() })
	return err
}

type FileStatus byte

const (
	StatusUnchecked  FileStatus = iota // unchecked
	StatusInProgress                   // in progress
	StatusSuccess                      // checked
	StatusFailure                      // failed to check
	StatusIgnored                      // ignored
	StatusMissing                      // missing
)

var (
	uncheckedColor  = color.New(color.FgBlack)
	inProgressColor = color.New(color.FgBlue)
	successColor    = color.New(color.FgGreen)
	failureColor    = color.New(color.FgRed)
	ignoredColor    = color.New(color.FgBlack)
	missingColor    = color.New(color.FgHiRed)
)

func statusColor(status FileStatus) *color.Color {
	switch status {
	case StatusUnchecked:
		return uncheckedColor

	case StatusInProgress:
		return inProgressColor

	case StatusSuccess:
		return successColor

	case StatusFailure:
		return failureColor

	case StatusIgnored:
		return ignoredColor

	case StatusMissing:
		return missingColor

	default:
		panic("invalid enum value")
	}
}

func statusChar(status FileStatus) rune {
	// TODO show legend to user in verbose mode
	switch status {
	case StatusUnchecked:
		return '-'

	case StatusInProgress:
		return '*'

	case StatusSuccess:
		return '✓'

	case StatusFailure:
		return '!'

	case StatusIgnored:
		return '#'

	case StatusMissing:
		return '?'

	default:
		panic("invalid enum value")
	}
}

func checkBuiltInPackage() error {
	if config.NoBuiltinPackage {
		return nil
	}

	report.Hint("checking package 'core'")

	compilerPath := filepath.Dir(config.CompilerFilepath)
	builtinPackagePath := filepath.Join(compilerPath, config.BuiltinPackagePath)

	if dir, err := os.Stat(builtinPackagePath); errors.Is(err, fs.ErrNotExist) ||
		(dir != nil && !dir.IsDir()) {

		return report.Join(ErrWhileCheckingPackageCore, ErrPackageWasNotFound)
	}

	corePkgDir := filepath.Join(builtinPackagePath, "core")

	if dir, err := os.Stat(corePkgDir); errors.Is(err, fs.ErrNotExist) ||
		(dir != nil && !dir.IsDir()) {

		return report.Join(ErrWhileCheckingPackageCore, ErrPackageWasNotFound)
	}

	corePkgFiles, err := os.ReadDir(corePkgDir)

	if err != nil {
		return errors.Join(errors.New("while reading package 'core'"), err)
	}

	var files [builtinFilesCount]string

	for _, entry := range corePkgFiles {
		if entry.Type().IsRegular() {
			name := entry.Name()
			ext := filepath.Ext(name)
			base := name[:len(name)-len(ext)]

			if ext == ".jet" {
				if fileIndex := slices.Index(builtinFiles[:], base); fileIndex != -1 {
					files[fileIndex] = filepath.Join(corePkgDir, name)
				}
			}
		}
	}

	// FIXME test it
	var info *report.Info
	var errs = []error{nil}
	var hasMissingFiles = false

	for index, path := range files {
		if path == "" {
			if info == nil {
				info = &report.Info{Title: "missing 'builtin' package files"}
				errs[0] = info
			}

			builtinFileStatus[index] = StatusMissing
			hasMissingFiles = true
		} else {
			builtinFileStatus[index] = StatusInProgress
			file, err := config.ReadFile(path)

			if err != nil {
				errs = append(errs, report.Wrapf(err, "failed to read file: %s", path))
				builtinFileStatus[index] = StatusFailure
				continue
			}

			module, err := CheckFile(file)

			if err != nil {
				errs = append(errs, report.Wrapf(err, "failed to check file: %s", path))
				builtinFileStatus[index] = StatusFailure
				continue
			}

			builtinModules[index] = module

			if index == builtinModuleIndex {
				for _, sym := range module.Scope.symbols {
					_ = Global.Define(sym)
				}
			}

			builtinFileStatus[index] = StatusSuccess
		}
	}

	if info != nil && hasMissingFiles {
		buf := strings.Builder{}
		buf.WriteString("package 'core' have fixed file set that needs to be " +
			"checked before anything else")

		for index, status := range builtinFileStatus {
			buf.WriteString("\n\t")

			module := builtinFiles[index]
			color := statusColor(status)
			char := statusChar(status)

			color.Fprintf(&buf, "%c %s", char, module)
		}

		info.Suggestions = append(info.Suggestions, report.Suggestion{
			Message: buf.String(),
		})
	}

	if len(errs) > 1 || errs[0] != nil {
		return report.Join(append([]error{ErrWhileCheckingPackageCore}, errs...)...)
	}

	return nil
}
