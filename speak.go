// Speak is an experimental compiler construction playground. Inspired by Gocc,
// Speak generates lexers and parses from language grammars expressed in EBNF.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/kr/pretty"
	"github.com/pkg/errors"
	"golang.org/x/exp/ebnf"
)

func usage() {
	const use = `
speak [OPTION]... FILE.ebnf

Flags:`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	var (
		// start represents the initial production rule of the grammar.
		start string
	)
	flag.StringVar(&start, "start", "Program", "initial production rule of the grammar")
	//flag.StringVar(&skip, "skip", "skip", "comma-separated list of terminals to ignore (e.g. whitespace, comments)")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	grammarPath := flag.Arg(0)
	if err := speak(grammarPath, start); err != nil {
		log.Fatal(err)
	}
}

// speak generates lexers and parsers for the given EBNF grammar.
func speak(grammarPath, start string) error {
	// Parse the grammar.
	f, err := os.Open(grammarPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()
	br := bufio.NewReader(f)
	grammar, err := ebnf.Parse(filepath.Base(grammarPath), br)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := validate(grammar, start); err != nil {
		return errors.WithStack(err)
	}
	fmt.Println("=== [ Grammar ] ===")
	pretty.Println(grammar)

	// Extract terminals from grammar.
	terms := extractTerms(grammar)
	fmt.Println("=== [ Terminals ] ===")
	pretty.Println(terms)

	// Return extracted terminals.
	var names lexemes
	var tokens lexemes
	for id, term := range terms.names {
		reg := regexpString(grammar, term)
		lex := &lexeme{
			id:  id,
			reg: reg,
		}
		names = append(names, lex)
	}
	for id, term := range terms.tokens {
		reg := regexpString(grammar, term)
		lex := &lexeme{
			id:  id,
			reg: reg,
		}
		tokens = append(tokens, lex)
	}
	sort.Sort(names)
	sort.Sort(tokens)

	fmt.Println("=== [ Regular expressions ] ===")
	pretty.Println("names:", names)
	pretty.Println("tokens:", tokens)
	return nil
}

// lexeme represents a lexeme of the grammar.
type lexeme struct {
	// Lexeme ID.
	id string
	// Regular expression of the lexeme.
	reg string
}

// lexemes implements sort.Sort, sorting based on lexeme ID.
type lexemes []*lexeme

func (ls lexemes) Len() int           { return len(ls) }
func (ls lexemes) Less(i, j int) bool { return ls[i].id < ls[j].id }
func (ls lexemes) Swap(i, j int)      { ls[i], ls[j] = ls[j], ls[i] }
