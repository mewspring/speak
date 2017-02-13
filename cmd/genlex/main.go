// The genlex command generates lexers from JSON input containing regular
// expressions for terminals of a given input grammar.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/mewmew/speak/terminals"
	"github.com/pkg/errors"
)

func usage() {
	const use = `
genlex [OPTION]... FILE.json

Flags:`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	// Parse command line arguments.
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	jsonPath := flag.Arg(0)

	// Parse regular expressions for terminators from JSON input.
	ids, regs, err := parseJSON(jsonPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create a regular expression for identifying the different token
	// alternatives.
	reg, err := createRegexp(regs)
	if err != nil {
		log.Fatal(err)
	}

	// Generate lexer based on the regular expression identifying for
	// terminators.
	if err := genLexer(ids, reg); err != nil {
		log.Fatal(err)
	}
}

// genLexer generates a lexer based on the regular expression for identifying
// terminators of the input grammar.
func genLexer(ids []string, reg string) error {
	// Parse templates.
	t, err := template.ParseFiles("token.go.tmpl", "lexer.go.tmpl")
	if err != nil {
		return errors.WithStack(err)
	}
	// TODO: Figure out how to embed token.go.tmpl and lexer.go.tmpl into the
	// compiled binary of genlex.

	// Generate token/token.go.
	t1 := t.Lookup("token.go.tmpl")
	if err := os.MkdirAll("token", 0755); err != nil {
		return errors.WithStack(err)
	}
	f1, err := os.Create("token/token.go")
	if err != nil {
		return errors.WithStack(err)
	}
	defer f1.Close()
	if err := t1.Execute(f1, ids); err != nil {
		return errors.WithStack(err)
	}
	// Generate lexer/lexer.go.
	t2 := t.Lookup("lexer.go.tmpl")
	if err := os.MkdirAll("lexer", 0755); err != nil {
		return errors.WithStack(err)
	}
	f2, err := os.Create("lexer/lexer.go")
	if err != nil {
		return errors.WithStack(err)
	}
	defer f2.Close()
	if err := t2.Execute(f2, reg); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// createRegexp creates a regular expression for identifying the different
// token alternatives.
func createRegexp(regs []string) (string, error) {
	tokenAlts := "(" + strings.Join(regs, ")|(") + ")"
	regstr := "^(" + tokenAlts + ")"
	// Verify that the regular expression compiles successfully.
	if _, err := regexp.Compile(regstr); err != nil {
		return "", errors.WithStack(err)
	}
	return regstr, nil
}

// parseJSON parses and returns the regular expressions for terminators and
// their associated IDs, based on the given JSON input.
func parseJSON(jsonPath string) (ids, regs []string, err error) {
	terms, err := terminals.DecodeFile(jsonPath)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	for _, term := range terms.Names {
		id := fmt.Sprintf("name(%d, `%s`)", len(ids), term.ID)
		ids = append(ids, id)
		regs = append(regs, term.Reg)
	}
	for _, term := range terms.Tokens {
		id := fmt.Sprintf("token(%d, `%s`)", len(ids), term.ID)
		ids = append(ids, id)
		regs = append(regs, term.Reg)
	}
	for _, term := range terms.Skip {
		id := fmt.Sprintf("skip(%d, `%s`)", len(ids), term.ID)
		ids = append(ids, id)
		regs = append(regs, term.Reg)
	}
	return ids, regs, nil
}
