// The terms command extracts regular expressions for terminals of a given input
// grammar, and outputs them as JSON.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mewmew/speak/terminals"
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
		// Comma-separated list of terminals to ignore (e.g. whitespace,
		// comments).
		skip commaSepList
	)
	flag.BoolVar(&indent, "indent", false, "indent JSON output")
	flag.StringVar(&output, "o", "", "output path")
	flag.StringVar(&start, "start", "Program", "initial production rule of the grammar")
	flag.Var(&skip, "skip", "comma-separated list of terminals to ignore (e.g. whitespace, comments)")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	grammarPath := flag.Arg(0)

	// Extract regular expressions for the terminators of the input grammar, and
	// output them as JSON.
	if err := outputTerms(grammarPath, start, output, indent, skip); err != nil {
		log.Fatal(err)
	}
}

// outputTerms extract regular expressions for the terminators of the input
// grammar, and outputs them as JSON.
func outputTerms(grammarPath, start, output string, indent bool, skip []string) error {
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
	if err := validate(grammar, start, skip); err != nil {
		return errors.WithStack(err)
	}

	// Extract terminals from grammar.
	terms := extractTerms(grammar)

	jsonTerms := &terminals.Terminals{}
	for id, term := range terms.names {
		reg, err := regexpString(grammar, term)
		if err != nil {
			return errors.WithStack(err)
		}
		lex := &terminals.Lexeme{
			ID:  id,
			Reg: reg,
		}
		jsonTerms.Names = append(jsonTerms.Names, lex)
	}
	for id, term := range terms.tokens {
		reg, err := regexpString(grammar, term)
		if err != nil {
			return errors.WithStack(err)
		}
		lex := &terminals.Lexeme{
			ID:  id,
			Reg: reg,
		}
		jsonTerms.Tokens = append(jsonTerms.Tokens, lex)
	}
	for _, id := range skip {
		prod := grammar[id]
		reg, err := regexpString(grammar, prod.Expr)
		if err != nil {
			return errors.WithStack(err)
		}
		lex := &terminals.Lexeme{
			ID:  id,
			Reg: reg,
		}
		jsonTerms.Skip = append(jsonTerms.Skip, lex)
	}
	sort.Sort(jsonTerms.Names)
	sort.Sort(jsonTerms.Tokens)
	sort.Sort(jsonTerms.Skip)

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
	if err := terminals.Encode(w, jsonTerms, indent); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// commaSepList implements the flag.Value interface for comma-separated list of
// strings.
type commaSepList []string

// String returns the comma-separated string representation of v.
func (v *commaSepList) String() string {
	return strings.Join(*v, ",")
}

// Set sets v to the list of strings represented by s.
func (v *commaSepList) Set(s string) error {
	ss := strings.Split(s, ",")
	*v = commaSepList(ss)
	return nil
}
