package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mewmew/speak/internal/ebnf"
)

// isLexical reports whether the given production name denotes a lexical
// production.
func isLexical(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return !unicode.IsUpper(ch)
}

// extract contains data used when extracting terminals from a grammar.
type extract struct {
	// Terminal production names.
	names map[string]*ebnf.Name
	// Token terminals.
	tokens map[string]*ebnf.Token
	// Range terminals.
	ranges map[string]*ebnf.Range
}

// Terminals extracts the terminals defined within the given grammar.
func (extract *extract) Terminals(grammar ebnf.Grammar) (terms []ebnf.Expression) {
	// Sort production keys.
	var keys []string
	for key := range grammar {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Extract terminals from within the expressions of each non-terminal
	// production.
	for _, name := range keys {
		prod := grammar[name]
		if !isLexical(name) {
			extract.Expr(prod.Expr)
		}
	}

	// Sort terminal keys.
	var termNames []string
	for key := range extract.names {
		termNames = append(termNames, key)
	}
	sort.Strings(termNames)
	var termTokens []string
	for key := range extract.tokens {
		termTokens = append(termTokens, key)
	}
	sort.Strings(termTokens)
	var termRanges []string
	for key := range extract.ranges {
		termRanges = append(termRanges, key)
	}
	sort.Strings(termRanges)

	// Return extracted terminals.
	for _, termName := range termNames {
		terms = append(terms, extract.names[termName])
	}
	for _, termToken := range termTokens {
		terms = append(terms, extract.tokens[termToken])
	}
	for _, termRange := range termRanges {
		terms = append(terms, extract.ranges[termRange])
	}
	return terms
}

// Expr extracts the terminals defined within the given expression.
func (extract *extract) Expr(expr ebnf.Expression) {
	switch x := expr.(type) {
	case nil:
		// empty expression
	case ebnf.Alternative:
		for _, e := range x {
			extract.Expr(e)
		}
	case ebnf.Sequence:
		for _, e := range x {
			extract.Expr(e)
		}
	case *ebnf.Name:
		if name := x.String; isLexical(name) {
			extract.names[name] = x
		}
	case *ebnf.Token:
		extract.tokens[x.String] = x
	case *ebnf.Range:
		// TODO: Refactor to make use of RegexpString.
		name := "[" + x.Begin.String + "-" + x.End.String + "]"
		extract.ranges[name] = x
	case *ebnf.Group:
		extract.Expr(x.Body)
	case *ebnf.Option:
		extract.Expr(x.Body)
	case *ebnf.Repetition:
		extract.Expr(x.Body)
	default:
		panic(fmt.Sprintf("internal error: unexpected type %T", expr))
	}
}

// Terminals returns the terminals used by the given grammar. As a precondition,
// the grammar must have been validated using ebnf.Verify.
func Terminals(grammar ebnf.Grammar) []ebnf.Expression {
	extract := &extract{
		names:  make(map[string]*ebnf.Name),
		tokens: make(map[string]*ebnf.Token),
		ranges: make(map[string]*ebnf.Range),
	}
	return extract.Terminals(grammar)
}

// hasAlternatives reports whether the given expression has more than one
// alternative. As a precondition, the grammar must have been validated using
// ebnf.Verify.
func hasAlternatives(grammar ebnf.Grammar, expr ebnf.Expression) bool {
	switch x := expr.(type) {
	case *ebnf.Name:
		prod := grammar[x.String]
		return hasAlternatives(grammar, prod.Expr)
	case ebnf.Alternative:
		return true
	}
	return false
}

// RegexpString returns a regular expression of the given terminal. As a
// precondition, the grammar must have been validated using ebnf.Verify.
func RegexpString(grammar ebnf.Grammar, term ebnf.Expression) string {
	switch x := term.(type) {
	case nil:
		// empty expression
		return ""
	case ebnf.Alternative:
		var ss []string
		for _, e := range x {
			ss = append(ss, RegexpString(grammar, e))
		}
		return strings.Join(ss, "|")
	case ebnf.Sequence:
		var ss []string
		for _, e := range x {
			s := RegexpString(grammar, e)
			if hasAlternatives(grammar, e) {
				s = "(" + s + ")"
			}
			ss = append(ss, s)
		}
		return strings.Join(ss, "")
	case *ebnf.Name:
		prod := grammar[x.String]
		return RegexpString(grammar, prod.Expr)
	case *ebnf.Token:
		return x.String
	case *ebnf.Range:
		// TODO: Excape a and b in [a-b]
		return "[" + x.Begin.String + "-" + x.End.String + "]"
	case *ebnf.Group:
		return "(" + RegexpString(grammar, x.Body) + ")"
	case *ebnf.Option:
		s := RegexpString(grammar, x.Body)
		if hasAlternatives(grammar, x.Body) {
			s = "(" + s + ")"
		}
		return s + "?"
	case *ebnf.Repetition:
		s := RegexpString(grammar, x.Body)
		if hasAlternatives(grammar, x.Body) {
			s = "(" + s + ")"
		}
		return s + "*"
	default:
		panic(fmt.Sprintf("internal error: unexpected type %T", term))
	}
}
