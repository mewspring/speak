// Speak is an experimental compiler construction playground. Inspired by Gocc,
// Speak generates lexers and parses from language grammars expressed in EBNF.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kr/pretty"
	"github.com/pkg/errors"
	"golang.org/x/exp/ebnf"
)

func usage() {
	const use = `
speak FILE.ebnf

Flags:`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	grammarPath := flag.Arg(0)
	if err := speak(grammarPath); err != nil {
		log.Fatal(err)
	}
}

// speak generates a lexer and parser for the given EBNF grammar.
func speak(grammarPath string) error {
	// Parse the grammar.
	f, err := os.Open(grammarPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	grammar, err := ebnf.Parse(filepath.Base(grammarPath), f)
	if err != nil {
		return errors.WithStack(err)
	}
	if err = ebnf.Verify(grammar, "Program"); err != nil {
		return errors.WithStack(err)
	}

	pretty.Println(grammar)

	terms := Terminals(grammar)

	_ = pretty.Print

	//fmt.Println("=== [ Grammar ] ===")
	//pretty.Println(grammar)

	//fmt.Println("=== [ Terminals ] ===")
	//pretty.Println(terms)

	fmt.Println("=== [ Regular expressions ] ===")
	for _, term := range terms {
		//pretty.Println(term)
		fmt.Println("term:", RegexpString(grammar, term))
	}

	return nil
}
