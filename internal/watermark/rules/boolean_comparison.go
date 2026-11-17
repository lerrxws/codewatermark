package rules

import (
	"go/ast"
	"go/token"
	"go/types"
)

type BooleanComparison struct {
	Scopes map[ast.Node]*types.Scope
}

func (rule BooleanComparison) Match(node ast.Stmt) bool {
	conditional, ok := node.(*ast.IfStmt)
	if !ok || conditional == nil || conditional.Cond == nil {
		return false
	}

	scope := rule.Scopes[conditional]
	if scope == nil {
		return false
	}

	_, object := scope.LookupParent("true", conditional.Cond.Pos())
	if object != types.Universe.Lookup("true") {
		return false
	}

	if _, ok := conditional.Cond.(*ast.Ident); ok {
		return true
	}

	comparison, ok := conditional.Cond.(*ast.BinaryExpr)
	if !ok || comparison.Op != token.EQL {
		return false
	}

	if _, ok := comparison.X.(*ast.Ident); !ok {
		return false
	}

	if right, ok := comparison.Y.(*ast.Ident); !(ok && right.Name == "true") {
		return false
	}

	return true
}

func (rule BooleanComparison) Encode(node ast.Stmt, bit byte) (ast.Stmt, bool) {
	if !rule.Match(node) {
		return nil, false
	}

	conditional := node.(*ast.IfStmt)

	switch bit {
	case '0':
		if _, ok := conditional.Cond.(*ast.Ident); ok {
			return conditional, true
		}

		comparison := conditional.Cond.(*ast.BinaryExpr)
		conditional.Cond = comparison.X

	case '1':
		if _, ok := conditional.Cond.(*ast.BinaryExpr); ok {
			return conditional, true
		}

		conditional.Cond = &ast.BinaryExpr{
			X:  conditional.Cond,
			Op: token.EQL,
			Y:  ast.NewIdent("true"),
		}

	default:
		return nil, false
	}

	return conditional, true
}
