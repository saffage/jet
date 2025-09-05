package parser_test

import (
	"testing"

	"github.com/saffage/jet/parser"
)

func TestParser(t *testing.T) {
	p := parser.FromFile(nil)
	_ = p
}
