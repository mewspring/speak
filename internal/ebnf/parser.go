// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ebnf

import (
	"io"
	"strconv"
	"text/scanner"
)

type parser struct {
	errors  errorList
	scanner scanner.Scanner
	// token position
	pos scanner.Position
	// one token look-ahead
	tok rune
	// token literal
	lit string
}

func (p *parser) next() {
	p.tok = p.scanner.Scan()
	p.pos = p.scanner.Position
	p.lit = p.scanner.TokenText()
}

func (p *parser) error(pos scanner.Position, msg string) {
	p.errors = append(p.errors, newError(pos, msg))
}

func (p *parser) errorExpected(pos scanner.Position, msg string) {
	msg = `expected "` + msg + `"`
	if pos.Offset == p.pos.Offset {
		// the error happened at the current position; make the error message more
		// specific
		msg += ", found " + scanner.TokenString(p.tok)
		if p.tok < 0 {
			msg += " " + p.lit
		}
	}
	p.error(pos, msg)
}

func (p *parser) expect(tok rune) scanner.Position {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, scanner.TokenString(tok))
	}

	// make progress in any case
	p.next()

	return pos
}

func (p *parser) parseAction() string {
	// TODO: Replace with scanner.Find('»').
	for p.tok != scanner.EOF {
		if p.tok == '»' {
			break
		}
		p.next()
	}
	return "<not yet implemented>"
}

func (p *parser) parseIdentifier() *Name {
	pos := p.pos
	name := p.lit
	p.expect(scanner.Ident)
	return &Name{StringPos: pos, String: name}
}

func (p *parser) parseToken() *Token {
	pos := p.pos
	value := ""
	if p.tok == scanner.String {
		value, _ = strconv.Unquote(p.lit)
		// Unquote may fail with an error, but only if the scanner found an
		// illegal string in the first place. In this case the error has already
		// been reported.
		p.next()
	} else {
		p.expect(scanner.String)
	}
	return &Token{StringPos: pos, String: value}
}

// parseTerm returns nil if no term was found.
func (p *parser) parseTerm() (x Expression) {
	pos := p.pos

	switch p.tok {
	case scanner.Ident:
		x = p.parseIdentifier()

	case scanner.String:
		tok := p.parseToken()
		x = tok
		const ellipsis = '…' // U+2026, the horizontal ellipsis character
		if p.tok == ellipsis {
			p.next()
			x = &Range{Begin: tok, End: p.parseToken()}
		}

	case '(':
		p.next()
		x = &Group{Lparen: pos, Body: p.parseExpression()}
		p.expect(')')

	case '[':
		p.next()
		x = &Option{Lbrack: pos, Body: p.parseExpression()}
		p.expect(']')

	case '{':
		p.next()
		x = &Repetition{Lbrace: pos, Body: p.parseExpression()}
		p.expect('}')
	}

	return x
}

func (p *parser) parseSequence() (x Expression) {
	var list Sequence

	for term := p.parseTerm(); term != nil; term = p.parseTerm() {
		list = append(list, term)
	}

	// no need for a sequence if len(list) < 2
	switch len(list) {
	case 0:
		p.errorExpected(p.pos, "term")
		return &Bad{TokPos: p.pos, Error: "term expected"}
	case 1:
		x = list[0]
	default:
		x = list
	}

	// Parse optional action.
	if p.tok == '«' {
		p.next()
		x = &Action{Expr: x, Larrow: p.pos, Body: p.parseAction()}
		p.expect('»')
	}

	return x
}

func (p *parser) parseExpression() Expression {
	var list Alternatives

	for {
		list = append(list, p.parseSequence())
		// TODO: Parse production actions; if p.tok == '<'
		if p.tok != '|' {
			break
		}
		p.next()
	}
	// len(list) > 0

	// no need for an Alternatives node if len(list) < 2
	if len(list) == 1 {
		return list[0]
	}

	return list
}

func (p *parser) parseProduction() *Production {
	name := p.parseIdentifier()
	p.expect('=')
	var expr Expression
	if p.tok != '.' {
		expr = p.parseExpression()
	}
	p.expect('.')
	return &Production{Name: name, Expr: expr}
}

func (p *parser) parse(filename string, src io.Reader) Grammar {
	p.scanner.Init(src)
	p.scanner.Filename = filename
	// initializes pos, tok, lit
	p.next()

	grammar := make(Grammar)
	for p.tok != scanner.EOF {
		prod := p.parseProduction()
		name := prod.Name.String
		if _, found := grammar[name]; !found {
			grammar[name] = prod
		} else {
			p.error(prod.Pos(), name+" declared already")
		}
	}

	return grammar
}

// Parse parses a set of EBNF productions from source src. It returns a set of
// productions. Errors are reported for incorrect syntax and if a production is
// declared more than once; the filename is used only for error positions.
func Parse(filename string, src io.Reader) (Grammar, error) {
	var p parser
	grammar := p.parse(filename, src)
	return grammar, p.errors.Err()
}
