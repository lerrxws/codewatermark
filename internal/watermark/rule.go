package watermark

import "go/ast"

type Rule interface {
	// Match must not modify the statement.
	Match(statement ast.Stmt) bool

	// Encode must support both bits on matched statements and preserve other sites.
	Encode(statement ast.Stmt, bit byte) (ast.Stmt, bool)
}
