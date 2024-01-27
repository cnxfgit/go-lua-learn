package state

import (
	"luago/api"
	"luago/binchunk"
	"luago/compiler"
	"luago/vm"
)

func (ls *luaState) Load(chunk []byte, chunkName string, mode string) int {
	var proto *binchunk.Prototype
	if binchunk.IsBinaryChunk(chunk) {
		proto = binchunk.Undump(chunk)
	} else {
		proto = compiler.Compile(string(chunk), chunkName)
	}

	c := newLuaClosure(proto)
	ls.stack.push(c)
	if len(proto.Upvalues) > 0 {
		env := ls.registry.get(api.LUA_RIDX_GLOBALS)
		c.upvals[0] = &upvalue{&env}
	}
	return api.LUA_OK
}

func (ls *luaState) Call(nArgs, nResults int) {
	val := ls.stack.get(-(nArgs + 1))

	c, ok := val.(*closure)
	if !ok {
		if mf := getMetafield(val, "__call", ls); mf != nil {
			if c, ok = mf.(*closure); ok {
				ls.stack.push(val)
				ls.Insert(-(nArgs + 2))
				nArgs += 1
			}
		}
	}

	if ok {
		if c.proto != nil {
			ls.callLuaClosure(nArgs, nResults, c)
		} else {
			ls.callGoClosure(nArgs, nResults, c)
		}
	} else {
		panic("not function!")
	}
}

func (ls *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1

	newStack := newLuaStack(nRegs+api.LUA_MINSTACK, ls)
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

func (ls *luaState) callGoClosure(nArgs, nResults int, c *closure) {
	newStack := newLuaStack(nArgs+api.LUA_MINSTACK, ls)
	newStack.closure = c

	if nArgs > 0 {
		args := ls.stack.popN(nArgs)
		newStack.pushN(args, nArgs)
	}
	ls.stack.pop()

	ls.pushLuaStack(newStack)
	r := c.goFunc(ls)
	ls.popLuaStack()

	if nResults != 0 {
		results := newStack.popN(r)
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

func (ls *luaState) PCall(nArgs, nResults, msgh int) (status int) {
	caller := ls.stack
	status = api.LUA_ERRRUN

	defer func() {
		if err := recover(); err != nil {
			for ls.stack != caller {
				ls.popLuaStack()
			}
			ls.stack.push(err)
		}
	}()

	ls.Call(nArgs, nResults)
	status = api.LUA_OK
	return
}
