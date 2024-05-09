package report

import (
	"fmt"
	"os"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/internal/log"
	"github.com/saffage/jet/token"
)

func Note(cfg *config.Config, tag, message string, start, end token.Loc) {
	Report(log.KindNote, cfg, tag, message, start, end)
}

func Hint(cfg *config.Config, tag, message string, start, end token.Loc) {
	Report(log.KindHint, cfg, tag, message, start, end)
}

func Warning(cfg *config.Config, tag, message string, start, end token.Loc) {
	Report(log.KindWarning, cfg, tag, message, start, end)
}

func Error(cfg *config.Config, tag, message string, start, end token.Loc) {
	Report(log.KindError, cfg, tag, message, start, end)
}

func InternalError(cfg *config.Config, tag, message string, start, end token.Loc) {
	Report(log.KindInternalError, cfg, tag, message, start, end)
}

func Report(kind log.Kind, cfg *config.Config, tag, message string, start, end token.Loc) {
	if start.FileID != end.FileID {
		panic("start & end position have different file IDs")
	}

	file := os.Stdout
	report := ""

	if kind == log.KindError {
		file = os.Stderr
	}

	if cfg != nil {
		if fileInfo, ok := cfg.Files[start.FileID]; ok {
			if end.Offset > fileInfo.Buf.Len() {
				panic(fmt.Sprintf(
					"invalid position, offset if bigger than buffer size (%d > %d)",
					end.Offset,
					fileInfo.Buf.Len(),
				))
			}

			report = generateReport(kind, start, end, fileInfo.Buf.Bytes()) + "\n"
		}
	}

	fmt.Fprintf(file, "%s %s\n%s", kind.LabelWithTag(tag), messageStyle.Sprint(message), report)
}
