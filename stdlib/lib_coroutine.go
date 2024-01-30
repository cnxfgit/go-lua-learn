package stdlib

import "luago/api"

var coFuncs = map[string]api.GoFunction{
	"create":      coCreate,    // coroutine.create (f)
	"resume":      coResume,    // coroutine.resume (co [, val1, ·])
	"yield":       coYield,     // coroutine.yield (·)
	"status":      coStatus,    // coroutine.status (co)
	"isyieldable": coYieldable, // coroutine.isyieldable ()
	"running":     coRunning,   // coroutine.running ()
	"wrap":        coWrap,      // coroutine.wrap (f)
}

func coCreate(ls api.LuaState) int {
	ls.CheckType(1, api.LUA_TFUNCTION)
	ls2 := ls.NewThread()
	ls.PushValue(1)
	ls.XMove(ls2, 1)
	return 1
}

func coResume(ls api.LuaState) int {
	co := ls.ToThread(1)
	ls.ArgCheck(co != nil, 1, "thread expected")

	if r := _auxResume(ls, co, ls.GetTop()-1); r < 0 {
		ls.PushBoolean(false)
		ls.Insert(-2)
		return 2
	} else {
		ls.PushBoolean(true)
		ls.Insert(-(r + 1))
		return r + 1
	}
}

func _auxResume(ls, co api.LuaState, narg int) int {
	if !ls.CheckStack(narg) {
		ls.PushString("too many arguments to resume")
		return -1
	}
	if co.Status() == api.LUA_OK && co.GetTop() == 0 {
		ls.PushString("cannot resume dead coroutine")
		return -1
	}

	ls.XMove(co, narg)
	status := co.Resume(ls, narg)
	if status == api.LUA_OK || status == api.LUA_YIELD {
		nres := co.GetTop()
		if !ls.CheckStack(nres + 1) {
			co.Pop(nres)
			ls.PushString("too many results to resume")
			return -1
		}
		co.XMove(ls, nres)
		return nres
	} else {
		co.XMove(ls, 1)
		return -1
	}
}

func coYield(ls api.LuaState) int {
	return ls.Yield(ls.GetTop())
}

func coStatus(ls api.LuaState) int {
	co := ls.ToThread(1)
	ls.ArgCheck(co != nil, 1, "thread expected")
	if ls == co {
		ls.PushString("running")
	} else {
		switch co.Status() {
		case api.LUA_YIELD:
			ls.PushString("suspended")
		case api.LUA_OK:
			if co.GetStack() {
				ls.PushString("normal")
			} else if co.GetTop() == 0 {
				ls.PushString("dead")
			} else {
				ls.PushString("suspended")
			}
		default:
			ls.PushString("dead")
		}
	}
	return 1
}

func OpenCoroutineLib(ls api.LuaState) int {
	ls.NewLib(coFuncs)
	return 1
}

func coYieldable(ls api.LuaState) int {
	ls.PushBoolean(ls.IsYieldable())
	return 1
}

func coRunning(ls api.LuaState) int {
	isMain := ls.PushThread()
	ls.PushBoolean(isMain)
	return 2
}

func coWrap(ls api.LuaState) int {
	panic("todo: coWrap!")
}
