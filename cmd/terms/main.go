package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"golang.org/x/exp/ebnf"
)

func usage() {
	const use = `
terms [OPTION]... FILE.ebnf

Flags:`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line parameters.
	var (
		// Specifies whether to indent JSON output.
		indent bool
		// Initial production rule of the grammar.
		start string
		// Output path.
		output string
	)
	flag.BoolVar(&indent, "indent", false, "indent JSON output")
	flag.StringVar(&output, "o", "", "output path")
	flag.StringVar(&start, "start", "Program", "initial production rule of the grammar")
	//flag.StringVar(&skip, "skip", "skip", "comma-separated list of terminals to ignore (e.g. whitespace, comments)")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	grammarPath := flag.Arg(0)

	// Extract regular expressions for the terminators of the input grammar, and
	// output them as JSON.
	if err := outputTerms(grammarPath, start, output, indent); err != nil {
		log.Fatal(err)
	}
}

// outputTerms extract regular expressions for the terminators of the input
// grammar, and outputs them as JSON.
func outputTerms(grammarPath, start, output string, indent bool) error {
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

	// Extract terminals from grammar.
	terms := extractTerms(grammar)

	jsonTerms := &jsonTerms{}
	for id, term := range terms.names {
		reg := regexpString(grammar, term)
		lex := &Lexeme{
			ID:  id,
			Reg: reg,
		}
		jsonTerms.Names = append(jsonTerms.Names, lex)
	}
	for id := range terms.tokens {
		jsonTerms.Tokens = append(jsonTerms.Tokens, id)
	}
	sort.Sort(jsonTerms.Names)
	sort.Strings(jsonTerms.Tokens)

	// Print the JSON output to stdout or the path specified by the "-o" flag.
	w := os.Stdout
	if len(output) > 0 {
		f, err := os.Create(output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w = f
	}
	if err := writeJSON(w, jsonTerms, indent); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// writeJSON writes the terminals in JSON format to w.
func writeJSON(w io.Writer, jsonTerms *jsonTerms, indent bool) error {
	var src io.Reader
	if indent {
		buf, err := json.MarshalIndent(jsonTerms, "", "\t")
		if err != nil {
			return errors.WithStack(err)
		}
		buf = append(buf, '\n')
		src = bytes.NewReader(buf)
	} else {
		buf, err := json.Marshal(jsonTerms)
		if err != nil {
			return errors.WithStack(err)
		}
		buf = append(buf, '\n')
		src = bytes.NewReader(buf)
	}
	if _, err := io.Copy(w, src); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Output JSON encoded terminals.
type jsonTerms struct {
	Names  lexemes  `json:"names"`
	Tokens []string `json:"tokens"`
}

// Lexeme represents a lexeme of the grammar.
type Lexeme struct {
	// Lexeme ID.
	ID string `json:"id"`
	// Regular expression of the lexeme.
	Reg string `json:"reg"`
}

// lexemes implements sort.Sort, sorting based on lexeme ID.
type lexemes []*Lexeme

func (ls lexemes) Len() int           { return len(ls) }
func (ls lexemes) Less(i, j int) bool { return ls[i].ID < ls[j].ID }
func (ls lexemes) Swap(i, j int)      { ls[i], ls[j] = ls[j], ls[i] }
