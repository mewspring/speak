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
	terms := &Terminals{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(&terms); err != nil {
		return nil, errors.WithStack(err)
	}
	sort.Sort(terms.Names)
	sort.Strings(terms.Tokens)
	return terms, nil
}

// DecodeFile decodes the terminals of a grammar, reading JSON input from path.
func DecodeFile(path string) (*Terminals, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	br := bufio.NewReader(f)
	return Decode(br)
}
