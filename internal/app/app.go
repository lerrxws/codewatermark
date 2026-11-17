package app

import (
	"bytes"
	"fmt"
	"github.com/lerrxws/codewatermark/internal/watermark"
	"github.com/lerrxws/codewatermark/internal/watermark/rules"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
)

type Options struct {
	SourcePath string
	OutputPath string
	Bits       string
}

func Run(options Options) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, options.SourcePath, nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return fmt.Errorf("parse source: %w", err)
	}

	typeInfo := &types.Info{
		Scopes: make(map[ast.Node]*types.Scope),
	}
	typeChecker := types.Config{Importer: importer.Default()}
	if _, err := typeChecker.Check(file.Name.Name, fset, []*ast.File{file}, typeInfo); err != nil {
		return fmt.Errorf("type-check source: %w", err)
	}

	if err := ast.Print(fset, file); err != nil {
		return fmt.Errorf("print AST: %w", err)
	}

	rule := rules.BooleanComparison{Scopes: typeInfo.Scopes}
	matches := 0

	ast.Inspect(file, func(node ast.Node) bool {
		statement, ok := node.(ast.Stmt)
		if ok && rule.Match(statement) {
			matches++
			fmt.Printf("Matched %T at %s\n",
				statement,
				fset.Position(statement.Pos()),
			)
		}

		return true
	})

	fmt.Printf("Total matches: %d\n", matches)

	if err := watermark.Embed(file, options.Bits, buildRules(typeInfo.Scopes)); err != nil {
		return fmt.Errorf("embed watermark: %w", err)
	}

	var output bytes.Buffer
	if err := format.Node(&output, fset, file); err != nil {
		return fmt.Errorf("format output: %w", err)
	}

	if err := os.WriteFile(options.OutputPath, output.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func buildRules(scopes map[ast.Node]*types.Scope) []watermark.Rule {
	return []watermark.Rule{
		// rules.Increment{},
		// rules.Assignment{},
		// rules.Declaration{},
		rules.BooleanComparison{Scopes: scopes},
	}
}
