// TODO: optimize by calculating first sets.

// Speak parses input by runtime evaluation of language grammars expressed in
// EBNF.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mewkiz/pkg/ioutilx"
	"github.com/mewkiz/pkg/term"
	"github.com/pkg/errors"
	"golang.org/x/exp/ebnf"
)

var (
	// dbg is a logger with the "speak:" prefix which logs debug messages to
	// standard error.
	dbg = log.New(ioutil.Discard, term.MagentaBold("speak:")+" ", 0)
	// warn is a logger with the "speak:" prefix which logs warning messages to
	// standard error.
	warn = log.New(ioutil.Discard, term.RedBold("speak:")+" ", 0)
)

func usage() {
	const use = `
Usage: speak [OPTION]... FILE...

Flags:`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line arguments.
	var (
		// path to EBNF grammar
		grammarPath string
		// Start production rule.
		start string
	)
	flag.StringVar(&grammarPath, "grammar", "grammar.ebnf", "path to EBNF grammar")
	flag.StringVar(&start, "start", "", "start production rule")
	flag.Usage = usage
	flag.Parse()

	// Parse and validate grammar.
	grammar, firstProd, err := parseGrammar(grammarPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if len(start) == 0 {
		start = firstProd
	}
	dbg.Println("start:", start)
	// Remove skip before validate.
	skip, ok := grammar["skip"]
	// TODO: Remove skip production rules recursively before validate.
	if ok {
		delete(grammar, "skip")
	}
	if err := ebnf.Verify(grammar, start); err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}
	// Add skip after validate.
	if ok {
		grammar["skip"] = skip
	}

	// Parse input by runtime evaluation of the grammar.
	for _, inputPath := range flag.Args() {
		input, err := ioutilx.ReadFile(inputPath)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		if err := speak(grammar, start, input); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

// speak parses the given input by runtime evaluation of the grammar from the
// start production rule.
func speak(grammar ebnf.Grammar, start string, input []byte) error {
	p := &parser{
		grammar: grammar,
		input:   input,
	}
	// Calculate first set.
	//first := p.firstSet(grammar)
	//pretty.Println("first:", first)
	//return nil
	ret := p.evalProd(p.grammar[start])
	p.skip()
	dbg.Println("speak:")
	dbg.Printf("   speak.ret: %v", ret)
	dbg.Printf("   speak.len: %v %v", len(input), p.pos)
	return nil
}

// parser holds the state of the EBNF grammar used for parsing.
type parser struct {
	// EBNF language grammar.
	grammar ebnf.Grammar
	// Input source.
	input []byte
	// Current position in input source.
	pos int
	// End of input has been reached.
	eof bool
	// Currently skipping whitespace and comments in evalExpr.
	skipping bool
}

// skip evaluates the skip production rule to ignore whitespace and comments.
func (p *parser) skip() {
	if p.skipping {
		return
	}
	p.skipping = true
	if skip, ok := p.grammar["skip"]; ok {
		dbg.Println("skip:", exprString(skip))
		// record pos, and reset if no whitespace found.
		for {
			bak := p.pos
			if !p.evalExpr(skip.Expr) {
				// reset pos.
				p.pos = bak
				break
			}
		}
	}
	p.skipping = false
}

func (p *parser) evalProd(x *ebnf.Production) bool {
	dbg.Println("evalProd:", exprString(x))
	ret := p.evalExpr(x.Expr)
	dbg.Printf("   evalProd.ret: %v", ret)
	return ret
}

func (p *parser) evalExpr(x ebnf.Expression) bool {
	dbg.Println("evalExpr:", exprString(x))
	// skip whitespace and comments in between expressions.
	p.skip()
	switch x := x.(type) {
	case *ebnf.Production:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	case ebnf.Alternative:
		ret := p.evalAlt(x)
		dbg.Printf("   evalExpr.evalAlt.ret: %v", ret)
		return ret
	case ebnf.Sequence:
		ret := p.evalSeq(x)
		dbg.Printf("   evalExpr.evalSeq.ret: %v", ret)
		return ret
	case *ebnf.Name:
		ret := p.evalName(x)
		dbg.Printf("   evalExpr.evalName.ret: %v", ret)
		return ret
	case *ebnf.Token:
		ret := p.evalToken(x)
		dbg.Printf("   evalExpr.evalToken.ret: %v", ret)
		return ret
	case *ebnf.Range:
		ret := p.evalRange(x)
		dbg.Printf("   evalExpr.evalRange.ret: %v", ret)
		return ret
	case *ebnf.Group:
		ret := p.evalGroup(x)
		dbg.Printf("   evalExpr.evalGroup.ret: %v", ret)
		return ret
	case *ebnf.Option:
		ret := p.evalOpt(x)
		dbg.Printf("   evalExpr.evalOpt.ret: %v", ret)
		return ret
	case *ebnf.Repetition:
		ret := p.evalRep(x)
		dbg.Printf("   evalExpr.evalRep.ret: %v", ret)
		return ret
	default:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	}
}

// evalAlt evaluates a list of alternative expressions. One must be valid.
//
//    x | y | z
func (p *parser) evalAlt(x ebnf.Alternative) bool {
	dbg.Println("evalAlt:", exprString(x))
	// TODO: Figure out how to try handle multiple valid alternatives. Is this
	// even needed?
	for _, e := range x {
		// record pos, and reset for invalid alternatives.
		bak := p.pos
		if p.evalExpr(e) {
			return true
		}
		// reset pos.
		p.pos = bak
	}
	return false
}

// evalSeq evaluates a list of sequential expressions. All must be valid.
//
//    x y z
func (p *parser) evalSeq(x ebnf.Sequence) bool {
	dbg.Println("evalSeq:", exprString(x))
	for _, e := range x {
		if !p.evalExpr(e) {
			return false
		}
	}
	return true
}

// evalName evaluates a the expression of a production name. Must be valid.
//
//    foo
func (p *parser) evalName(x *ebnf.Name) bool {
	dbg.Println("evalName:", exprString(x))
	prod := p.grammar[x.String]
	return p.evalProd(prod)
}

// evalToken evaluates a literal. Must be valid.
//
//    "foo"
func (p *parser) evalToken(x *ebnf.Token) bool {
	dbg.Println("evalToken:", exprString(x))
	for _, q := range x.String {
		r := p.nextRune()
		if r == eof {
			if !p.skipping {
				warn.Printf("unexpected EOF when evaluating token %v", exprString(x))
			}
			return false
		}
		if r != q {
			if !p.skipping {
				warn.Printf("   mismatch %q (expected %q)", r, q)
			}
			return false
		}
		dbg.Printf("   match %q", r)
	}
	return true
}

// evalRange evaluates a range of characters. Must be valid.
//
//    a … z
func (p *parser) evalRange(x *ebnf.Range) bool {
	dbg.Println("evalRange:", exprString(x))
	from, _ := utf8.DecodeRuneInString(x.Begin.String)
	to, _ := utf8.DecodeRuneInString(x.End.String)
	r := p.nextRune()
	if r == eof {
		if !p.skipping {
			warn.Printf("unexpected EOF when evaluating range %v", exprString(x))
		}
		return false
	}
	ret := from <= r && r <= to
	if ret {
		dbg.Printf("   match: %q in %q … %q", r, from, to)
	} else {
		if !p.skipping {
			warn.Printf("   mismatch: %q not in %q … %q", r, from, to)
		}
	}
	return ret
}

// evalGroup evaluates a grouped expression. Must be valid.
//
//    ( body )
func (p *parser) evalGroup(x *ebnf.Group) bool {
	dbg.Println("evalGroup:", exprString(x))
	return p.evalExpr(x.Body)
}

// evalOpt evaluates an optional expression. Must have zero or one valid
// expressions.
//
//    [ body ]
func (p *parser) evalOpt(x *ebnf.Option) bool {
	dbg.Println("evalOpt:", exprString(x))
	// store position and try to parse the optional.
	bak := p.pos
	// EOF is valid in option
	if !p.eof && !p.evalExpr(x.Body) {
		// invalid body is valid in option
		// reset position
		p.pos = bak
	}
	return true
}

// evalRep evaluates a repeated expression. Must have zero or more valid
// expressions.
//
//    { body }
func (p *parser) evalRep(x *ebnf.Repetition) bool {
	dbg.Println("evalRep:", exprString(x))
	// EOF is valid in repetition
	for !p.eof {
		// store position and try to parse a repetition.
		bak := p.pos
		fmt.Println("bak:", bak)
		if !p.evalExpr(x.Body) {
			// invalid body is valid in repetition
			// reset position
			fmt.Println("p.pos:", p.pos)
			p.pos = bak
			break
		}
	}
	return true
}

// ### [ Helper functions ] ####################################################

// exprString returns the string representation of the given EBNF expression.
func exprString(x ebnf.Expression) string {
	switch x := x.(type) {
	case *ebnf.Production:
		return fmt.Sprintf("%v = %v .", exprString(x.Name), exprString(x.Expr))
	case ebnf.Alternative:
		buf := strings.Builder{}
		for i, e := range x {
			if i != 0 {
				buf.WriteString(" | ")
			}
			buf.WriteString(exprString(e))
		}
		return buf.String()
	case ebnf.Sequence:
		buf := strings.Builder{}
		for i, e := range x {
			if i != 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(exprString(e))
		}
		return buf.String()
	case *ebnf.Name:
		return x.String
	case *ebnf.Token:
		return fmt.Sprintf("%q", x.String)
	case *ebnf.Range:
		return fmt.Sprintf("%v … %v", exprString(x.Begin), exprString(x.End))
	case *ebnf.Group:
		return fmt.Sprintf("( %v )", exprString(x.Body))
	case *ebnf.Option:
		return fmt.Sprintf("[ %v ]", exprString(x.Body))
	case *ebnf.Repetition:
		return fmt.Sprintf("{ %v }", exprString(x.Body))
	default:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	}
}

// eof signals end of input.
const eof rune = -1

// nextRune returns the next Unicode rune of the input source.
func (p *parser) nextRune() rune {
	if p.pos >= len(p.input) {
		p.eof = true
		fmt.Println("eof")
		return eof
	}
	r, size := utf8.DecodeRune(p.input[p.pos:])
	p.pos += size
	fmt.Println("pos:", p.pos, len(p.input))
	return r
}

// parseGrammar parses the given EBNF grammar and determines its start
// production rule.
func parseGrammar(grammarPath string) (ebnf.Grammar, string, error) {
	f, err := os.Open(grammarPath)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}
	defer f.Close()
	br := bufio.NewReader(f)
	grammar, err := ebnf.Parse(grammarPath, br)
	if err != nil {
		return nil, "", errors.WithStack(err)
	}
	// Find first syntactic production rule by minimum file offset.
	var firstProd string
	min := -1
	for name, prod := range grammar {
		r, _ := utf8.DecodeRuneInString(name)
		if unicode.IsUpper(r) {
			off := prod.Name.Pos().Offset
			if min == -1 || off < min {
				firstProd = name
				min = off
			}
		}
	}
	if len(firstProd) == 0 {
		return nil, "", errors.Errorf("unable to located first syntactic production rule (capital letter) in grammar %q", grammarPath)
	}
	return grammar, firstProd, nil
}

func (p *parser) firstSet(grammar ebnf.Grammar) map[string]map[rune]bool {
	m := make(map[string]map[rune]bool)
	for name, prod := range grammar {
		m[name] = make(map[rune]bool)
		p.firstProd(prod, m, name)
	}
	return m
}

func (p *parser) firstProd(x *ebnf.Production, m map[string]map[rune]bool, name string) bool {
	return p.firstExpr(x.Expr, m, name)
}

func (p *parser) firstExpr(x ebnf.Expression, m map[string]map[rune]bool, name string) bool {
	switch x := x.(type) {
	case *ebnf.Production:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	case ebnf.Alternative:
		return p.firstAlt(x, m, name)
	case ebnf.Sequence:
		return p.firstSeq(x, m, name)
	case *ebnf.Name:
		return p.firstName(x, m, name)
	case *ebnf.Token:
		return p.firstToken(x, m, name)
	case *ebnf.Range:
		return p.firstRange(x, m, name)
	case *ebnf.Group:
		return p.firstGroup(x, m, name)
	case *ebnf.Option:
		return p.firstOpt(x, m, name)
	case *ebnf.Repetition:
		return p.firstRep(x, m, name)
	default:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	}
}

// the return value report whether the expression can be empty.
func (p *parser) firstAlt(x ebnf.Alternative, m map[string]map[rune]bool, name string) bool {
	empty := false
	for _, e := range x {
		if p.firstExpr(e, m, name) {
			empty = true
		}
	}
	return empty
}

func (p *parser) firstSeq(x ebnf.Sequence, m map[string]map[rune]bool, name string) bool {
	empty := true
	for _, e := range x {
		if !p.firstExpr(e, m, name) {
			empty = false
		}
	}
	return empty
}

func (p *parser) firstName(x *ebnf.Name, m map[string]map[rune]bool, name string) bool {
	return p.firstProd(p.grammar[x.String], m, name)
}

func (p *parser) firstToken(x *ebnf.Token, m map[string]map[rune]bool, name string) bool {
	r, _ := utf8.DecodeRuneInString(x.String)
	m[name][r] = true
	return false
}

func (p *parser) firstRange(x *ebnf.Range, m map[string]map[rune]bool, name string) bool {
	from, _ := utf8.DecodeRuneInString(x.Begin.String)
	to, _ := utf8.DecodeRuneInString(x.End.String)
	for r := from; r <= to; r++ {
		m[name][r] = true
	}
	return false
}

func (p *parser) firstGroup(x *ebnf.Group, m map[string]map[rune]bool, name string) bool {
	return p.firstExpr(x.Body, m, name)
}

func (p *parser) firstOpt(x *ebnf.Option, m map[string]map[rune]bool, name string) bool {
	p.firstExpr(x.Body, m, name)
	return true
}

func (p *parser) firstRep(x *ebnf.Repetition, m map[string]map[rune]bool, name string) bool {
	p.firstExpr(x.Body, m, name)
	return true
}
