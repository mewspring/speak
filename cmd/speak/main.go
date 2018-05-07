package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mewkiz/pkg/ioutilx"
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
	)
	flag.StringVar(&grammarPath, "g", "grammar.ebnf", "path to EBNF grammar")
	flag.Usage = usage
	flag.Parse()

	// Parse grammar.
	grammar, err := parseGrammar(grammarPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	for _, path := range flag.Args() {
		input, err := ioutilx.ReadFile(path)
		if err != nil {
			log.Fatalf("%+v", err)
		}
		if err := speak(grammar, input); err != nil {
			log.Fatalf("%+v", err)
		}
	}
}

func parseGrammar(grammarPath string) (ebnf.Grammar, error) {
	g, err := ebnf.Parse(grammarPath, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return g, nil
}

func speak(grammar ebnf.Grammar, input []byte) error {
	return nil
}
