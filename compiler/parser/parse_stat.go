package parser

import (
	"luago/compiler/ast"
	"luago/compiler/lexer"
)

func parseStat(l *lexer.Lexer) ast.Stat {
	switch l.LookAhead() {
	case lexer.TOKEN_SEP_SEMI:
		return parseEmptyStat(l)
	case lexer.TOKEN_KW_BREAK:
		return parseBreakStat(l)
	case lexer.TOKEN_SEP_LABEL:
		return parseLabelStat(l)
	case lexer.TOKEN_KW_GOTO:
		return parseGotoStat(l)
	case lexer.TOKEN_KW_DO:
		return parseDoStat(l)
	case lexer.TOKEN_KW_WHILE:
		return parseWhileStat(l)
	case lexer.TOKEN_KW_REPEAT:
		return parseRepeatStat(l)
	case lexer.TOKEN_KW_IF:
		return parseIfStat(l)
	case lexer.TOKEN_KW_FOR:
		return parseForStat(l)
	case lexer.TOKEN_KW_FUNCTION:
		return parseFuncDefStat(l)
	case lexer.TOKEN_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(l)
	default:
		return parseAssignOrFuncCallStat(l)
	}
}

func parseEmptyStat(l *lexer.Lexer) *ast.EmptyStat {
	l.NextTokenOfKind(lexer.TOKEN_SEP_SEMI)
	return &ast.EmptyStat{}
}

func parseBreakStat(l *lexer.Lexer) *ast.BreakStat {
	l.NextTokenOfKind(lexer.TOKEN_KW_BREAK)
	return &ast.BreakStat{Line: l.Line()}
}

func parseLabelStat(l *lexer.Lexer) *ast.LabelStat {
	l.NextTokenOfKind(lexer.TOKEN_SEP_LABEL)
	_, name := l.NextIdentifier()
	l.NextTokenOfKind(lexer.TOKEN_SEP_LABEL)
	return &ast.LabelStat{Name: name}
}

func parseGotoStat(l *lexer.Lexer) *ast.GotoStat {
	l.NextTokenOfKind(lexer.TOKEN_KW_GOTO)
	_, name := l.NextIdentifier()
	return &ast.GotoStat{Name: name}
}

func parseDoStat(l *lexer.Lexer) *ast.DoStat {
	l.NextTokenOfKind(lexer.TOKEN_KW_DO)
	block := parseBlock(l)
	l.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.DoStat{Block: block}
}

func parseWhileStat(l *lexer.Lexer) *ast.WhileStat {
	l.NextTokenOfKind(lexer.TOKEN_KW_WHILE)
	exp := parseExp(l)
	l.NextTokenOfKind(lexer.TOKEN_KW_DO)
	block := parseBlock(l)
	l.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.WhileStat{Exp: exp, Block: block}
}

func parseRepeatStat(l *lexer.Lexer) *ast.RepeatStat {
	l.NextTokenOfKind(lexer.TOKEN_KW_REPEAT)
	block := parseBlock(l)
	l.NextTokenOfKind(lexer.TOKEN_KW_UNTIL)
	exp := parseExp(l)
	return &ast.RepeatStat{Block: block, Exp: exp}
}

func parseIfStat(l *lexer.Lexer) *ast.IfStat {
	exps := make([]ast.Exp, 0, 4)
	blocks := make([]*ast.Block, 0, 4)

	l.NextTokenOfKind(lexer.TOKEN_KW_IF)
	exps = append(exps, parseExp(l))
	l.NextTokenOfKind(lexer.TOKEN_KW_THEN)
	blocks = append(blocks, parseBlock(l))

	for l.LookAhead() == lexer.TOKEN_KW_ELSEIF {
		l.NextToken()
		exps = append(exps, parseExp(l))
		l.NextTokenOfKind(lexer.TOKEN_KW_THEN)
		blocks = append(blocks, parseBlock(l))
	}

	if l.LookAhead() == lexer.TOKEN_KW_ELSE {
		l.NextToken()
		exps = append(exps, &ast.TrueExp{Line: l.Line()})
		blocks = append(blocks, parseBlock(l))
	}

	l.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.IfStat{Exps: exps, Blocks: blocks}
}

func parseForStat(l *lexer.Lexer) ast.Stat {
	lineOfFor, _ := l.NextTokenOfKind(lexer.TOKEN_KW_FOR)
	_, name := l.NextIdentifier()
	if l.LookAhead() == lexer.TOKEN_OP_ASSIGN {
		return _finishForNumStat(l, lineOfFor, name)
	} else {
		return _finishForInStat(l, name)
	}
}

func _finishForNumStat(l *lexer.Lexer, lineOfFor int, varName string) *ast.ForNumStat {
	l.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN)
	initExp := parseExp(l)
	l.NextTokenOfKind(lexer.TOKEN_SEP_COMMA)
	limitExp := parseExp(l)

	var stepExp ast.Exp
	if l.LookAhead() == lexer.TOKEN_SEP_COMMA {
		l.NextToken()
		stepExp = parseExp(l)
	} else {
		stepExp = &ast.IntegerExp{Line: l.Line(), Val: 1}
	}

	lineOfDo, _ := l.NextTokenOfKind(lexer.TOKEN_KW_DO)
	block := parseBlock(l)
	l.NextTokenOfKind(lexer.TOKEN_KW_END)

	return &ast.ForNumStat{
		LineOfFor: lineOfFor,
		LineOfDo:  lineOfDo,
		VarName:   varName,
		InitExp:   initExp,
		LimitExp:  limitExp,
		StepExp:   stepExp,
		Block:     block,
	}
}

func _finishForInStat(l *lexer.Lexer, name0 string) *ast.ForInStat {
	nameList := _finishNameList(l, name0)

	l.NextTokenOfKind(lexer.TOKEN_KW_IN)
	expList := parseExpList(l)
	lineOfDo, _ := l.NextTokenOfKind(lexer.TOKEN_KW_DO)
	block := parseBlock(l)
	l.NextTokenOfKind(lexer.TOKEN_KW_END)
	return &ast.ForInStat{
		LineOfDo: lineOfDo,
		NameList: nameList,
		ExpList:  expList,
		Block:    block,
	}
}

func _finishNameList(l *lexer.Lexer, name0 string) []string {
	names := []string{name0}
	for l.LookAhead() == lexer.TOKEN_SEP_COMMA {
		l.NextToken()
		_, name := l.NextIdentifier()
		names = append(names, name)
	}
	return names
}

func parseLocalAssignOrFuncDefStat(l *lexer.Lexer) ast.Stat {
	l.NextTokenOfKind(lexer.TOKEN_KW_LOCAL)
	if l.LookAhead() == lexer.TOKEN_KW_FUNCTION {
		return _finishLocalFuncDefStat(l)
	} else {
		return _finishLocalVarDeclStat(l)
	}
}

func _finishLocalFuncDefStat(l *lexer.Lexer) *ast.LocalFuncDefStat {
	l.NextTokenOfKind(lexer.TOKEN_KW_FUNCTION)
	_, name := l.NextIdentifier()
	fdExp := parseFuncDefExp(l)
	return &ast.LocalFuncDefStat{Name: name, Exp: fdExp}
}

func _finishLocalVarDeclStat(l *lexer.Lexer) *ast.LocalVarDeclStat {
	_, name0 := l.NextIdentifier()
	nameList := _finishNameList(l, name0)
	var expList []ast.Exp = nil
	if l.LookAhead() == lexer.TOKEN_OP_ASSIGN {
		l.NextToken()
		expList = parseExpList(l)
	}
	lastLine := l.Line()
	return &ast.LocalVarDeclStat{
		LastLine: lastLine,
		NameList: nameList,
		ExpList:  expList,
	}
}

func parseAssignOrFuncCallStat(l *lexer.Lexer) ast.Stat {
	prefixExp := parsePrefixExp(l)
	if fc, ok := prefixExp.(*ast.FuncCallExp); ok {
		return fc
	} else {
		return parseAssignStat(l, prefixExp)
	}
}

func parseAssignStat(l *lexer.Lexer, var0 ast.Exp) *ast.AssignStat {
	varList := _finishVarList(l, var0)
	l.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN)
	expList := parseExpList(l)
	lastLine := l.Line()
	return &ast.AssignStat{LastLine: lastLine, VarList: varList, ExpList: expList}
}

func _finishVarList(l *lexer.Lexer, var0 ast.Exp) []ast.Exp {
	vars := []ast.Exp{_checkVar(l, var0)}
	for l.LookAhead() == lexer.TOKEN_SEP_COMMA {
		l.NextToken()
		exp := parsePrefixExp(l)
		vars = append(vars, _checkVar(l, exp))
	}
	return vars
}

func _checkVar(l *lexer.Lexer, exp ast.Exp) ast.Exp {
	switch exp.(type) {
	case *ast.NameExp, *ast.TableAccessExp:
		return exp
	}

	l.NextTokenOfKind(-1)
	panic("unreachable!")
}

func parseFuncDefStat(l *lexer.Lexer) *ast.AssignStat {
	l.NextTokenOfKind(lexer.TOKEN_KW_FUNCTION)
	fnExp, hasColon := _parseFuncName(l)
	fdExp := parseFuncDefExp(l)
	if hasColon {
		fdExp.ParList = append(fdExp.ParList, "")
		copy(fdExp.ParList[1:], fdExp.ParList)
		fdExp.ParList[0] = "self"
	}

	return &ast.AssignStat{
		LastLine: fdExp.Line,
		VarList:  []ast.Exp{fnExp},
		ExpList:  []ast.Exp{fdExp},
	}
}

func _parseFuncName(l *lexer.Lexer) (exp ast.Exp, hasColon bool) {
	line, name := l.NextIdentifier()
	exp = &ast.NameExp{Line: line, Name: name}

	for l.LookAhead() == lexer.TOKEN_SEP_DOT {
		l.NextToken()
		line, name := l.NextIdentifier()
		idx := &ast.StringExp{Line: line, Str: name}
		exp = &ast.TableAccessExp{LastLine: line, PrefixExp: exp, KeyExp: idx}
	}

	if l.LookAhead() == lexer.TOKEN_SEP_COLON {
		l.NextToken()
		line, name := l.NextIdentifier()
		idx := &ast.StringExp{Line: line, Str: name}
		exp = &ast.TableAccessExp{LastLine: line, PrefixExp: exp, KeyExp: idx}
		hasColon = true
	}

	return
}
