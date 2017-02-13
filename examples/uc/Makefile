all: lexer

# uc.ebnf -> uc.json
%.json: %.ebnf
	terms -indent -start File -skip "whitespace,comment" -o $@ $<

# uc.json -> lexer and token packages
lexer: uc.json
	genlex $<

clean:
	rm -rf token lexer
	rm -f uc.json

.PHONY: all clean
