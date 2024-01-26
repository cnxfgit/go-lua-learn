package parser

import (
	"luago/compiler/ast"
	"luago/compiler/lexer"
)

func Parse(chunk, chunkName string) *ast.Block {
	l := lexer.NewLexer(chunk, chunkName)
	block := parseBlock(l)
	l.NextTokenOfKind(lexer.TOKEN_EOF)
	return block
}
