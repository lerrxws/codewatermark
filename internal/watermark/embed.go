package watermark

import (
	"errors"
	"fmt"
	"go/ast"
	"sort"
)

type encodingSite struct {
	statement *ast.Stmt
	rule      Rule
}

// Embed uses the first matching rule for each statement in a statement list.
func Embed(file *ast.File, bits string, rules []Rule) error {
	if err := validateInput(file, rules); err != nil {
		return err
	}

	sites := findSites(file, rules)
	if len(bits) > len(sites) {
		return fmt.Errorf("insufficient capacity: need %d bits, source has %d eligible sites", len(bits), len(sites))
	}

	return encodeBits(sites, bits)
}

func validateInput(file *ast.File, rules []Rule) error {
	if file == nil {
		return errors.New("source AST is nil")
	}

	if len(rules) == 0 {
		return errors.New("no encoding rules configured")
	}

	for i, rule := range rules {
		if rule == nil {
			return fmt.Errorf("encoding rule %d is nil", i+1)
		}
	}

	return nil
}

func findSites(file *ast.File, rules []Rule) []encodingSite {
	var sites []encodingSite
	claimed := make(map[ast.Stmt]bool)

	ast.Inspect(file, func(node ast.Node) bool {
		if statement, ok := node.(ast.Stmt); ok && claimed[statement] {
			return false
		}

		var statements []ast.Stmt
		switch body := node.(type) {
		case *ast.BlockStmt:
			statements = body.List
		case *ast.CaseClause:
			statements = body.Body
		case *ast.CommClause:
			statements = body.Body
		}

		for i := range statements {
			for _, rule := range rules {
				if rule.Match(statements[i]) {
					sites = append(sites, encodingSite{statement: &statements[i], rule: rule})
					claimed[statements[i]] = true
					break
				}
			}
		}

		return true
	})

	sort.SliceStable(sites, func(i, j int) bool {
		return (*sites[i].statement).Pos() < (*sites[j].statement).Pos()
	})

	return sites
}

func encodeBits(sites []encodingSite, bits string) error {
	for i := 0; i < len(bits); i++ {
		site := sites[i]
		replacement, ok := site.rule.Encode(*site.statement, bits[i])
		if !ok || replacement == nil {
			return fmt.Errorf("rule failed to encode bit at position %d", i+1)
		}

		*site.statement = replacement
	}

	return nil
}
