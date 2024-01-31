package parser

import (
	"luago/compiler/ast"
	"luago/compiler/lexer"
	"luago/number"
)

func parseExpList(l *lexer.Lexer) []ast.Exp {
	exps := make([]ast.Exp, 0, 4)
	exps = append(exps, parseExp(l))
	for l.LookAhead() == lexer.TOKEN_SEP_COMMA {
		l.NextToken()
		exps = append(exps, parseExp(l))
	}
	return exps
}

func parseExp(l *lexer.Lexer) ast.Exp {
	return parseExp12(l)
}

func parseExp12(l *lexer.Lexer) ast.Exp {
	exp := parseExp11(l)
	for l.LookAhead() == lexer.TOKEN_OP_OR {
		line, op, _ := l.NextToken()
		exp = &ast.BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp11(l),
		}
	}
	return exp
}

func parseExp11(l *lexer.Lexer) ast.Exp {
	exp := parseExp10(l)
	for l.LookAhead() == lexer.TOKEN_OP_AND {
		line, op, _ := l.NextToken()
		land := &ast.BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp10(l),
		}
		exp = optimizeLogicalAnd(land)
	}
	return exp
}

func parseExp10(l *lexer.Lexer) ast.Exp {
	exp := parseExp9(l)
	for {
		switch l.LookAhead() {
		case lexer.TOKEN_OP_LT, lexer.TOKEN_OP_GT, lexer.TOKEN_OP_NE,
			lexer.TOKEN_OP_LE, lexer.TOKEN_OP_GE, lexer.TOKEN_OP_EQ:
			line, op, _ := l.NextToken()
			exp = &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp9(l)}
		default:
			return exp
		}
	}
}

func parseExp9(l *lexer.Lexer) ast.Exp {
	exp := parseExp8(l)
	for l.LookAhead() == lexer.TOKEN_OP_BOR {
		line, op, _ := l.NextToken()
		bor := &ast.BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp8(l),
		}
		exp = optimizeBitwiseBinaryOp(bor)
	}
	return exp
}

func parseExp8(l *lexer.Lexer) ast.Exp {
	exp := parseExp7(l)
	for l.LookAhead() == lexer.TOKEN_OP_BXOR {
		line, op, _ := l.NextToken()
		bxor := &ast.BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp7(l),
		}
		exp = optimizeBitwiseBinaryOp(bxor)
	}
	return exp
}

func parseExp7(l *lexer.Lexer) ast.Exp {
	exp := parseExp6(l)
	for l.LookAhead() == lexer.TOKEN_OP_BAND {
		line, op, _ := l.NextToken()
		band := &ast.BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp6(l),
		}
		exp = optimizeBitwiseBinaryOp(band)
	}
	return exp
}

func parseExp6(l *lexer.Lexer) ast.Exp {
	exp := parseExp5(l)
	for {
		switch l.LookAhead() {
		case lexer.TOKEN_OP_SHL, lexer.TOKEN_OP_SHR:
			line, op, _ := l.NextToken()
			shx := &ast.BinopExp{
				Line: line,
				Op:   op,
				Exp1: exp,
				Exp2: parseExp5(l),
			}
			exp = optimizeBitwiseBinaryOp(shx)
		default:
			return exp
		}
	}
}

func parseExp5(l *lexer.Lexer) ast.Exp {
	exp := parseExp4(l)
	if l.LookAhead() != lexer.TOKEN_OP_CONCAT {
		return exp
	}

	line := 0
	exps := []ast.Exp{exp}
	for l.LookAhead() == lexer.TOKEN_OP_CONCAT {
		line, _, _ = l.NextToken()
		exps = append(exps, parseExp4(l))
	}
	return &ast.ConcatExp{Line: line, Exps: exps}
}

func parseExp4(l *lexer.Lexer) ast.Exp {
	exp := parseExp3(l)
	for {
		switch l.LookAhead() {
		case lexer.TOKEN_OP_ADD, lexer.TOKEN_OP_SUB:
			line, op, _ := l.NextToken()
			arith := &ast.BinopExp{
				Line: line,
				Op:   op,
				Exp1: exp,
				Exp2: parseExp3(l),
			}
			exp = optimizeArithBinaryOp(arith)
		default:
			return exp
		}
	}
}

func parseExp3(l *lexer.Lexer) ast.Exp {
	exp := parseExp2(l)
	for {
		switch l.LookAhead() {
		case lexer.TOKEN_OP_MUL, lexer.TOKEN_OP_MOD, lexer.TOKEN_OP_DIV,
			lexer.TOKEN_OP_IDIV:
			line, op, _ := l.NextToken()
			arith := &ast.BinopExp{
				Line: line, 
				Op:   op, 
				Exp1: exp,
			    Exp2: parseExp2(l)}
			exp = optimizeArithBinaryOp(arith)
		default:
			return exp
		}
	}
}

func parseExp2(l *lexer.Lexer) ast.Exp {
	switch l.LookAhead() {

	case lexer.TOKEN_OP_UNM, lexer.TOKEN_OP_BNOT, lexer.TOKEN_OP_LEN,
		lexer.TOKEN_OP_NOT:
		line, op, _ := l.NextToken()
		exp := &ast.UnopExp{Line: line, Op: op, Exp: parseExp2(l)}
		return optimizeUnaryOp(exp)
	}
	return parseExp1(l)
}

func parseExp1(l *lexer.Lexer) ast.Exp {
	exp := parseExp0(l)
	if l.LookAhead() == lexer.TOKEN_OP_POW {
		line, op, _ := l.NextToken()
		exp = &ast.BinopExp{
			Line: line,
			Op:   op,
			Exp1: exp,
			Exp2: parseExp2(l),
		}
	}
	return exp
}

func parseExp0(l *lexer.Lexer) ast.Exp {
	switch l.LookAhead() {
	case lexer.TOKEN_VARARG: // `...`
		line, _, _ := l.NextToken()
		return &ast.VarargExp{Line: line}
	case lexer.TOKEN_KW_NIL: // nil
		line, _, _ := l.NextToken()
		return &ast.NilExp{Line: line}
	case lexer.TOKEN_KW_TRUE: // true
		line, _, _ := l.NextToken()
		return &ast.TrueExp{Line: line}
	case lexer.TOKEN_KW_FALSE: // false
		line, _, _ := l.NextToken()
		return &ast.FalseExp{Line: line}
	case lexer.TOKEN_STRING: // LiteralString
		line, _, token := l.NextToken()
		return &ast.StringExp{Line: line, Str: token}
	case lexer.TOKEN_NUMBER: // Numeral
		return parseNumberExp(l)
	case lexer.TOKEN_SEP_LCURLY: // tableconstructor
		return parseTableConstructorExp(l)
	case lexer.TOKEN_KW_FUNCTION: // functiondef
		l.NextToken()
		return parseFuncDefExp(l)
	default: // prefixexp
		return parsePrefixExp(l)
	}
}

func parseNumberExp(l *lexer.Lexer) ast.Exp {
	line, _, token := l.NextToken()
	if i, ok := number.ParseInteger(token); ok {
		return &ast.IntegerExp{Line: line, Val: i}
	} else if f, ok := number.ParseFloat(token); ok {
		return &ast.FloatExp{Line: line, Val: f}
	} else {
		panic("not a number: " + token)
	}
}

func parseFuncDefExp(l *lexer.Lexer) *ast.FuncDefExp {
	line := l.Line()
	l.NextTokenOfKind(lexer.TOKEN_SEP_LPAREN)
	parList, isVararg := _parseParList(l)
	l.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN)
	block := parseBlock(l)
	lastLine, _ := l.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.FuncDefExp{
		Line:     line,
		LastLine: lastLine,
		ParList:  parList,
		IsVararg: isVararg,
		Block:    block,
	}
}

func _parseParList(l *lexer.Lexer) (names []string, isVararg bool) {
	switch l.LookAhead() {
	case lexer.TOKEN_SEP_RPAREN:
		return nil, false
	case lexer.TOKEN_VARARG:
		l.NextToken()
		return nil, true
	}
	_, name := l.NextIdentifier()
	names = append(names, name)
	for l.LookAhead() == lexer.TOKEN_SEP_COMMA {
		l.NextToken()
		if l.LookAhead() == lexer.TOKEN_IDENTIFIER {
			_, name := l.NextIdentifier()
			names = append(names, name)
		} else {
			l.NextTokenOfKind(lexer.TOKEN_VARARG)
			isVararg = true
			break
		}
	}
	return
}

func parseTableConstructorExp(l *lexer.Lexer) *ast.TableConstructorExp {
	line := l.Line()
	l.NextTokenOfKind(lexer.TOKEN_SEP_LCURLY)
	keyExps, valExps := _parseFieldList(l)
	l.NextTokenOfKind(lexer.TOKEN_SEP_RCURLY)
	lastLine := l.Line()
	return &ast.TableConstructorExp{
		Line:     line,
		LastLine: lastLine,
		KeyExps:  keyExps,
		ValExps:  valExps,
	}
}

func _parseFieldList(l *lexer.Lexer) (ks, vs []ast.Exp) {
	if l.LookAhead() != lexer.TOKEN_SEP_RCURLY {
		k, v := _parseField(l) // field
		ks = append(ks, k)
		vs = append(vs, v)               //
		for _isFieldSep(l.LookAhead()) { // {
			l.NextToken()                                // fieldsep
			if l.LookAhead() != lexer.TOKEN_SEP_RCURLY { //
				k, v := _parseField(l) // field
				ks = append(ks, k)
				vs = append(vs, v) //
			} else { // }
				break // [fieldsep]
			}
		}
	}
	return
}

func _isFieldSep(tokenKind int) bool {
	return tokenKind == lexer.TOKEN_SEP_COMMA || tokenKind == lexer.TOKEN_SEP_SEMI
}

func _parseField(l *lexer.Lexer) (k, v ast.Exp) {
	if l.LookAhead() == lexer.TOKEN_SEP_LBRACK {
		l.NextToken()
		k = parseExp(l)
		l.NextTokenOfKind(lexer.TOKEN_SEP_RBRACK)
		l.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN)
		v = parseExp(l)
		return
	}

	exp := parseExp(l)
	if nameExp, ok := exp.(*ast.NameExp); ok {
		if l.LookAhead() == lexer.TOKEN_OP_ASSIGN {
			l.NextToken()
			k = &ast.StringExp{Line: nameExp.Line, Str: nameExp.Name}
			v = parseExp(l)
			return
		}
	}

	return nil, exp
}
