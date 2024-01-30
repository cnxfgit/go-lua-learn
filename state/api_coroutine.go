package state

import "luago/api"

func (ls *luaState) NewThread() api.LuaState {
	t := &luaState{registry: ls.registry}
	t.pushLuaStack(newLuaStack(api.LUAI_MAXSTACK, t))
	ls.stack.push(t)
	return t
}

func (ls *luaState) Resume(from api.LuaState, nArgs int) int {
	lsFrom := from.(*luaState)
	if lsFrom.coChan == nil {
		lsFrom.coChan = make(chan int)
	}

	if ls.coChan == nil {
		ls.coChan = make(chan int)
		ls.coCaller = lsFrom
		go func() {
			ls.coStatus = ls.PCall(nArgs, -1, 0)
			lsFrom.coChan <- 1
		}()
	} else {
		ls.coStatus = api.LUA_OK
		ls.coChan <- 1
	}

	<-lsFrom.coChan
	return ls.coStatus
}

func (ls *luaState) Yield(nResults int) int {
	ls.coStatus = api.LUA_YIELD
	ls.coCaller.coChan <- 1
	<-ls.coChan
	return ls.GetTop()
}

func (ls *luaState) Status() int {
	return ls.coStatus
}

func (ls *luaState) GetStack() bool {
	return ls.stack.prev != nil
}

func (ls *luaState) IsYieldable() bool {
	if ls.isMainThread() {
		return false
	}
	return ls.coStatus != api.LUA_YIELD // todo
}