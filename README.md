# Code Watermark

A university project for deterministic watermarking of Go source code using
abstract syntax trees (`go/ast` and `go/parser`). The current MVP supports three
encoding techniques. The assignment's remaining techniques will be added later.

## Requirements

- Go 1.26.5 or newer.

## Usage

```sh
go run ./cmd/watermark -source examples/source/main.go -output examples/output/main.go -bits 001
```

- `-source`: path to the original Go source file.
- `-output`: path to the generated Go source file; its parent directory must exist.
- `-bits`: a required, nonempty sequence of `0` and `1` characters.

Each eligible source location stores one bit, even when that statement runs
multiple times in a loop. The example contains three eligible locations:
two variable declarations followed by the addition assignment.

## Assignment rule

| Bit | Assignment form |
| --- | --- |
| `0` | `x += y` |
| `1` | `x = x + y` |

The same mapping applies to every Go compound assignment operator:
`+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`, and `&^=`.
For example, `x -= y` encodes `0`, while `x = x - y` encodes `1`.

The rule recognizes both forms and requires a single, nonblank identifier on
the left. Indexed expressions, selectors, declarations, and unrelated
assignments are skipped. The input is expected to be a valid Go program;
parsing does not perform type checking. These operators are variants of one
encoding technique.

## Declaration rule

| Bit | Declaration form |
| --- | --- |
| `0` | `x := expression` |
| `1` | `var x = expression` |

Only single-variable declarations with one initializer and an inferred type
are supported. Explicit types, grouped declarations, and multiple names are
skipped. Package-level declarations are excluded.

## Increment rule

| Bit | Increment | Decrement |
| --- | --- | --- |
| `0` | `i++` | `i--` |
| `1` | `i += 1` | `i -= 1` |

Only plain, nonblank variable targets are supported. The operand must be the
integer literal `1`. Existing `i = i + 1` and `i = i - 1` forms are also
recognized and normalized. Increment has priority over Assignment, so these
locations retain the same rule when re-embedded.

All rules operate on statements directly contained in blocks, switch cases,
and select cases. Header initializers, loop post statements, and labeled
statements are excluded. This narrows the previous assignment rule's scope so
that either declaration form can be installed in every discovered location.

Bits are assigned to the first eligible locations in source order. Remaining
locations retain their current form. Invalid bits and insufficient capacity
produce an error before any AST changes. Reapplying bits works with either
existing encoding form.

Each matched statement reserves its entire subtree, even if no bit reaches
that site. Nested statements inside it are not separate sites. This keeps
encoding locations independent and excludes nested initializer functions
inside already matched declarations.

## Project structure

```text
codewatermark/
├── cmd/
│   └── watermark/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── app.go
│   └── watermark/
│       ├── embed.go
│       ├── rule.go
│       └── rules/
│           ├── assignment.go
│           ├── declaration.go
│           └── increment.go
├── examples/
│   ├── source/
│   │   └── main.go
│   └── output/
├── go.mod
└── README.md
```

- `cmd/watermark`: flags, input validation, user-facing messages, and exit status.
- `internal/app`: options and orchestration of parsing, rule construction,
  embedding, formatting, and file writing.
- `internal/watermark`: the rule interface, shared traversal, and capacity checks.
- `internal/watermark/rules`: concrete encoding rules.
- `examples/source`: a sample Go program for development.
- `examples/output`: a directory for generated Go source files.

The command passes `app.Options` to `app.Run`. The application constructs rules
in priority order: Increment, Assignment, and Declaration, using its private
`buildRules` function.
The engine uses its `watermark.Rule` interface without importing concrete rules.
Concrete rules use only standard Go AST packages. For each statement, the first
matching rule in the supplied list is selected. The embedding function
discovers all locations before applying any changes. `Encode` returns the
resulting statement, and the engine installs it in the original statement
list. This allows a rule to replace one AST statement type with another.

`Match` must not modify nodes. A matching rule must support both bit values and
preserve other discovered locations. If a rule fails during encoding, earlier
AST changes may remain in memory; the CLI saves output only after successful
embedding and formatting.

## Development status

The MVP embeds a supplied bit sequence using three rules and writes formatted
Go source. Seven further techniques and technique selection remain
to be implemented for the full assignment.

## Verification

```sh
go vet ./...
go build ./...
```

Compare the behavior of the original and generated examples:

```sh
go run ./examples/source
go run ./examples/output
```
