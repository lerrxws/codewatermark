package rules

import (
	"go/ast"
	"go/token"
)

// Declaration encodes x := expression as '0' and var x = expression as '1'.
type Declaration struct{}

func (Declaration) Match(statement ast.Stmt) bool {
	_, _, ok := matchDeclaration(statement)
	return ok
}

func (Declaration) Encode(statement ast.Stmt, bit byte) (ast.Stmt, bool) {
	name, value, ok := matchDeclaration(statement)
	if !ok {
		return nil, false
	}

	switch bit {
	case '0':
		if _, ok := statement.(*ast.AssignStmt); ok {
			return statement, true
		}

		return &ast.AssignStmt{
			Lhs: []ast.Expr{name},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{value},
		}, true

	case '1':
		if _, ok := statement.(*ast.DeclStmt); ok {
			return statement, true
		}

		return &ast.DeclStmt{
			Decl: &ast.GenDecl{
				TokPos: statement.Pos(),
				Tok:    token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{Names: []*ast.Ident{name}, Values: []ast.Expr{value}},
				},
			},
		}, true
	}

	return nil, false
}

func matchDeclaration(statement ast.Stmt) (*ast.Ident, ast.Expr, bool) {
	switch declaration := statement.(type) {
	case *ast.AssignStmt:
		if declaration == nil || declaration.Tok != token.DEFINE {
			return nil, nil, false
		}

		if len(declaration.Lhs) != 1 || len(declaration.Rhs) != 1 {
			return nil, nil, false
		}

		name, ok := declaration.Lhs[0].(*ast.Ident)
		if !ok || name.Name == "_" {
			return nil, nil, false
		}

		return name, declaration.Rhs[0], true

	case *ast.DeclStmt:
		if declaration == nil {
			return nil, nil, false
		}

		group, ok := declaration.Decl.(*ast.GenDecl)
		if !ok || group.Tok != token.VAR {
			return nil, nil, false
		}

		if group.Lparen.IsValid() || len(group.Specs) != 1 {
			return nil, nil, false
		}

		spec, ok := group.Specs[0].(*ast.ValueSpec)
		if !ok || spec.Type != nil {
			return nil, nil, false
		}

		if len(spec.Names) != 1 || spec.Names[0].Name == "_" || len(spec.Values) != 1 {
			return nil, nil, false
		}

		return spec.Names[0], spec.Values[0], true
	}

	return nil, nil, false
}
