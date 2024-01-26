package parser

import (
	"luago/compiler/ast"
	"luago/compiler/lexer"
)

func parseBlock(l *lexer.Lexer) *ast.Block {
	return &ast.Block{
		Stats:    parseStats(l),
		RetExps:  parseRetExps(l),
		LastLine: l.Line(),
	}
}

func parseStats(l *lexer.Lexer) []ast.Stat {
	stats := make([]ast.Stat, 0, 8)
	for !_isReturnOrBlockEnd(l.LookAhead()) {
		stat := parseStat(l)
		if _, ok := stat.(*ast.EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

func _isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case lexer.TOKEN_KW_RETURN, lexer.TOKEN_EOF, lexer.TOKEN_KW_END,
		lexer.TOKEN_KW_ELSE, lexer.TOKEN_KW_ELSEIF, lexer.TOKEN_KW_UNTIL:
		return true
	}
	return false
}

func parseRetExps(l *lexer.Lexer) []ast.Exp {
	if l.LookAhead() != lexer.TOKEN_KW_RETURN {
		return nil
	}

	l.NextToken()
	switch l.LookAhead() {
	case lexer.TOKEN_EOF, lexer.TOKEN_KW_END,
		lexer.TOKEN_KW_ELSE, lexer.TOKEN_KW_ELSEIF, lexer.TOKEN_KW_UNTIL:
		return []ast.Exp{}
	case lexer.TOKEN_SEP_SEMI:
		l.NextToken()
		return []ast.Exp{}
	default:
		exps := parseExpList(l)
		if l.LookAhead() == lexer.TOKEN_SEP_SEMI {
			l.NextToken()
		}
		return exps
	}
}
