package state

import (
	"fmt"
	"luago/binchunk"
	"luago/vm"
)

func (ls *luaState) Load(chunk []byte, chunkName string, mode string) int {
	proto := binchunk.Undump(chunk)
	c := newLuaClosure(proto)
	ls.stack.push(c)
	return 0
}

func (ls *luaState) Call(nArgs, nResults int) {
	val := ls.stack.get(-(nArgs + 1))
	if c, ok := val.(*closure); ok {
		fmt.Printf("call %s<%d,%d>\n", c.proto.Source,
			c.proto.LineDefined, c.proto.LastLineDefined)
		ls.callLuaClosure(nArgs, nResults, c)
	} else {
		panic("not function!")
	}
}

func (ls *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	newStack := newLuaStack(nRegs + 20)
	newStack.closure = c

	funcAndArgs := ls.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	ls.pushLuaStack(newStack)
	ls.runLuaClosure()
	ls.popLuaStack()

	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		ls.stack.check(len(results))
		ls.stack.pushN(results, nResults)
	}
}

func (ls *luaState) runLuaClosure() {
	for {
		inst := vm.Instruction(ls.Fetch())
		inst.Execute(ls)
		if inst.Opcode() == vm.OP_RETURN {
			break
		}
	}
}
