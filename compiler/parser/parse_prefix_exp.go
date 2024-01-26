package parser

import (
	"luago/compiler/ast"
	"luago/compiler/lexer"
)

func parsePrefixExp(l *lexer.Lexer) ast.Exp {
	var exp ast.Exp
	if l.LookAhead() == lexer.TOKEN_IDENTIFIER {
		line, name := l.NextIdentifier()
		exp = &ast.NameExp{Line: line, Name: name}
	} else {
		exp = parseParensExp(l)
	}
	return _finishPrefixExp(l, exp)
}

func _finishPrefixExp(l *lexer.Lexer, exp ast.Exp) ast.Exp {
	for {
		switch l.LookAhead() {
		case lexer.TOKEN_SEP_LBRACK:
			l.NextToken()                             // `[`
			keyExp := parseExp(l)                     // exp
			l.NextTokenOfKind(lexer.TOKEN_SEP_RBRACK) // `]`
			exp = &ast.TableAccessExp{
				LastLine:  l.Line(),
				PrefixExp: exp,
				KeyExp:    keyExp,
			}
		case lexer.TOKEN_SEP_DOT:
			l.NextToken()                    // `.`
			line, name := l.NextIdentifier() // Name
			keyExp := &ast.StringExp{Line: line, Str: name}
			exp = &ast.TableAccessExp{LastLine: line, PrefixExp: exp, KeyExp: keyExp}
		case lexer.TOKEN_SEP_COLON,
			lexer.TOKEN_SEP_LPAREN, lexer.TOKEN_SEP_LCURLY, lexer.TOKEN_STRING:
			exp = _finishFuncCallExp(l, exp) // [`:` Name] args
		default:
			return exp
		}
	}
	return exp
}

func parseParensExp(l *lexer.Lexer) ast.Exp {
	l.NextTokenOfKind(lexer.TOKEN_SEP_LPAREN)
	exp := parseExp(l)
	l.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN)

	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp, *ast.NameExp, *ast.TableAccessExp:
		return &ast.ParensExp{Exp: exp}
	}
	return exp
}

func _finishFuncCallExp(l *lexer.Lexer, prefixExp ast.Exp) *ast.FuncCallExp {
	nameExp := _parseNameExp(l) // [`:` Name]
	line := l.Line()            //
	args := _parseArgs(l)       // args
	lastLine := l.Line()        //
	return &ast.FuncCallExp{
		Line:      line,
		LastLine:  lastLine,
		PrefixExp: prefixExp,
		NameExp:   nameExp,
		Args:      args,
	}
}

func _parseNameExp(l *lexer.Lexer) *ast.StringExp {
	if l.LookAhead() == lexer.TOKEN_SEP_COLON {
		l.NextToken()
		line, name := l.NextIdentifier()
		return &ast.StringExp{Line: line, Str: name}
	}
	return nil
}

func _parseArgs(l *lexer.Lexer) (args []ast.Exp) {
	switch l.LookAhead() {
	case lexer.TOKEN_SEP_LPAREN:
		l.NextToken()
		if l.LookAhead() != lexer.TOKEN_SEP_RPAREN {
			args = parseExpList(l)
		}
		l.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN)
	case lexer.TOKEN_SEP_LCURLY:
		args = []ast.Exp{parseTableConstructorExp(l)}
	default:
		line, str := l.NextTokenOfKind(lexer.TOKEN_STRING)
		args = []ast.Exp{&ast.StringExp{Line: line, Str: str}}
	}
	return
}
