package rules

import (
	"go/ast"
	"go/token"
)

type Assignment struct{}

var assignmentOperators = map[token.Token]token.Token{
	token.ADD_ASSIGN:     token.ADD,
	token.SUB_ASSIGN:     token.SUB,
	token.MUL_ASSIGN:     token.MUL,
	token.QUO_ASSIGN:     token.QUO,
	token.REM_ASSIGN:     token.REM,
	token.AND_ASSIGN:     token.AND,
	token.OR_ASSIGN:      token.OR,
	token.XOR_ASSIGN:     token.XOR,
	token.SHL_ASSIGN:     token.SHL,
	token.SHR_ASSIGN:     token.SHR,
	token.AND_NOT_ASSIGN: token.AND_NOT,
}

func (Assignment) Match(node ast.Stmt) bool {
	_, _, _, ok := matchAssignment(node)
	return ok
}

func (Assignment) Encode(node ast.Stmt, bit byte) (ast.Stmt, bool) {
	assignment, operator, operand, ok := matchAssignment(node)
	if !ok {
		return nil, false
	}

	switch bit {
	case '0':
		if assignment.Tok != token.ASSIGN {
			return assignment, true
		}

		compound, _ := compoundOperator(operator)
		assignment.Tok = compound
		assignment.Rhs[0] = operand

	case '1':
		if assignment.Tok == token.ASSIGN {
			return assignment, true
		}

		lhs := assignment.Lhs[0].(*ast.Ident)
		assignment.Tok = token.ASSIGN
		assignment.Rhs[0] = &ast.BinaryExpr{
			X:  ast.NewIdent(lhs.Name),
			Op: operator,
			Y:  operand,
		}

	default:
		return nil, false
	}

	return assignment, true
}

func matchAssignment(node ast.Stmt) (*ast.AssignStmt, token.Token, ast.Expr, bool) {
	assignment, ok := node.(*ast.AssignStmt)
	if !ok || assignment == nil {
		return nil, token.ILLEGAL, nil, false
	}

	if len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		return nil, token.ILLEGAL, nil, false
	}

	lhs, ok := assignment.Lhs[0].(*ast.Ident)
	if !ok || lhs.Name == "_" {
		return nil, token.ILLEGAL, nil, false
	}

	if operator, supported := assignmentOperators[assignment.Tok]; supported {
		return assignment, operator, assignment.Rhs[0], true
	}

	if assignment.Tok == token.ASSIGN {
		expression, ok := assignment.Rhs[0].(*ast.BinaryExpr)
		if !ok {
			return nil, token.ILLEGAL, nil, false
		}

		if _, supported := compoundOperator(expression.Op); !supported {
			return nil, token.ILLEGAL, nil, false
		}

		first, ok := expression.X.(*ast.Ident)
		if ok && first.Name == lhs.Name {
			return assignment, expression.Op, expression.Y, true
		}
	}

	return nil, token.ILLEGAL, nil, false
}

func compoundOperator(binaryOperator token.Token) (token.Token, bool) {
	for compound, binary := range assignmentOperators {
		if binary == binaryOperator {
			return compound, true
		}
	}

	return token.ILLEGAL, false
}
