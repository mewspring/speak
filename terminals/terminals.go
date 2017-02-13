// Package terminals implements access to the regular expressions associated
// with the terminals of a grammar.
package terminals

// Terminals represents the terminals of a grammar.
type Terminals struct {
	// Terminal names.
	Names Lexemes `json:"names,emitempty"`
	// Terminal tokens.
	Tokens Lexemes `json:"tokens,emitempty"`
	// Ignored terminals.
	Skip Lexemes `json:"skip,emitempty"`
}

// Lexeme represents a lexeme of the grammar.
type Lexeme struct {
	// Lexeme ID.
	ID string `json:"id"`
	// Regular expression of the lexeme.
	Reg string `json:"reg"`
}

// Lexemes implements sort.Sort, sorting based on lexeme ID.
type Lexemes []*Lexeme

func (ls Lexemes) Len() int           { return len(ls) }
func (ls Lexemes) Less(i, j int) bool { return ls[i].ID < ls[j].ID }
func (ls Lexemes) Swap(i, j int)      { ls[i], ls[j] = ls[j], ls[i] }
