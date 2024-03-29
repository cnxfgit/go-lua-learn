package vm

import "luago/api"

type Instruction uint32

const MAXARG_Bx = 1<<18 - 1       // 2 ^ 18 - 1 = 262143
const MAXARG_sBx = MAXARG_Bx >> 1 // 262143 / 2 = 131071

// 0x3F = 111111 提取指令后6位的opcode
func (i Instruction) Opcode() int {
	return int(i & 0x3F)
}

func (i Instruction) ABC() (a, b, c int) {
	a = int(i >> 6 & 0xFF)   // opcode 6位后的8位
	c = int(i >> 14 & 0x1FF) // opcode 6位+A8位后的9位
	b = int(i >> 23 & 0x1FF) // opcode 6位+A8位+C9位后的9位
	return
}

func (i Instruction) ABx() (a, bx int) {
	a = int(i >> 6 & 0xFF) // opcode 6位后的8位
	bx = int(i >> 14)      // 移除a 和 opcode 的14位 剩下的18位就是bx
	return
}

// sBx表示的是有符号整数
func (i Instruction) AsBx() (a, sbx int) {
	a, bx := i.ABx()
	return a, bx - MAXARG_sBx
}

func (i Instruction) Ax() int {
	return int(i >> 6) // 移除opcode 的6位 剩下的26位就是Ax
}

func (i Instruction) OpName() string {
	return opcodes[i.Opcode()].name
}

func (i Instruction) OpMode() byte {
	return opcodes[i.Opcode()].opMode
}

func (i Instruction) BMode() byte {
	return opcodes[i.Opcode()].argBMode
}

func (i Instruction) CMode() byte {
	return opcodes[i.Opcode()].argCMode
}

func (i Instruction) Execute(vm api.LuaVM) {
	action := opcodes[i.Opcode()].action
	if action != nil {
		action(i, vm)
	} else {
		panic(i.OpName())
	}
}
