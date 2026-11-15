package app

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"

	"github.com/lerrxws/codewatermark/internal/watermark"
	"github.com/lerrxws/codewatermark/internal/watermark/rules"
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

	if err := watermark.Embed(file, options.Bits, buildRules()); err != nil {
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

func buildRules() []watermark.Rule {
	return []watermark.Rule{
		rules.Increment{},
		rules.Assignment{},
		rules.Declaration{},
	}
}
