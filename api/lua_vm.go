package api

type LuaVM interface {
	LuaState
	PC() int			// 返回当前PC 仅测试用
	AddPC(n int)		// 修改PC
	Fetch() uint32		// 取出当前指令： 将pc指向下一个指令
	GetConst(idx int)	// 将置顶常量推入栈顶
	GetRK(rk int)		// 将置顶常量或者值推入栈顶
}
