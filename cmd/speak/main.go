// Speak parses input by runtime evaluation of language grammars expressed in
// EBNF.
package main

import (
	"bufio"
	"flag"
	"fmt"
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
	dbg = log.New(os.Stderr, term.MagentaBold("speak:")+" ", 0)
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
	flag.StringVar(&grammarPath, "g", "grammar.ebnf", "path to EBNF grammar")
	flag.StringVar(&start, "start", "", "start production rule")
	flag.Usage = usage
	flag.Parse()

	// Parse and validate grammar.
	grammar, first, err := parseGrammar(grammarPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if len(start) == 0 {
		start = first
	}
	dbg.Println("start:", start)
	if err := ebnf.Verify(grammar, start); err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
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
	ret := p.evalProd(p.grammar[start])
	dbg.Println("speak:")
	dbg.Printf("   ret: %v", ret)
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
}

func (p *parser) evalProd(x *ebnf.Production) bool {
	dbg.Println("evalProd:", exprString(x))
	dbg.Printf("   name: %s", x.Name.String)
	return p.evalExpr(x.Expr)
}

func (p *parser) evalExpr(x ebnf.Expression) bool {
	dbg.Println("evalExpr:", exprString(x))
	if p.eof {
		return true
	}
	switch x := x.(type) {
	case *ebnf.Production:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	case ebnf.Alternative:
		return p.evalAlt(x)
	case ebnf.Sequence:
		return p.evalSeq(x)
	case *ebnf.Name:
		return p.evalName(x)
	case *ebnf.Token:
		return p.evalToken(x)
	case *ebnf.Range:
		return p.evalRange(x)
	case *ebnf.Group:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	case *ebnf.Option:
		return p.evalOpt(x)
	case *ebnf.Repetition:
		return p.evalRep(x)
	default:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	}
}

func (p *parser) evalAlt(x ebnf.Alternative) bool {
	dbg.Println("evalAlt:", exprString(x))
	for _, e := range x {
		// record pos, and reset if alternative mismatches.
		bak := p.pos
		if p.evalExpr(e) {
			return true
		}
		// reset pos.
		p.pos = bak
	}
	return false
}

func (p *parser) evalSeq(x ebnf.Sequence) bool {
	dbg.Println("evalSeq:", exprString(x))
	for _, e := range x {
		if !p.evalExpr(e) {
			return false
		}
	}
	return true
}

func (p *parser) evalName(x *ebnf.Name) bool {
	dbg.Println("evalName:", exprString(x))
	prod := p.grammar[x.String]
	return p.evalProd(prod)
}

func (p *parser) evalToken(x *ebnf.Token) bool {
	dbg.Println("evalToken:", exprString(x))
	for _, q := range x.String {
		r := p.nextRune()
		if r == eof {
			return false
		}
		if r != q {
			return false
		}
	}
	return true
}

func (p *parser) evalRange(x *ebnf.Range) bool {
	dbg.Println("evalRange:", exprString(x))
	f, _ := utf8.DecodeRuneInString(x.Begin.String)
	t, _ := utf8.DecodeRuneInString(x.End.String)
	r := p.nextRune()
	if r == eof {
		dbg.Println("   eof")
		return false
	}
	ret := f <= r && r <= t
	dbg.Printf("   r: %c", r)
	dbg.Printf("   from: %c", f)
	dbg.Printf("   to: %c", t)
	dbg.Printf("   ret: %v", ret)
	return ret
}

func (p *parser) evalOpt(x *ebnf.Option) bool {
	dbg.Println("evalOpt:", exprString(x))
	if p.eof {
		return true
	}
	// store position and try to parse optional.
	bak := p.pos
	if !p.evalExpr(x.Body) {
		// reset position
		p.pos = bak
	}
	return true
}

func (p *parser) evalRep(x *ebnf.Repetition) bool {
	dbg.Println("evalRep:", exprString(x))
	for !p.eof && p.evalExpr(x.Body) {
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
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
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
		return eof
	}
	r, size := utf8.DecodeRune(p.input[p.pos:])
	p.pos += size
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
	var first string
	min := -1
	for name, prod := range grammar {
		r, _ := utf8.DecodeRuneInString(name)
		if unicode.IsUpper(r) {
			off := prod.Name.Pos().Offset
			if min == -1 || off < min {
				first = name
				min = off
			}
		}
	}
	if len(first) == 0 {
		return nil, "", errors.Errorf("unable to located first syntactic production rule (capital letter) in grammar %q", grammarPath)
	}
	return grammar, first, nil
}
