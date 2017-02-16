// The genlex command generates lexers from JSON input containing regular
// expressions for terminals of a given input grammar.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/mewkiz/pkg/goutil"
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
	tokenData, regs, err := parseJSON(jsonPath)
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
	if err := genLexer(tokenData, reg); err != nil {
		log.Fatal(err)
	}
}

// genLexer generates a lexer based on the regular expression for identifying
// terminators of the input grammar.
func genLexer(tokenData map[string]interface{}, reg string) error {
	// Parse templates.
	dir, err := goutil.SrcDir("github.com/mewmew/speak/cmd/genlex")
	if err != nil {
		return errors.WithStack(err)
	}
	tokenTmplPath := filepath.Join(dir, "token.go.tmpl")
	lexerTmplPath := filepath.Join(dir, "lexer.go.tmpl")
	t, err := template.ParseFiles(tokenTmplPath, lexerTmplPath)
	if err != nil {
		return errors.WithStack(err)
	}

	// Generate token/token.go.
	log.Println(`Creating "token/token.go"`)
	t1 := t.Lookup("token.go.tmpl")
	if err := os.MkdirAll("token", 0755); err != nil {
		return errors.WithStack(err)
	}
	f1, err := os.Create("token/token.go")
	if err != nil {
		return errors.WithStack(err)
	}
	defer f1.Close()
	if err := t1.Execute(f1, tokenData); err != nil {
		return errors.WithStack(err)
	}

	// Locate import path of the token package.
	tokenImportPath, err := goutil.RelImpPath("token")
	if err != nil {
		return errors.WithStack(err)
	}

	// Generate lexer/lexer.go.
	log.Println(`Creating "lexer/lexer.go"`)
	t2 := t.Lookup("lexer.go.tmpl")
	if err := os.MkdirAll("lexer", 0755); err != nil {
		return errors.WithStack(err)
	}
	f2, err := os.Create("lexer/lexer.go")
	if err != nil {
		return errors.WithStack(err)
	}
	defer f2.Close()
	lexerData := map[string]string{
		"ImportPath": tokenImportPath,
		"Regexp":     reg,
	}
	if err := t2.Execute(f2, lexerData); err != nil {
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
func parseJSON(jsonPath string) (tokenData map[string]interface{}, regs []string, err error) {
	terms, err := terminals.DecodeFile(jsonPath)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	var ids []string
	tokenData = make(map[string]interface{})
	minName := -1
	maxName := -1
	minToken := -1
	maxToken := -1
	minSkip := -1
	maxSkip := -1
	if len(terms.Names) > 0 {
		minName = len(ids)
	}
	for _, term := range terms.Names {
		id := fmt.Sprintf("name(%d, `%s`)", len(ids), term.ID)
		ids = append(ids, id)
		regs = append(regs, term.Reg)
	}
	if len(terms.Names) > 0 {
		maxName = len(ids) - 1
	}
	if len(terms.Tokens) > 0 {
		minToken = len(ids)
	}
	for _, term := range terms.Tokens {
		id := fmt.Sprintf("token(%d, `%s`)", len(ids), term.ID)
		ids = append(ids, id)
		regs = append(regs, term.Reg)
	}
	if len(terms.Tokens) > 0 {
		maxToken = len(ids) - 1
	}
	if len(terms.Skip) > 0 {
		minSkip = len(ids)
	}
	for _, term := range terms.Skip {
		id := fmt.Sprintf("skip(%d, `%s`)", len(ids), term.ID)
		ids = append(ids, id)
		regs = append(regs, term.Reg)
	}
	if len(terms.Skip) > 0 {
		maxSkip = len(ids) - 1
	}
	tokenData["MinName"] = minName
	tokenData["MaxName"] = maxName
	tokenData["MinToken"] = minToken
	tokenData["MaxToken"] = maxToken
	tokenData["MinSkip"] = minSkip
	tokenData["MaxSkip"] = maxSkip
	tokenData["IDs"] = ids
	return tokenData, regs, nil
}
