package vm

import "luago/api"

// 将b c两个寄存器的值op运算 放到寄存器a
func _binaryArith(i Instruction, vm api.LuaVM, op api.ArithOp) {
	a, b, c := i.ABC()
	a += 1

	vm.GetRK(b)
	vm.GetRK(c)
	vm.Arith(op)
	vm.Replace(a)
}

func _unaryArith(i Instruction, vm api.LuaVM, op api.ArithOp) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.PushValue(b)
	vm.Arith(op)
	vm.Replace(a)
}

// +
func add(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPADD)
}

// -
func sub(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPSUB)
}

// *
func mul(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPMUL)
}

// %
func mod(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPMOD)
}

// ^
func pow(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPPOW)
}

// /
func div(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPDIV)
}

// //
func idiv(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPIDIV)
}

// &
func band(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPBAND)
}

// |
func bor(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPBOR)
}

// ~
func bxor(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPBXOR)
}

// <<
func shl(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPSHL)
}

// >>
func shr(i Instruction, vm api.LuaVM) {
	_binaryArith(i, vm,
		api.LUA_OPSHR)
}

// -
func unm(i Instruction, vm api.LuaVM) {
	_unaryArith(i, vm,
		api.LUA_OPUNM)
}

// ~
func bnot(i Instruction, vm api.LuaVM) {
	_unaryArith(i, vm,
		api.LUA_OPBNOT)
}

// 将b寄存器的值求长度 将长度放到a寄存器
func _len(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Len(b)
	vm.Replace(a)
}

// 将寄存器b到寄存器c之间的值拼接 放到寄存器a
func concat(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	c += 1

	n := c - b + 1
	vm.CheckStack(n)
	for i := b; i <= c; i++ {
		vm.PushValue(i)
	}
	vm.Concat(n)
	vm.Replace(a)
}

// 将寄存器b c的值比较， 如何结果和 a 一致则pc+1
func _compare(i Instruction, vm api.LuaVM, op api.CompareOp) {
	a, b, c := i.ABC()

	vm.GetRK(b)
	vm.GetRK(c)
	if vm.Compare(-2, -1, op) != (a != 0) {
		vm.AddPC(1)
	}
	vm.Pop(2)
}

func eq(i Instruction, vm api.LuaVM) {
	_compare(i, vm, api.LUA_OPEQ)
}

func lt(i Instruction, vm api.LuaVM) {
	_compare(i, vm, api.LUA_OPLT)
}

func le(i Instruction, vm api.LuaVM) {
	_compare(i, vm, api.LUA_OPLE)
}

// 将b寄存器的值取反 放到a寄存器
func not(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.PushBoolean(!vm.ToBoolean(b))
	vm.Replace(a)
}

// 判断寄存器b转化成布尔值之后是否和c一致，一致复制b的值到a，否则跳过下一条指令
func testSet(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	if vm.ToBoolean(b) == (c != 0) {
		vm.Copy(b, a)
	} else {
		vm.AddPC(1)
	}
}

// 判断寄存器a转化成布尔值之后是否和c一致，一致跳过下一条指令
func test(i Instruction, vm api.LuaVM) {
	a, _, c := i.ABC()
	a += 1

	if vm.ToBoolean(a) != (c != 0) {
		vm.AddPC(1)
	}
}

func forPrep(i Instruction, vm api.LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	// R(A) -= R(A+2)
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(api.LUA_OPSUB)
	vm.Replace(a)

	// pc+=sBx
	vm.AddPC(sBx)
}

func forLoop(i Instruction, vm api.LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	// R(A) += R(A+2)
	vm.PushValue(a + 2)
	vm.PushValue(a)
	vm.Arith(api.LUA_OPADD)
	vm.Replace(a)

	// R(A) <?= R(A+1)
	isPositiveStep := vm.ToNumber(a+2) >= 0
	if isPositiveStep && vm.Compare(a, a+1, api.LUA_OPLE) ||
		!isPositiveStep && vm.Compare(a+1, a, api.LUA_OPLE) {
		vm.AddPC(sBx)   // pc += sBx
		vm.Copy(a, a+3) // R(A+3) = R(A)
	}
}
