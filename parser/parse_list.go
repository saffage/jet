package parser

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

type (
	parseFunc        func() ast.Node
	parseLabeledFunc func(label *ast.Lower, colon text.Pos) ast.Node
)

// Grammar:
//
//	item {separator item}
func (parse *parser) sequence(item parseFunc, separator ...token.Kind) []ast.Node {
	if parse.pushTrace(stringify(separator)...) {
		defer parse.popTrace()
	}

	nodes := []ast.Node{}

	for parse.Kind != token.EOF {
		// TODO: Determine a reliable way to recover the parser state to allow
		// parsing subsequent items without skipping too many tokens.
		//
		// For example, this would be parsed correctly:
		//
		//	a, b., c
		//
		// But this is not:
		//
		// 	a, b * , c
		nodes = append(nodes, parse.try(item))

		if !parse.skipAny(separator...) {
			break
		}

		parse.skipNewLines()
	}

	return nodes
}

// Grammar:
//
//	open [item {separator item} [separator]] close
func (parse *parser) listOpenClose(
	item parseFunc,
	open, close token.Kind,
	separator ...token.Kind,
) ([]ast.Node, text.Span) {
	if parse.pushTrace(stringify(append([]token.Kind{open, close}, separator...))...) {
		defer parse.popTrace()
	}

	openTok := parse.expect(open)

	parse.skipNewLines()

	nodes := parse.listUntil(item, openTok.Span, close, separator...)
	closeTok := parse.expect(close)

	return nodes, text.Span{
		From: openTok.Span.From,
		To:   closeTok.Span.From,
	}
}

// The begin parameter is used if EOF was reached (i.e. delimiter wasn't found).
//
// Grammar:
//
//	[item {separator item} [separator]] delimiter
func (parse *parser) listUntil(
	item parseFunc,
	begin text.Span,
	delimiter token.Kind,
	separator ...token.Kind,
) []ast.Node {
	delimiters := append([]token.Kind{delimiter}, separator...)

	if parse.pushTrace(stringify(delimiters)...) {
		defer parse.popTrace()
	}

	nodes := []ast.Node{}

	// Possible cases:
	//  - empty list `DELIM`
	//  - regular list `... DELIM`
	//  - with trailing separator `..., DELIM`
	//  - unterminated list `... EOF`
	for !parse.matchAny(delimiter, token.EOF) {
		tracing := parse.pushTrace("item")
		nodeStart := parse.Span.From
		node, err := catch(item)

		if !err.IsValid() {
			if parse.skipAny(separator...) || parse.match(delimiter) {
				parse.skipNewLines()

				if tracing {
					parse.popTrace()
				}

				nodes = append(nodes, node)
				continue
			}

			// The node is correct, but no separator\delimiter was found.
			err = errUnterminatedExpr(node.Range(), delimiters...)
			node = &ast.BadNode{DesiredPos: nodeStart}
		}

		// Something went wrong, advance to separator\delimiter and continue
		// parsing elements until we find the delimiter.
		parse.skipUntil(append(delimiters, token.Newline)...)
		parse.consumeAny(append(separator, token.Newline)...)
		parse.handleError(err)
		parse.popTrace()

		nodes = append(nodes, &ast.BadNode{DesiredPos: nodeStart})
	}

	if parse.match(token.EOF) && delimiter != token.EOF {
		parse.handleError(errUnterminatedList(begin))
	}

	return nodes
}

// The handleError invokes the [token.Scanner.ErrorHandler] if the error
// is valid.
func (parse *parser) handleError(err report.Builder) {
	if err.IsValid() {
		parse.HandleError(err)
	}
}

// The try function catches any error panic while parsing the item and
// replaces node with [*ast.BadNode] in case of error.
//
// This function always return non-nil node.
func (parse *parser) try(item func() ast.Node) ast.Node {
	start := parse.Span.From
	node, err := catch(item)

	if err.IsValid() {
		parse.handleError(err)
		node = &ast.BadNode{DesiredPos: start}
	}

	return node
}

// The ensure function catches any error panic while parsing the item and
// replaces node with nil in case of error.
func (parse *parser) ensure(item func() ast.Node) ast.Node {
	node, err := catch(item)

	if err.IsValid() {
		parse.handleError(err)
		node = nil
	}

	return node
}

// The catch function catches any error panic while parsing the item. If the
// panic value is not an error, it re-panics with the original value.
//
// The item must parse a non-nil node, otherwise this function will panic.
func catch(item func() ast.Node) (node ast.Node, err report.Builder) {
	defer func() {
		if p := recover(); p != nil {
			if b, ok := p.(report.Builder); ok {
				err = b
			} else {
				panic(p)
			}
		}
	}()

	if node = item(); node != nil {
		return
	}

	panic("the error was not emitted while parsing a term, nil node produced")
}

func stringify[T fmt.Stringer](items []T) []string {
	stringified := make([]string, 0, len(items))

	for _, item := range items {
		stringified = append(stringified, item.String())
	}

	return stringified
}
