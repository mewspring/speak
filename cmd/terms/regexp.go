package main

import (
	"fmt"
	"regexp/syntax"
	"unicode/utf8"

	"github.com/pkg/errors"

	"golang.org/x/exp/ebnf"
)

// regexp returns a regular expression of the given terminal. As a precondition,
// the grammar must have been validated using ebnf.Verify.
func regexp(grammar ebnf.Grammar, expr ebnf.Expression) *syntax.Regexp {
	switch expr := expr.(type) {
	case nil:
		// empty expression
		return &syntax.Regexp{
			Op: syntax.OpEmptyMatch,
		}
	case ebnf.Alternative:
		// x | y | z
		var subs []*syntax.Regexp
		for _, e := range expr {
			sub := regexp(grammar, e)
			subs = append(subs, sub)
		}
		return &syntax.Regexp{
			Op:  syntax.OpAlternate,
			Sub: subs,
		}
	case ebnf.Sequence:
		// x y z
		var subs []*syntax.Regexp
		for _, e := range expr {
			sub := regexp(grammar, e)
			subs = append(subs, sub)
		}
		return &syntax.Regexp{
			Op:  syntax.OpConcat,
			Sub: subs,
		}
	case *ebnf.Name:
		// foo
		prod := grammar[expr.String]
		return regexp(grammar, prod.Expr)
	case *ebnf.Token:
		// "foo"
		runes := []rune(expr.String)
		return &syntax.Regexp{
			Op:   syntax.OpLiteral,
			Rune: runes,
		}
	case *ebnf.Range:
		// "a" … "z"
		start, _ := utf8.DecodeRuneInString(expr.Begin.String)
		end, _ := utf8.DecodeRuneInString(expr.End.String)
		runes := []rune{start, end}
		return &syntax.Regexp{
			Op:   syntax.OpCharClass,
			Rune: runes,
		}
	case *ebnf.Group:
		// (body)
		sub := regexp(grammar, expr.Body)
		subs := []*syntax.Regexp{sub}
		return &syntax.Regexp{
			Op:  syntax.OpCapture,
			Sub: subs,
		}
	case *ebnf.Option:
		// [body]
		sub := regexp(grammar, expr.Body)
		subs := []*syntax.Regexp{sub}
		return &syntax.Regexp{
			Op:  syntax.OpQuest,
			Sub: subs,
		}
	case *ebnf.Repetition:
		// {body}
		sub := regexp(grammar, expr.Body)
		subs := []*syntax.Regexp{sub}
		return &syntax.Regexp{
			Op:  syntax.OpStar,
			Sub: subs,
		}
	default:
		panic(fmt.Sprintf("internal error: unexpected type %T", expr))
	}
}

// regexpString returns the string representation of a regular expression of the
// given terminal. As a precondition, the grammar must have been validated using
// ebnf.Verify.
func regexpString(grammar ebnf.Grammar, expr ebnf.Expression) (string, error) {
	reg := regexp(grammar, expr)
	simple, err := syntax.Parse(reg.String(), syntax.Perl)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return simple.String(), nil
}
