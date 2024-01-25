package vm

import "luago/api"

func tForLoop(i Instruction, vm api.LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	if !vm.IsNil(a + 1) {
		vm.Copy(a+1, a)
		vm.AddPC(sBx)
	}
}
