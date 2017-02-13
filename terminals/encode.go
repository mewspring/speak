package terminals

import (
	"encoding/json"
	"io"
	"sort"

	"github.com/pkg/errors"
)

// Encode encodes the terminals of a grammar, writing JSON output to w.
func Encode(w io.Writer, terms *Terminals, indent bool) error {
	enc := json.NewEncoder(w)
	if indent {
		enc.SetIndent("", "\t")
	}
	sort.Sort(terms.Names)
	sort.Sort(terms.Tokens)
	sort.Sort(terms.Skip)
	if err := enc.Encode(terms); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
