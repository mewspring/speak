package main

import (
	"fmt"
	"sort"
	"unicode"
	"unicode/utf8"

	"golang.org/x/exp/ebnf"
)

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
		name := x.String
		extract.tokens[name] = x
	case *ebnf.Range:
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

// isLexical reports whether the given production name denotes a lexical
// production.
func isLexical(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return !unicode.IsUpper(ch)
}
