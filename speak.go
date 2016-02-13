// speak is an experimental compiler construction playground, written for the µC
// compiler as a learning experience. Inspired by Gocc, speak generates lexers
// and parses from language grammars expressed in EBNF with annotate production
// actions.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kr/pretty"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewmew/speak/internal/ebnf"
)

func usage() {
	const use = `
speak grammar.ebnf

Flags:`
	fmt.Fprintln(os.Stderr, use[1:])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	for _, grammarPath := range flag.Args() {
		if err := speak(grammarPath); err != nil {
			log.Fatal(err)
		}
	}
}

func speak(grammarPath string) error {
	// Parse the grammar.
	f, err := os.Open(grammarPath)
	if err != nil {
		return errutil.Err(err)
	}
	defer f.Close()
	grammar, err := ebnf.Parse(filepath.Base(grammarPath), f)
	if err != nil {
		return errutil.Err(err)
	}
	if err = ebnf.Verify(grammar, "Program"); err != nil {
		return errutil.Err(err)
	}

	pretty.Print(grammar)

	return nil
}