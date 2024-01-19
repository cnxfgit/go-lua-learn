package state

func (ls *luaState) PC() int {
	return ls.pc
}

func (ls *luaState) AddPC(n int) {
	ls.pc += n
}

func (ls *luaState) Fetch() uint32 {
	i := ls.proto.Code[ls.pc]
	ls.pc++
	return i
}

func (ls *luaState) GetConst(idx int) {
	c := ls.proto.Constants[idx]
	ls.stack.push(c)
}

// 传递的时OpArgK 9位  首位决定时常量表还是寄存器
func (ls *luaState) GetRK(rk int) {
	if rk > 0xFF { // constant
		ls.GetConst(rk & 0xFF)
	} else { // register
		ls.PushValue(rk + 1)
	}
}
