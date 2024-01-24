package binchunk

const (
	LUA_SIGNATURE    = "\x1bLua"
	LUAC_VERSION     = 0x53
	LUAC_FORMAT      = 0
	LUAC_DATA        = "\x19\x93\r\n\x1a\n"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 8
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_NUMBER_SIZE  = 8
	LUAC_INT         = 0x5678
	LUAC_NUM         = 370.5
)

const (
	TAG_NIL       = 0x00
	TAG_BOOLEAN   = 0x01
	TAG_NUMBER    = 0x03
	TAG_INTEGER   = 0x13
	TAG_SHORT_STR = 0x04
	TAG_LONG_STR  = 0x14
)

type binaryChunk struct {
	header                  // 头部
	sizeUpvalues byte       // 主函数upvalue数量
	mainFunc     *Prototype // 主函数原型
}

type header struct {
	signature       [4]byte // 签名 "\x1bLua"
	version         byte    // 版本号
	format          byte    // 格式号
	luacData        [6]byte // 校验数据 固定数据
	cintSize        byte    // int 字节数
	sizetSize       byte    // size_t 字节数
	instructionSize byte    // 虚拟机指令字节数
	luaIntegerSize  byte    // lua整数字节数
	luaNumberSize   byte    // lua浮点数字节数
	luacInt         int64   // luac整数值 为了检测机器大小端方式 0x5678
	luacNum         float64 // luac浮点型 检测机器浮点数格式和chunk是否匹配 370.5
}

type Prototype struct {
	Source          string        // 源文件名
	LineDefined     uint32        // 开始行
	LastLineDefined uint32        // 结束行
	NumParams       byte          // 固定参数数
	IsVararg        byte          // 是否有可变参数
	MaxStackSize    byte          // 寄存器数量
	Code            []uint32      // 指令表
	Constants       []interface{} // 常量表
	Upvalues        []Upvalue     // 提升值表
	Protos          []*Prototype  // 子函数原型表
	LineInfo        []uint32      // 行号表
	LocVars         []LocVar      // 局部变量表
	UpvalueNames    []string      // 提升值名
}

type Upvalue struct {
	Instack byte	// 1-当前函数的局部变量 0-已经被当前函数捕获
	Idx     byte
}

type LocVar struct {
	VarName string
	StartPC uint32
	EndPC   uint32
}

func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()        // 校验头部
	reader.readByte()           // 跳过upvalue数量
	return reader.readProto("") // 读取函数原型
}
