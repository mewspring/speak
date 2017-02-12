package main

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"golang.org/x/exp/ebnf"
)

// extractTerms returns the terminals used by the given grammar. As a
// precondition, the grammar must have been validated using validate.
func extractTerms(grammar ebnf.Grammar) *terminals {
	terms := &terminals{
		names:  make(map[string]*ebnf.Name),
		tokens: make(map[string]*ebnf.Token),
	}
	// Extract terminals from within the expressions of each non-terminal
	// production.
	for name, prod := range grammar {
		if !isLexical(name) {
			terms.expr(prod.Expr)
		}
	}
	return terms
}

// terminals records the terminals of a grammar.
type terminals struct {
	// Terminal production names.
	names map[string]*ebnf.Name
	// Token terminals.
	tokens map[string]*ebnf.Token
}

// expr extracts the terminals defined within the given expression.
func (terms *terminals) expr(expr ebnf.Expression) {
	switch expr := expr.(type) {
	case nil:
		// empty expression
	case ebnf.Alternative:
		for _, e := range expr {
			terms.expr(e)
		}
	case ebnf.Sequence:
		for _, e := range expr {
			terms.expr(e)
		}
	case *ebnf.Name:
		if name := expr.String; isLexical(name) {
			terms.names[name] = expr
		}
	case *ebnf.Token:
		name := expr.String
		terms.tokens[name] = expr
	case *ebnf.Range:
		panic(fmt.Errorf("internal error: unexpected range `%q … %q` in non-terminal production rule", expr.Begin.String, expr.End.String))
	case *ebnf.Group:
		terms.expr(expr.Body)
	case *ebnf.Option:
		terms.expr(expr.Body)
	case *ebnf.Repetition:
		terms.expr(expr.Body)
	default:
		panic(fmt.Errorf("internal error: unexpected type %T", expr))
	}
}

// isLexical reports whether the given production name denotes a lexical
// production.
func isLexical(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return !unicode.IsUpper(ch)
}
