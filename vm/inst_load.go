package vm

import "luago/api"

// 在a的位置开始  放置b个nil
func loadNil(i Instruction, vm api.LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	vm.PushNil()
	for i := a; i <= a+b; i++ {
		vm.Copy(-1, i)
	}
	vm.Pop(1)
}

// 给寄存器a设定b（0-false else-true） c不为0 则跳过下一个指令
func loadBool(i Instruction, vm api.LuaVM) {
	a, b, c := i.ABC()
	a += 1

	vm.PushBoolean(b != 0)
	vm.Replace(a)
	if c != 0 {
		vm.AddPC(1)
	}
}

// 给寄存器a加载常量bx
func loadK(i Instruction, vm api.LuaVM) {
	a, bx := i.ABx()
	a += 1

	vm.GetConst(bx)
	vm.Replace(a)
}

// 给寄存器a加载常量ax
func loadKx(i Instruction, vm api.LuaVM) {
	a, _ := i.ABx()
	a += 1
	ax := Instruction(vm.Fetch()).Ax()

	vm.GetConst(ax)
	vm.Replace(a)
}
