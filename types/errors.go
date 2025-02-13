package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

var (
	ErrIncorrectArityNotEnough = errors.New("incorrect arity, not enough arguments")
	ErrIncorrectArityTooMany   = errors.New("incorrect arity, too many arguments")
	ErrArgTypeMismatch         = errors.New("argument type mismatch")
	ErrUndefinedIdent          = errors.New("identifier is undefined")
)

func errIncorrectArity(args text.Span, expected, got int) report.Builder {
	err := ErrIncorrectArityNotEnough

	if expected < got {
		err = ErrIncorrectArityTooMany
	}

	return report.Build(err).
		Tag("checker").
		SelectionF(args, "Expected %d arguments, got %d", expected, got)
}

func errArgTypeMismatch(
	arg text.Span,
	argType, expectedType Type,
	index int,
	variadic bool,
) report.Builder {
	b := report.Build(ErrArgTypeMismatch).Tag("checker")

	if variadic {
		b.SelectionF(
			arg,
			"expected `%s` for variadic argument, got `%s`",
			Render(expectedType),
			Render(argType),
		)
	} else {
		b.SelectionF(
			arg,
			"expected `%s` for %d-%s argument, got `%s`",
			Render(expectedType),
			index+1,
			ordinalSuffix(index+1),
			Render(argType),
		)
	}

	return b
}

type symbolWithLocalEnv interface {
	Symbol
	Local() *Env
}

func errUndefinedIdent(ident ast.Ident, owner symbolWithLocalEnv) report.Builder {
	selection := report.Selection{Range: ident.Range()}

	if owner != nil && owner.Local().kind == TypeEnv {
		selection.Hint = fmt.Sprintf(
			"type `%s` does not define a member with that name",
			owner.Name(),
		)
	}

	return report.Build(ErrUndefinedIdent).
		Tag("checker").
		SelectionV(selection)
}

func ordinalSuffix(num int) string {
	switch num % 100 {
	case 11, 12, 13:
		return "th"

	default:
		switch num % 10 {
		case 1:
			return "st"

		case 2:
			return "nd"

		case 3:
			return "rd"

		default:
			return "th"
		}
	}
}

func assert(ok bool, message ...any) {
	if !ok {
		if len(message) > 0 {
			panic("assertion failed: " + fmt.Sprint(message...))
		}
		panic("assertion failed")
	}
}

//
//
//
//
//
//
//
//
//
//
//
//

func errExprIsNotAType(expr ast.Node, type_ Type) report.Builder {
	panic("unimplemented")
}

func errUnimplementedFeature(s1, s2 string, span text.Span) report.Builder {
	panic("unimplemented")
}
func errAlreadyDefined(span1, span2 text.Span, what, s string) report.Builder {
	panic("unimplemented")
}

func errIllFormedAst(item ast.Node) report.Builder {
	panic("unimplemented")
}

func errPositionalParamAfterNamed(span1, span2 text.Span) report.Builder {
	panic("unimplemented")
}

func errParamAlreadyDefined(span1, span2 text.Span) report.Builder {
	panic("unimplemented")
}

func errInternal(span text.Span, args ...any) report.Builder {
	panic("unimplemented")
}

func errInvalidSelectorNode(node *ast.Dot, s1, s2 string) report.Builder {
	panic("unimplemented")
}

func errInvalidSelectorExprType(node *ast.Dot, s string) report.Builder {
	panic("unimplemented")
}

func errUnknownExtern(extern *ast.Extern, externName string) report.Builder {
	panic("unimplemented")
}

func errCannotInferTypeOfExpr(node ast.Node) report.Builder {
	panic("unimplemented")
}

func errTypeMismatch(node1, node2 ast.Node, s1, s2 string, todo any) report.Builder {
	panic("unimplemented")
}

func errNotAssignable(node ast.Node) report.Builder {
	panic("unimplemented")
}

func errExpectedConstantValue(node ast.Node, t Type) error {
	panic("unimplemented")
}

func errTypeIsNotParametrized(node ast.Node, t Type) error {
	panic("unimplemented")
}

func errExprIsNotCallable(node ast.Node, t Type) error {
	panic("unimplemented")
}

func errElemTypeMismatch(elem, node ast.Node, t, tListElem Type) error {
	panic("unimplemented")
}

//
//
//

func warnDiscardedFuncDef(span text.Span) report.Builder {
	panic("unimplemented")
}

//
//
//

func suggestionDefinedAt(
	b report.Builder,
	what string,
	where text.Span,
) report.Builder {
	b.SuggestionV(report.Suggestion{
		Message:   what + " is defined here",
		Selection: report.Selection{Range: where},
	})
	return b
}

func suggestionGuessTypeVariant(
	b report.Builder,
	typename string,
	variants []Variant,
) report.Builder {
	const maxGuesses = 5

	var guesses []string
	var n int

	if len(variants) > maxGuesses {
		n = maxGuesses
		guesses = make([]string, 0, maxGuesses+1)
	} else {
		n = len(variants)
		guesses = make([]string, 0, n)
	}

	for i := range n {
		guesses = append(guesses, variants[i].Name)
	}

	if len(guesses) < cap(guesses) {
		guesses = append(guesses, fmt.Sprintf(
			"... and %d more",
			len(variants)-maxGuesses,
		))
	}

	b.SuggestionV(report.Suggestion{
		Message: "type `%s` have the following variants",
		Selection: report.Selection{
			CustomContent: strings.Join(guesses, "\n"),
		},
	})
	return b
}
