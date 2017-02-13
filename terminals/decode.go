package terminals

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sort"

	"github.com/pkg/errors"
)

// Decode decodes the terminals of a grammar, reading JSON input from r.
func Decode(r io.Reader) (*Terminals, error) {
	br := bufio.NewReader(r)
	dec := json.NewDecoder(br)
	terms := &Terminals{}
	if err := dec.Decode(&terms); err != nil {
		return nil, errors.WithStack(err)
	}
	sort.Sort(terms.Names)
	sort.Sort(terms.Tokens)
	sort.Sort(terms.Skip)
	return terms, nil
}

// DecodeFile decodes the terminals of a grammar, reading JSON input from path.
func DecodeFile(path string) (*Terminals, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	return Decode(f)
}
