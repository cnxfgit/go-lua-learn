package codegen

import "luago/compiler/ast"

func cgStat(fi *funcInfo, node ast.Stat) {
	switch stat := node.(type) {
	case *ast.FuncCallStat:
		cgFuncCallStat(fi, stat)
	case *ast.BreakStat:
		cgBreakStat(fi, stat)
	case *ast.DoStat:
		cgDoStat(fi, stat)
	case *ast.RepeatStat:
		cgRepeatStat(fi, stat)
	case *ast.WhileStat:
		cgWhileStat(fi, stat)
	case *ast.IfStat:
		cgIfStat(fi, stat)
	case *ast.ForNumStat:
		cgForNumStat(fi, stat)
	case *ast.ForInStat:
		cgForInStat(fi, stat)
	case *ast.AssignStat:
		cgAssignStat(fi, stat)
	case *ast.LocalVarDeclStat:
		cgLocalVarDeclStat(fi, stat)
	case *ast.LocalFuncDefStat:
		cgLocalFuncDefStat(fi, stat)
	case *ast.LabelStat, *ast.GotoStat:
		panic("not supported!")
	}
}

func cgLocalFuncDefStat(fi *funcInfo, node *ast.LocalFuncDefStat) {
	r := fi.addLocVar(node.Name, fi.pc()+2)
	cgFuncDefExp(fi, node.Exp, r)
}

func cgFuncCallStat(fi *funcInfo, node *ast.FuncCallStat) {
	r := fi.allocReg()
	cgFuncCallExp(fi, node, r, 0)
	fi.freeReg()
}

func cgBreakStat(fi *funcInfo, node *ast.BreakStat) {
	pc := fi.emitJmp(node.Line, 0, 0)
	fi.addBreakJmp(pc)
}

func cgDoStat(fi *funcInfo, node *ast.DoStat) {
	fi.enterScope(false)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.exitScope(fi.pc() + 1)
}

func cgWhileStat(fi *funcInfo, node *ast.WhileStat) {
	pcBeforeExp := fi.pc()

	oldRegs := fi.usedRegs
	a, _ := expToOpArg(fi, node.Exp, ARG_REG)
	fi.usedRegs = oldRegs

	line := lastLineOf(node.Exp)
	fi.emitTest(line, a, 0)
	pcJmpToEnd := fi.emitJmp(line, 0, 0)

	fi.enterScope(true)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.emitJmp(node.Block.LastLine, 0, pcBeforeExp-fi.pc()-1)
	fi.exitScope(fi.pc())

	fi.fixSbx(pcJmpToEnd, fi.pc()-pcJmpToEnd)
}

func cgRepeatStat(fi *funcInfo, node *ast.RepeatStat) {
	fi.enterScope(true)

	pcBeforeBlock := fi.pc()
	cgBlock(fi, node.Block)

	oldRegs := fi.usedRegs
	a, _ := expToOpArg(fi, node.Exp, ARG_REG)
	fi.usedRegs = oldRegs

	line := lastLineOf(node.Exp)
	fi.emitTest(line, a, 0)
	fi.emitJmp(line, fi.getJmpArgA(), pcBeforeBlock-fi.pc()-1)
	fi.closeOpenUpvals(line)

	fi.exitScope(fi.pc() + 1)
}


func cgIfStat(fi *funcInfo, node *ast.IfStat) {
	pcJmpToEnds := make([]int, len(node.Exps))
	pcJmpToNextExp := -1

	for i, exp := range node.Exps {
		if pcJmpToNextExp >= 0 {
			fi.fixSbx(pcJmpToNextExp, fi.pc()-pcJmpToNextExp)
		}

		oldRegs := fi.usedRegs
		a, _ := expToOpArg(fi, exp, ARG_REG)
		fi.usedRegs = oldRegs

		line := lastLineOf(exp)
		fi.emitTest(line, a, 0)
		pcJmpToNextExp = fi.emitJmp(line, 0, 0)

		block := node.Blocks[i]
		fi.enterScope(false)
		cgBlock(fi, block)
		fi.closeOpenUpvals(block.LastLine)
		fi.exitScope(fi.pc() + 1)
		if i < len(node.Exps)-1 {
			pcJmpToEnds[i] = fi.emitJmp(block.LastLine, 0, 0)
		} else {
			pcJmpToEnds[i] = pcJmpToNextExp
		}
	}

	for _, pc := range pcJmpToEnds {
		fi.fixSbx(pc, fi.pc()-pc)
	}
}

func cgForNumStat(fi *funcInfo, node *ast.ForNumStat) {
	forIndexVar := "(for index)"
	forLimitVar := "(for limit)"
	forStepVar := "(for step)"

	fi.enterScope(true)

	cgLocalVarDeclStat(fi, &ast.LocalVarDeclStat{
		NameList: []string{forIndexVar, forLimitVar, forStepVar},
		ExpList:  []ast.Exp{node.InitExp, node.LimitExp, node.StepExp},
	})
	fi.addLocVar(node.VarName, fi.pc()+2)

	a := fi.usedRegs - 4
	pcForPrep := fi.emitForPrep(node.LineOfDo, a, 0)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	pcForLoop := fi.emitForLoop(node.LineOfFor, a, 0)

	fi.fixSbx(pcForPrep, pcForLoop-pcForPrep-1)
	fi.fixSbx(pcForLoop, pcForPrep-pcForLoop)

	fi.exitScope(fi.pc())
	fi.fixEndPC(forIndexVar, 1)
	fi.fixEndPC(forLimitVar, 1)
	fi.fixEndPC(forStepVar, 1)
}

func cgForInStat(fi *funcInfo, node *ast.ForInStat) {
	forGeneratorVar := "(for generator)"
	forStateVar := "(for state)"
	forControlVar := "(for control)"

	fi.enterScope(true)

	cgLocalVarDeclStat(fi, &ast.LocalVarDeclStat{
		//LastLine: 0,
		NameList: []string{forGeneratorVar, forStateVar, forControlVar},
		ExpList:  node.ExpList,
	})
	for _, name := range node.NameList {
		fi.addLocVar(name, fi.pc()+2)
	}

	pcJmpToTFC := fi.emitJmp(node.LineOfDo, 0, 0)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.fixSbx(pcJmpToTFC, fi.pc()-pcJmpToTFC)

	line := lineOf(node.ExpList[0])
	rGenerator := fi.slotOfLocVar(forGeneratorVar)
	fi.emitTForCall(line, rGenerator, len(node.NameList))
	fi.emitTForLoop(line, rGenerator+2, pcJmpToTFC-fi.pc()-1)

	fi.exitScope(fi.pc() - 1)
	fi.fixEndPC(forGeneratorVar, 2)
	fi.fixEndPC(forStateVar, 2)
	fi.fixEndPC(forControlVar, 2)
}

func cgLocalVarDeclStat(fi *funcInfo, node *ast.LocalVarDeclStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nNames := len(node.NameList)

	oldRegs := fi.usedRegs
	if nExps == nNames {
		for _, exp := range exps {
			a := fi.allocReg()
			cgExp(fi, exp, a, 1)
		}
	} else if nExps > nNames {
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else { // nNames > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nNames - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
		if !multRet {
			n := nNames - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(node.LastLine, a, n)
		}
	}

	fi.usedRegs = oldRegs
	startPC := fi.pc() + 1
	for _, name := range node.NameList {
		fi.addLocVar(name, startPC)
	}
}

func cgAssignStat(fi *funcInfo, node *ast.AssignStat) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nVars := len(node.VarList)

	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	vRegs := make([]int, nVars)
	oldRegs := fi.usedRegs

	for i, exp := range node.VarList {
		if taExp, ok := exp.(*ast.TableAccessExp); ok {
			tRegs[i] = fi.allocReg()
			cgExp(fi, taExp.PrefixExp, tRegs[i], 1)
			kRegs[i] = fi.allocReg()
			cgExp(fi, taExp.KeyExp, kRegs[i], 1)
		} else {
			name := exp.(*ast.NameExp).Name
			if fi.slotOfLocVar(name) < 0 && fi.indexOfUpval(name) < 0 {
				// global var
				kRegs[i] = -1
				if fi.indexOfConstant(name) > 0xFF {
					kRegs[i] = fi.allocReg()
				}
			}
		}
	}
	for i := 0; i < nVars; i++ {
		vRegs[i] = fi.usedRegs + i
	}

	if nExps >= nVars {
		for i, exp := range exps {
			a := fi.allocReg()
			if i >= nVars && i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else { // nVars > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
		if !multRet {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(node.LastLine, a, n)
		}
	}

	lastLine := node.LastLine
	for i, exp := range node.VarList {
		if nameExp, ok := exp.(*ast.NameExp); ok {
			varName := nameExp.Name
			if a := fi.slotOfLocVar(varName); a >= 0 {
				fi.emitMove(lastLine, a, vRegs[i])
			} else if b := fi.indexOfUpval(varName); b >= 0 {
				fi.emitSetUpval(lastLine, vRegs[i], b)
			} else if a := fi.slotOfLocVar("_ENV"); a >= 0 {
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTable(lastLine, a, b, vRegs[i])
				} else {
					fi.emitSetTable(lastLine, a, kRegs[i], vRegs[i])
				}
			} else { // global var
				a := fi.indexOfUpval("_ENV")
				if kRegs[i] < 0 {
					b := 0x100 + fi.indexOfConstant(varName)
					fi.emitSetTabUp(lastLine, a, b, vRegs[i])
				} else {
					fi.emitSetTabUp(lastLine, a, kRegs[i], vRegs[i])
				}
			}
		} else {
			fi.emitSetTable(lastLine, tRegs[i], kRegs[i], vRegs[i])
		}
	}

	// todo
	fi.usedRegs = oldRegs
}
