package codegen

import "luago/compiler/ast"

func isVarargOrFuncCall(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp:
		return true
	}
	return false
}

func removeTailNils(exps []ast.Exp) []ast.Exp {
	for n := len(exps) - 1; n >= 0; n-- {
		if _, ok := exps[n].(*ast.NilExp); !ok {
			return exps[0 : n+1]
		}
	}
	return nil
}

func lineOf(exp ast.Exp) int {
	switch x := exp.(type) {
	case *ast.NilExp:
		return x.Line
	case *ast.TrueExp:
		return x.Line
	case *ast.FalseExp:
		return x.Line
	case *ast.IntegerExp:
		return x.Line
	case *ast.FloatExp:
		return x.Line
	case *ast.StringExp:
		return x.Line
	case *ast.VarargExp:
		return x.Line
	case *ast.NameExp:
		return x.Line
	case *ast.FuncDefExp:
		return x.Line
	case *ast.FuncCallExp:
		return x.Line
	case *ast.TableConstructorExp:
		return x.Line
	case *ast.UnopExp:
		return x.Line
	case *ast.TableAccessExp:
		return lineOf(x.PrefixExp)
	case *ast.ConcatExp:
		return lineOf(x.Exps[0])
	case *ast.BinopExp:
		return lineOf(x.Exp1)
	default:
		panic("unreachable!")
	}
}

func lastLineOf(exp ast.Exp) int {
	switch x := exp.(type) {
	case *ast.NilExp:
		return x.Line
	case *ast.TrueExp:
		return x.Line
	case *ast.FalseExp:
		return x.Line
	case *ast.IntegerExp:
		return x.Line
	case *ast.FloatExp:
		return x.Line
	case *ast.StringExp:
		return x.Line
	case *ast.VarargExp:
		return x.Line
	case *ast.NameExp:
		return x.Line
	case *ast.FuncDefExp:
		return x.LastLine
	case *ast.FuncCallExp:
		return x.LastLine
	case *ast.TableConstructorExp:
		return x.LastLine
	case *ast.TableAccessExp:
		return x.LastLine
	case *ast.ConcatExp:
		return lastLineOf(x.Exps[len(x.Exps)-1])
	case *ast.BinopExp:
		return lastLineOf(x.Exp2)
	case *ast.UnopExp:
		return lastLineOf(x.Exp)
	default:
		panic("unreachable!")
	}
}
