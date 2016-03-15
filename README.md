# Speak

[![Build Status](https://travis-ci.org/mewmew/speak.svg?branch=master)](https://travis-ci.org/mewmew/speak)
[![Coverage Status](https://coveralls.io/repos/github/mewmew/speak/badge.svg?branch=master)](https://coveralls.io/github/mewmew/speak?branch=master)
[![GoDoc](https://godoc.org/github.com/mewmew/speak?status.svg)](https://godoc.org/github.com/mewmew/speak)

Speak is an experimental compiler construction playground, written for the [µC compiler] as a learning experience. Inspired by [Gocc], Speak generates lexers and parses from language grammars expressed in EBNF with annotated production actions.

The augmented EBNF format is described in the package documentation of the [ebnf] package.

The name *Speak* is derived from the pronunciation of the Swedish word *spik*, in commemoration of my father Peter Eklind who worked with nail guns (or *spikpistoler* in Swedish) for the better part of his life.

[µC compiler]: https://github.com/mewmew/uc
[Gocc]: https://github.com/goccmack/gocc
[ebnf]: https://godoc.org/github.com/mewmew/speak/internal/golang.org/x/exp/ebnf

## Public domain

The source code and any original content of this repository is hereby released into the [public domain].

[public domain]: https://creativecommons.org/publicdomain/zero/1.0/

## License

Any code or documentation directly derived from the [standard Go source code](https://github.com/golang) is governed by a [BSD license](http://golang.org/LICENSE).
