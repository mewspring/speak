package lexer_test

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/mewmew/speak/examples/uc/lexer"
	"github.com/pkg/errors"
)

func BenchmarkLexerScan(b *testing.B) {
	buf, err := ioutil.ReadFile("../testdata/input.c")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.SetBytes(int64(len(buf)))
	for i := 0; i < b.N; i++ {
		l := lexer.NewFromBytes(buf)
		for {
			_, err := l.Scan()
			if err != nil {
				if errors.Cause(err) == io.EOF {
					break
				}
				b.Fatalf("lexer error; %v", err)
			}
		}
	}
}
