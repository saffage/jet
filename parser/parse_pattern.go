package parser

import (
	"reflect"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

type patternKind byte

const (
	patternKindCase patternKind = iota
	patternKindValue
	patternKindParam
)

func (parse *parser) pattern(kind patternKind) ast.Pattern {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	node := ast.Pattern(nil)

	switch {
	case parse.matchAny(token.Int, token.Float, token.String):
		node = &ast.PatternLiteral{
			Value: parse.literal().(*ast.Literal),
		}

	case parse.matchAny(token.LowercaseIdent):
		return parse.patternTypeTestOrRebinding(parse.patternBinding(), kind)

	case parse.match(token.IdentPlaceholder):
		return parse.patternTypeTestOrRebinding(parse.patternPlaceholder(), kind)

	case parse.match(token.UppercaseIdent):
		name := parse.upperNode().(*ast.Capitalized)
		values := []ast.Pattern(nil)
		parens := text.Span{}

		if parse.match(token.LParen) {
			list := parse.parens(
				parse.labeled(
					func(label *ast.Lower, colon text.Pos) ast.Node {
						if label != nil {
							if colon.IsValid() {
								x := parse.try(func() ast.Node {
									return parse.pattern(kind)
								})

								if bad, _ := x.(*ast.BadNode); bad != nil {
									x = &ast.PatternInvalid{Span: text.Span{From: bad.DesiredPos}}
								}

								return &ast.PatternLabeled{
									Label:    label,
									X:        x.(ast.Pattern),
									ColonPos: colon,
								}
							} else {
								panic("TODO")
							}
						} else {
							if colon.IsValid() {
								panic("TODO")
							} else {
								return parse.pattern(kind)
							}
						}
					},
				),
			)

			parens = list.Span
			values = make([]ast.Pattern, 0, len(list.Nodes))

			for _, node := range list.Nodes {
				values = append(values, node.(ast.Pattern))
			}
		}

		node = &ast.PatternVariant{
			Name:   name,
			Values: values,
			Parens: parens,
		}

	case parse.match(token.LBracket):
		list := parse.brackets(func() ast.Node { return parse.patternListItem(kind) })

		node = &ast.PatternList{
			Items:    convertSlice[ast.Pattern](list.Nodes),
			Brackets: list.Span,
		}

	default:
		parse.error(errExpectedPattern(parse.tok.Span, 0))
		panic("unreachable")
	}

	if kind == patternKindCase {
		alternatives := parse.sequence(
			func() ast.Node { return parse.pattern(kind) },
			token.Bar,
		)
		return &ast.PatternAlternative{
			Items: convertSlice[ast.Pattern](alternatives),
		}
	}
	return parse.patternRebinding(node)
}

func (parse *parser) patternBinding() ast.Pattern {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	nameTok := parse.expect(token.LowercaseIdent)
	name := &ast.Lower{Data: nameTok.Data, Span: nameTok.Span}

	return &ast.PatternBinding{Name: name}
}

func (parse *parser) patternPlaceholder() ast.Pattern {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	nameTok := parse.expect(token.IdentPlaceholder)
	name := &ast.Placeholder{Data: nameTok.Data, Span: nameTok.Span}

	return &ast.PatternPlaceholder{Name: name}
}

func (parse *parser) patternTypeTestOrRebinding(x ast.Pattern, kind patternKind) ast.Pattern {
	switch kind {
	case patternKindCase:
		if parse.isTypeStart() {
			return &ast.PatternTypeTest{
				X:    x,
				Type: parse.typeExpr(),
			}
		}

	case patternKindValue, patternKindParam:
		if parse.isTypeStartOrSignature() {
			return &ast.PatternTypeTest{
				X:    x,
				Type: parse.typeExpr(),
			}
		}
	}

	// Useless but why not?
	return parse.patternRebinding(x)
}

func (parse *parser) patternRebinding(pattern ast.Pattern) ast.Pattern {
	if parse.pushTrace() {
		defer parse.popTrace()
	}

	if keywordTok, ok := parse.consume(token.KwAs); ok {
		nameTok := parse.expect(token.LowercaseIdent)
		name := &ast.Lower{Data: nameTok.Data, Span: nameTok.Span}

		// Warn about redundant pattern like `_ as foo`.
		if rebinding, _ := pattern.(*ast.PatternPlaceholder); rebinding != nil {
			span := text.Span{
				From: keywordTok.Span.From,
				To:   rebinding.Range().To,
			}
			parse.handleError(warnRedundantRebinding(span, name.Data))
		}

		return &ast.PatternRebinding{
			X:          pattern,
			Name:       name,
			KeywordPos: keywordTok.Span.From,
		}
	}

	return pattern
}

func (parse *parser) patternListItem(kind patternKind) ast.Node {
	if dots, ok := parse.consume(token.Ellipsis); ok {
		name := parse.lowercaseIdent()

		return &ast.PatternRange{
			Name:     name,
			DotsSpan: dots.Span,
		}
	}

	return parse.pattern(kind)
}

// func parens[T ast.SomeNode](parse *parser, f func() T) ([]T, text.Span) {
// 	if parse.pushTrace() {
// 		defer parse.popTrace()
// 	}

// 	var zero T

// 	nodes, span = parse.listOpenClose(
// 		f,
// 		token.LParen,
// 		token.RParen,
// 		token.Comma,
// 	)

// 	return
// }

// func (parse *parser) brackets(f func() ast.Node) *ast.List {
// 	if parse.pushTrace() {
// 		defer parse.popTrace()
// 	}

// 	nodes, span := parse.listOpenClose(
// 		f,
// 		token.LBracket,
// 		token.RBracket,
// 		token.Comma,
// 	)

// 	return &ast.List{
// 		Nodes: nodes,
// 		Span:  span,
// 	}
// }

// TODO: remove
func convertSlice[T, E any](s []E) []T {
	var t = reflect.TypeFor[T]()
	var casted = make([]T, len(s))

	for i, item := range s {
		var u T
		var v = reflect.ValueOf(item)
		if v.CanConvert(t) {
			u = v.Convert(t).Interface().(T)
		}
		casted[i] = u
	}

	return casted
}
