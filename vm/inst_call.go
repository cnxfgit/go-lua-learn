package vm

import "luago/api"

func closure(i Instruction, vm api.LuaVM) {
	a, bx := i.ABx()
	a += 1

	vm.LoadProto(bx)
	vm.Replace(a)
}

func call(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1

	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	_popResults(a, c, vm)
}

func _pushFuncAndArgs(a, b int, vm api.LuaVM) (nArgs int) {
	if b >= 1 { // b-1 args
		vm.CheckStack(b)
		for i := a; i < a+b; i++ {
			vm.PushValue(i)
		}
		return b - 1
	} else {
		_fixStack(a, vm)
		return vm.GetTop() - vm.RegisterCount() - 1
	}
}

func _popResults(a, c int, vm api.LuaVM) {
	if c == 1 { // no result
	} else if c > 1 {
		for i := a + c - 2; i >= a; i-- {
			vm.Replace(i)
		}
	} else {
		vm.CheckStack(1)
		vm.PushInteger(int64(a))
	}
}

func _fixStack(a int, vm api.LuaVM) {
	x := int(vm.ToInteger(-1))
	vm.Pop(1)

	vm.CheckStack(x - a)
	for i := a; i < x; i++ {
		vm.PushValue(i)
	}
	vm.Rotate(vm.RegisterCount()+1, x-a)
}

func _return(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	if b == 1 { // no return values
	} else if b > 1 {
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else {
		_fixStack(a, vm)
	}
}

func vararg(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	if b != 1 {
		vm.LoadVararg(b - 1)
		_popResults(a, b, vm)
	}
}

func tailCall(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	c := 0

	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	_popResults(a, c, vm)
}

func self(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.Copy(b, a+1)
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

func tForCall(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a += 1

	_pushFuncAndArgs(a, 3, vm)
	vm.Call(2, c)
	_popResults(a+3, c+1, vm)
}
