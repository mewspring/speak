// Speak parses input by runtime evaluation of language grammars expressed in
// EBNF.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
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
	flag.StringVar(&start, "start", "Start", "start production rule")
	flag.Usage = usage
	flag.Parse()

	// Parse grammar.
	grammar, err := parseGrammar(grammarPath, start)
	if err != nil {
		log.Fatalf("%+v", err)
	}
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
	// TODO: Remove?
	// Reached EOF.
	eof bool
}

func (p *parser) evalProd(x *ebnf.Production) bool {
	dbg.Print("evalProd:")
	dbg.Printf("   name: %s", x.Name.String)
	return p.evalExpr(x.Expr)
}

func (p *parser) evalExpr(x ebnf.Expression) bool {
	if p.eof {
		return true
	}
	dbg.Print("evalExpr:")
	switch x := x.(type) {
	case *ebnf.Production:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	case ebnf.Alternative:
		return p.evalAlt(x)
	case *ebnf.Name:
		return p.evalName(x)
	case *ebnf.Range:
		return p.evalRange(x)
	case *ebnf.Group:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	case *ebnf.Option:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	case *ebnf.Repetition:
		return p.evalRep(x)
	case ebnf.Sequence:
		return p.evalSeq(x)
	case *ebnf.Token:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	default:
		panic(fmt.Errorf("support for expression %T not yet implemented", x))
	}
}

func (p *parser) evalAlt(alt ebnf.Alternative) bool {
	for _, x := range alt {
		// record pos, and reset if alternative mismatches.
		bak := p.pos
		if p.evalExpr(x) {
			return true
		}
		// reset pos.
		p.pos = bak
	}
	return false
}

func (p *parser) evalSeq(seq ebnf.Sequence) bool {
	for _, x := range seq {
		if !p.evalExpr(x) {
			return false
		}
	}
	return true
}

func (p *parser) evalRep(x *ebnf.Repetition) bool {
	dbg.Print("evalRep:")
	if p.eof {
		return true
	}
	for p.evalExpr(x.Body) {
		if p.eof {
			return true
		}
	}
	return true
}

func (p *parser) evalName(x *ebnf.Name) bool {
	prod := p.grammar[x.String]
	return p.evalProd(prod)
}

func (p *parser) evalRange(x *ebnf.Range) bool {
	f, _ := utf8.DecodeRuneInString(x.Begin.String)
	t, _ := utf8.DecodeRuneInString(x.End.String)
	r := p.nextRune()
	dbg.Println("evalRange:")
	if r == eof {
		p.eof = true
		dbg.Println("   eof")
		return true
	}
	ret := f <= r && r <= t
	dbg.Printf("   r: %c", r)
	dbg.Printf("   from: %c", f)
	dbg.Printf("   to: %c", t)
	dbg.Printf("   ret: %v", ret)
	return ret
}

// ### [ Helper functions ] ####################################################

// eof signals end of input.
const eof rune = -1

// nextRune returns the next Unicode rune of the input source.
func (p *parser) nextRune() rune {
	if p.pos >= len(p.input) {
		return eof
	}
	r, size := utf8.DecodeRune(p.input[p.pos:])
	p.pos += size
	return r
}

// parseGrammar parses the given EBNF grammar and verifies it from the start
// production rule.
func parseGrammar(grammarPath, start string) (ebnf.Grammar, error) {
	f, err := os.Open(grammarPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	br := bufio.NewReader(f)
	grammar, err := ebnf.Parse(grammarPath, br)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := ebnf.Verify(grammar, start); err != nil {
		return nil, errors.WithStack(err)
	}
	return grammar, nil
}
