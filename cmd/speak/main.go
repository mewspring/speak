// Speak parses input by runtime evaluation of language grammars expressed in
// EBNF.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kr/pretty"
	"github.com/pkg/errors"
	"golang.org/x/exp/ebnf"
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
		var input io.Reader
		if inputPath == "-" {
			input = os.Stdin
		} else {
			f, err := os.Open(inputPath)
			if err != nil {
				log.Fatalf("%+v", errors.WithStack(err))
			}
			defer f.Close()
			input = f
		}
		if err := speak(grammar, start, input); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

// speak parses the given input by runtime evaluation of the grammar from the
// start production rule.
func speak(grammar ebnf.Grammar, start string, input io.Reader) error {
	p := &parser{
		grammar: grammar,
		cur:     start,
	}
	p.parse(input)
	return nil
}

// parser holds the state of the EBNF grammar used for parsing.
type parser struct {
	// EBNF language grammar.
	grammar ebnf.Grammar
	// Current production rule.
	cur string
}

// parse parses the given input by runtime evaluation of the grammar from the
// current production rule.
func (p *parser) parse(r io.Reader) {
	pretty.Println("grammar:", p.grammar)
}

// ### [ Helper functions ] ####################################################

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
