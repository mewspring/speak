# terms

The `terms` command extracts regular expressions for terminals of a given input grammar, and outputs them as JSON.

## Installation

```
go get github.com/mewmew/speak/cmd/terms
```

## Usage

```
terms [OPTION]... FILE.ebnf

Flags:
  -indent
        indent JSON output
  -o string
        output path
  -start string
        initial production rule of the grammar (default "Program")
```

## Examples

Invoke `terms` to extract regular expressions for tokens of the [uC input grammar](uc.ebnf).

```
$ terms -indent uc.ebnf
```

Output:

```json
{
    "names": [
        {
            "id": "ident",
            "reg": "[A-Z_a-z][0-9A-Z_a-z]*"
        },
        {
            "id": "int_lit",
            "reg": "[0-9][0-9]*"
        }
    ],
    "tokens": [
        "!",
        "!=",
        "\u0026\u0026",
        "(",
        ")",
        "*",
        "+",
        ",",
        "-",
        "/",
        ";",
        "\u003c",
        "\u003c=",
        "=",
        "==",
        "\u003e",
        "\u003e=",
        "[",
        "]",
        "else",
        "if",
        "return",
        "while",
        "{",
        "}"
    ]
}
```

## Public domain

The source code and any original content of this repository is hereby released into the [public domain].

[public domain]: https://creativecommons.org/publicdomain/zero/1.0/
