package parser

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

type parseFunc func() (ast.Node, error)

// Grammar:
//
//	f {sep f}
func (parse *parser) sequence(f parseFunc, sep token.Kind) ([]ast.Node, error) {
	var nodes []ast.Node

	for {
		node, err := f()

		if err != nil {
			return nodes, err
		}

		nodes = append(nodes, node)

		if !parse.skip(sep) {
			break
		}
	}

	return nodes, nil
}

// Grammar:
//
//	open f {sep f} [sep] close
func (parse *parser) listOpenClose(
	f parseFunc,
	open, close, sep token.Kind,
) (nodes []ast.Node, span text.Span, err error) {
	openTok, ok := parse.consume(open)

	if !ok {
		err = errUnexpectedToken(parse.Span, open)
		return
	}

	span.From = openTok.Span.From
	nodes, err = parse.listUntil(f, close, sep, openTok.Span)

	if err != nil {
		return
	}

	closeTok, ok := parse.consume(close)

	if !ok {
		err = errUnexpectedToken(parse.Span, close)
		return
	}

	span.To = closeTok.Span.From
	return
}

// The begin parameter is used if EOF was reached (i.e. delimiter wasn't found).
//
// Grammar:
//
//	f {sep f} [sep] delim
func (parse *parser) listUntil(
	f parseFunc,
	delimiter, separator token.Kind,
	begin text.Span,
) ([]ast.Node, error) {
	var nodes []ast.Node
	var errs []error

	// Possible cases:
	//  - empty list `{}`
	//  - regular list `{..., f}`
	//  - with trailing separator `{..., f,}`
	//  - unterminated list `{... EOF`
	for !parse.matchAny(delimiter, token.EOF) {
		nodeStart := parse.Span.From
		node, err := f()

		if err == nil {
			if parse.skip(separator) || parse.match(delimiter) {
				nodes = append(nodes, node)
				continue
			}

			// The node is correct, but no separator\delimiter was found.
			err = errUnterminatedExpr(node.Range(), separator, delimiter)
		}

		// Something went wrong, advance to separator\delimiter and continue
		// parsing elements until we find the delimiter.
		parse.skipUntil(separator, delimiter)
		parse.consume(separator)

		errs = append(errs, err)
		nodes = append(nodes, &ast.BadNode{DesiredPos: nodeStart})
	}

	if parse.match(token.EOF) && delimiter != token.EOF {
		errs = append(errs, errUnterminatedList(begin))
	}

	return nodes, report.Join(errs...)
}
