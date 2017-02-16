all: lexer

# uc.ebnf -> uc.json
%.json: %.ebnf
	terms -indent -start File -skip "whitespace,comment" -o $@ $<

# uc.json -> lexer and token packages
lexer: uc.json
	genlex $<

clean:
	rm -f token/token.go lexer/lexer.go
	-rmdir --ignore-fail-on-non-empty token
	-rmdir --ignore-fail-on-non-empty lexer
	rm -f uc.json

.PHONY: all clean
