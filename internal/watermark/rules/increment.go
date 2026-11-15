package rules

import (
	"go/ast"
	"go/token"
)

// Increment encodes ++/-- as '0' and += 1/-= 1 as '1'.
type Increment struct{}

func (Increment) Match(statement ast.Stmt) bool {
	_, _, ok := matchIncrement(statement)
	return ok
}

func (Increment) Encode(statement ast.Stmt, bit byte) (ast.Stmt, bool) {
	name, operator, ok := matchIncrement(statement)
	if !ok {
		return nil, false
	}

	switch bit {
	case '0':
		if _, ok := statement.(*ast.IncDecStmt); ok {
			return statement, true
		}

		increment := token.INC
		if operator == token.SUB {
			increment = token.DEC
		}

		return &ast.IncDecStmt{X: name, TokPos: name.End(), Tok: increment}, true

	case '1':
		compound := token.ADD_ASSIGN
		if operator == token.SUB {
			compound = token.SUB_ASSIGN
		}

		if assignment, ok := statement.(*ast.AssignStmt); ok && assignment.Tok == compound {
			return statement, true
		}

		return &ast.AssignStmt{
			Lhs:    []ast.Expr{name},
			TokPos: name.End(),
			Tok:    compound,
			Rhs:    []ast.Expr{&ast.BasicLit{Kind: token.INT, ValuePos: statement.End(), Value: "1"}},
		}, true
	}

	return nil, false
}

func matchIncrement(statement ast.Stmt) (*ast.Ident, token.Token, bool) {
	if increment, ok := statement.(*ast.IncDecStmt); ok {
		if increment == nil {
			return nil, token.ILLEGAL, false
		}

		name, ok := increment.X.(*ast.Ident)
		if !ok || name.Name == "_" {
			return nil, token.ILLEGAL, false
		}

		switch increment.Tok {
		case token.INC:
			return name, token.ADD, true
		case token.DEC:
			return name, token.SUB, true
		}

		return nil, token.ILLEGAL, false
	}

	assignment, operator, operand, ok := matchAssignment(statement)
	if !ok || (operator != token.ADD && operator != token.SUB) {
		return nil, token.ILLEGAL, false
	}

	one, ok := operand.(*ast.BasicLit)
	if !ok || one.Kind != token.INT || one.Value != "1" {
		return nil, token.ILLEGAL, false
	}

	return assignment.Lhs[0].(*ast.Ident), operator, true
}
