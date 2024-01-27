package codegen

import "luago/compiler/ast"

func cgBlock(fi *funcInfo, node *ast.Block) {
	for _, stat := range node.Stats {
		cgStat(fi, stat)
	}

	if node.RetExps != nil {
		cgRetStat(fi, node.RetExps, node.LastLine)
	}
}

func cgRetStat(fi *funcInfo, exps []ast.Exp, lastLine int) {
	nExps := len(exps)
	if nExps == 0 {
		fi.emitReturn(lastLine, 0, 0)
		return
	}

	multRet := isVarargOrFuncCall(exps[nExps-1])
	for i, exp := range exps {
		r := fi.allocReg()
		if i == nExps-1 && multRet {
			cgExp(fi, exp, r, -1)
		} else {
			cgExp(fi, exp, r, 1)
		}
	}
	fi.freeRegs(nExps)

	a := fi.usedRegs
	if multRet {
		fi.emitReturn(lastLine, a, -1)
	} else {
		fi.emitReturn(lastLine, a, nExps)
	}
}