package codegen

import (
	"luago/compiler/ast"
	"luago/compiler/lexer"
	"luago/vm"
)

var arithAndBitwiseBinops = map[int]int{
	lexer.TOKEN_OP_ADD:  vm.OP_ADD,
	lexer.TOKEN_OP_SUB:  vm.OP_SUB,
	lexer.TOKEN_OP_MUL:  vm.OP_MUL,
	lexer.TOKEN_OP_MOD:  vm.OP_MOD,
	lexer.TOKEN_OP_POW:  vm.OP_POW,
	lexer.TOKEN_OP_DIV:  vm.OP_DIV,
	lexer.TOKEN_OP_IDIV: vm.OP_IDIV,
	lexer.TOKEN_OP_BAND: vm.OP_BAND,
	lexer.TOKEN_OP_BOR:  vm.OP_BOR,
	lexer.TOKEN_OP_BXOR: vm.OP_BXOR,
	lexer.TOKEN_OP_SHL:  vm.OP_SHL,
	lexer.TOKEN_OP_SHR:  vm.OP_SHR,
}

type funcInfo struct {
	constants map[interface{}]int
	usedRegs  int
	maxRegs   int
	scopeLv   int
	locVars   []*locVarInfo
	locNames  map[string]*locVarInfo
	breaks    [][]int
	parent    *funcInfo
	upvalues  map[string]upvalInfo
	insts     []uint32
	subFuncs  []*funcInfo
	numParams int
	isVararg  bool
	lineNums  []uint32
	line      int
	lastLine  int
}

func newFuncInfo(parent *funcInfo, fd *ast.FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:    parent,
		subFuncs:  []*funcInfo{},
		constants: map[interface{}]int{},
		upvalues:  map[string]upvalInfo{},
		locNames:  map[string]*locVarInfo{},
		locVars:   make([]*locVarInfo, 0, 8),
		breaks:    make([][]int, 1),
		insts:     make([]uint32, 0, 8),
		isVararg:  fd.IsVararg,
		numParams: len(fd.ParList),
	}
}

func (fi *funcInfo) indexOfConstant(k interface{}) int {
	if idx, found := fi.constants[k]; found {
		return idx
	}

	idx := len(fi.constants)
	fi.constants[k] = idx
	return idx
}

func (fi *funcInfo) allocReg() int {
	fi.usedRegs++
	if fi.usedRegs >= 255 {
		panic("function or expression needs too many registers!")
	}
	if fi.usedRegs > fi.maxRegs {
		fi.maxRegs = fi.usedRegs
	}
	return fi.usedRegs - 1
}

func (fi *funcInfo) allocRegs(n int) int {
	if n <= 0 {
		panic("n <= 0 !")
	}
	for i := 0; i < n; i++ {
		fi.allocReg()
	}
	return fi.usedRegs - n
}

func (fi *funcInfo) freeReg() {
	fi.usedRegs--
}

func (fi *funcInfo) freeRegs(n int) {
	for i := 0; i < n; i++ {
		fi.freeReg()
	}
}

type locVarInfo struct {
	prev     *locVarInfo
	name     string
	scopeLv  int
	slot     int
	startPC  int
	endPC    int
	captured bool
}

func (fi *funcInfo) enterScope(breakable bool) {
	fi.scopeLv++
	if breakable {
		fi.breaks = append(fi.breaks, []int{})
	} else {
		fi.breaks = append(fi.breaks, nil)
	}
}

func (fi *funcInfo) addBreakJmp(pc int) {
	for i := fi.scopeLv; i >= 0; i-- {
		if fi.breaks[i] != nil {
			fi.breaks[i] = append(fi.breaks[i], pc)
			return
		}
	}
	panic("<break> at line ? not inside a loop!")
}

func (fi *funcInfo) addLocVar(name string, startPC int) int {
	newVar := &locVarInfo{
		name:    name,
		prev:    fi.locNames[name],
		scopeLv: fi.scopeLv,
		slot:    fi.allocReg(),
		startPC: startPC,
		endPC:   0,
	}

	fi.locVars = append(fi.locVars, newVar)
	fi.locNames[name] = newVar

	return newVar.slot
}

func (fi *funcInfo) slotOfLocVar(name string) int {
	if locVar, found := fi.locNames[name]; found {
		return locVar.slot
	}
	return -1
}

func (fi *funcInfo) exitScope(endPC int) {
	pendingBreakJmps := fi.breaks[len(fi.breaks)-1]
	fi.breaks = fi.breaks[:len(fi.breaks)-1]

	a := fi.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := fi.pc() - pc
		i := (sBx+vm.MAXARG_sBx)<<14 | a<<6 | vm.OP_JMP
		fi.insts[pc] = uint32(i)
	}

	fi.scopeLv--
	for _, locVar := range fi.locNames {
		if locVar.scopeLv > fi.scopeLv { // out of scope
			locVar.endPC = endPC
			fi.removeLocVar(locVar)
		}
	}
}

func (fi *funcInfo) removeLocVar(locVar *locVarInfo) {
	fi.freeReg()
	if locVar.prev == nil {
		delete(fi.locNames, locVar.name)
	} else if locVar.prev.scopeLv == locVar.scopeLv {
		fi.removeLocVar(locVar.prev)
	} else {
		fi.locNames[locVar.name] = locVar.prev
	}
}

type upvalInfo struct {
	locVarSlot int
	upvalIndex int
	index      int
}

func (fi *funcInfo) indexOfUpval(name string) int {
	if upval, ok := fi.upvalues[name]; ok {
		return upval.index
	}
	if fi.parent != nil {
		if locVar, found := fi.parent.locNames[name]; found {
			idx := len(fi.upvalues)
			fi.upvalues[name] = upvalInfo{
				locVarSlot: locVar.slot,
				upvalIndex: -1,
				index:      idx,
			}
			locVar.captured = true
			return idx
		}
		if uvIdx := fi.parent.indexOfUpval(name); uvIdx >= 0 {
			idx := len(fi.upvalues)
			fi.upvalues[name] = upvalInfo{
				locVarSlot: -1,
				upvalIndex: uvIdx,
				index:      idx,
			}
			return idx
		}
	}
	return -1
}

func (fi *funcInfo) emitABC(line, opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
	fi.lineNums = append(fi.lineNums, uint32(line))
}

func (fi *funcInfo) emitABx(line, opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
	fi.lineNums = append(fi.lineNums, uint32(line))
}

func (fi *funcInfo) emitAsBx(line, opcode, a, b int) {
	i := (b+vm.MAXARG_sBx)<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
	fi.lineNums = append(fi.lineNums, uint32(line))
}

func (fi *funcInfo) emitAx(line, opcode, ax int) {
	i := ax<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
	fi.lineNums = append(fi.lineNums, uint32(line))
}

func (fi *funcInfo) pc() int {
	return len(fi.insts) - 1
}

func (fi *funcInfo) fixSbx(pc, sBx int) {
	i := fi.insts[pc]
	i = i << 18 >> 18
	i = i | uint32(sBx+vm.MAXARG_sBx)<<14
	fi.insts[pc] = i
}

func (fi *funcInfo) closeOpenUpvals(line int) {
	a := fi.getJmpArgA()
	if a > 0 {
		fi.emitJmp(line, a, 0)
	}
}

func (fi *funcInfo) getJmpArgA() int {
	hasCapturedLocVars := false
	minSlotOfLocVars := fi.maxRegs
	for _, locVar := range fi.locNames {
		if locVar.scopeLv == fi.scopeLv {
			for v := locVar; v != nil && v.scopeLv == fi.scopeLv; v = v.prev {
				if v.captured {
					hasCapturedLocVars = true
				}
				if v.slot < minSlotOfLocVars && v.name[0] != '(' {
					minSlotOfLocVars = v.slot
				}
			}
		}
	}
	if hasCapturedLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}

func (fi *funcInfo) emitUnaryOp(line, op, a, b int) {
	switch op {
	case lexer.TOKEN_OP_NOT:
		fi.emitABC(line, vm.OP_NOT, a, b, 0)
	case lexer.TOKEN_OP_BNOT:
		fi.emitABC(line, vm.OP_BNOT, a, b, 0)
	case lexer.TOKEN_OP_LEN:
		fi.emitABC(line, vm.OP_LEN, a, b, 0)
	case lexer.TOKEN_OP_UNM:
		fi.emitABC(line, vm.OP_UNM, a, b, 0)
	}
}

func (fi *funcInfo) emitBinaryOp(line, op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		fi.emitABC(line, opcode, a, b, c)
	} else {
		switch op {
		case lexer.TOKEN_OP_EQ:
			fi.emitABC(line, vm.OP_EQ, 1, b, c)
		case lexer.TOKEN_OP_NE:
			fi.emitABC(line, vm.OP_EQ, 0, b, c)
		case lexer.TOKEN_OP_LT:
			fi.emitABC(line, vm.OP_LT, 1, b, c)
		case lexer.TOKEN_OP_GT:
			fi.emitABC(line, vm.OP_LT, 1, c, b)
		case lexer.TOKEN_OP_LE:
			fi.emitABC(line, vm.OP_LE, 1, b, c)
		case lexer.TOKEN_OP_GE:
			fi.emitABC(line, vm.OP_LE, 1, c, b)
		}
		fi.emitJmp(line, 0, 1)
		fi.emitLoadBool(line, a, 0, 1)
		fi.emitLoadBool(line, a, 1, 0)
	}
}

func (fi *funcInfo) emitLoadBool(line, a, b, c int) {
	fi.emitABC(line, vm.OP_LOADBOOL, a, b, c)
}

func (fi *funcInfo) emitJmp(line, a, sBx int) int {
	fi.emitAsBx(line, vm.OP_JMP, a, sBx)
	return len(fi.insts) - 1
}

func (fi *funcInfo) emitReturn(line, a, n int) {
	fi.emitABC(line, vm.OP_RETURN, a, n+1, 0)
}

func (fi *funcInfo) emitLoadNil(line, a, n int) {
	fi.emitABC(line, vm.OP_LOADNIL, a, n-1, 0)
}

func (fi *funcInfo) emitLoadK(line, a int, k interface{}) {
	idx := fi.indexOfConstant(k)
	if idx < (1 << 18) {
		fi.emitABx(line, vm.OP_LOADK, a, idx)
	} else {
		fi.emitABx(line, vm.OP_LOADKX, a, 0)
		fi.emitAx(line, vm.OP_EXTRAARG, idx)
	}
}

func (fi *funcInfo) emitVararg(line, a, n int) {
	fi.emitABC(line, vm.OP_VARARG, a, n+1, 0)
}

func (fi *funcInfo) emitClosure(line, a, bx int) {
	fi.emitABx(line, vm.OP_CLOSURE, a, bx)
}

func (fi *funcInfo) emitSetList(line, a, b, c int) {
	fi.emitABC(line, vm.OP_SETLIST, a, b, c)
}

func (fi *funcInfo) emitSetTable(line, a, b, c int) {
	fi.emitABC(line, vm.OP_SETTABLE, a, b, c)
}

func (fi *funcInfo) emitNewTable(line, a, nArr, nRec int) {
	fi.emitABC(line, vm.OP_NEWTABLE, a, vm.Int2fb(nArr), vm.Int2fb(nRec))
}

func (fi *funcInfo) emitTestSet(line, a, b, c int) {
	fi.emitABC(line, vm.OP_TESTSET, a, b, c)
}

func (fi *funcInfo) emitMove(line, a, b int) {
	fi.emitABC(line, vm.OP_MOVE, a, b, 0)
}

func (fi *funcInfo) emitGetUpval(line, a, b int) {
	fi.emitABC(line, vm.OP_GETUPVAL, a, b, 0)
}

func (fi *funcInfo) emitCall(line, a, nArgs, nRet int) {
	fi.emitABC(line, vm.OP_CALL, a, nArgs+1, nRet+1)
}

func (fi *funcInfo) emitGetTabUp(line, a, b, c int) {
	fi.emitABC(line, vm.OP_GETTABUP, a, b, c)
}

func (fi *funcInfo) emitGetTable(line, a, b, c int) {
	fi.emitABC(line, vm.OP_GETTABLE, a, b, c)
}

func (fi *funcInfo) emitSelf(line, a, b, c int) {
	fi.emitABC(line, vm.OP_SELF, a, b, c)
}

func (fi *funcInfo) emitTest(line, a, c int) {
	fi.emitABC(line, vm.OP_TEST, a, 0, c)
}

func (fi *funcInfo) emitForPrep(line, a, sBx int) int {
	fi.emitAsBx(line, vm.OP_FORPREP, a, sBx)
	return len(fi.insts) - 1
}

func (fi *funcInfo) emitForLoop(line, a, sBx int) int {
	fi.emitAsBx(line, vm.OP_FORLOOP, a, sBx)
	return len(fi.insts) - 1
}

func (fi *funcInfo) fixEndPC(name string, delta int) {
	for i := len(fi.locVars) - 1; i >= 0; i-- {
		locVar := fi.locVars[i]
		if locVar.name == name {
			locVar.endPC += delta
			return
		}
	}
}

func (fi *funcInfo) emitTForCall(line, a, c int) {
	fi.emitABC(line, vm.OP_TFORCALL, a, 0, c)
}

func (fi *funcInfo) emitTForLoop(line, a, sBx int) {
	fi.emitAsBx(line, vm.OP_TFORLOOP, a, sBx)
}

func (fi *funcInfo) emitSetTabUp(line, a, b, c int) {
	fi.emitABC(line, vm.OP_SETTABUP, a, b, c)
}

func (fi *funcInfo) emitSetUpval(line, a, b int) {
	fi.emitABC(line, vm.OP_SETUPVAL, a, b, 0)
}
