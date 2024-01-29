package stdlib

import (
	"luago/api"
	"luago/number"
	"math"
	"math/rand"
)

var mathLib = map[string]api.GoFunction{
	"random":     mathRandom,     // 计算随机数
	"randomseed": mathRandomSeed, // 设置随机数种子
	"max":        mathMax,        // 求最大值
	"min":        mathMin,        // 求最小值
	"exp":        mathExp,        // 计算e的指数
	"log":        mathLog,        // 计算对数
	"deg":        mathDeg,        // 把弧度转为角度
	"rad":        mathRad,        // 把角度转为弧度
	"sin":        mathSin,        // 正弦函数
	"cos":        mathCos,        // 余弦函数
	"tan":        mathTan,        // 正切函数
	"asin":       mathAsin,       // 反正弦函数
	"acos":       mathAcos,       // 反余弦函数
	"atan":       mathAtan,       // 反正切函数
	"ceil":       mathCeil,       // 向上取整
	"floor":      mathFloor,      // 向下取整
	"abs":        mathAbs,        // 求绝对值
	"sqrt":       mathSqrt,       // 开平方
	"fmod":       mathFmod,       // 请参考Lua手册
	"modf":       mathModf,       // 请参考Lua手册
	"ult":        mathUlt,        // 请参考Lua手册
	"tointeger":  mathToInt,      //
	"type":       mathType,       //
}

func OpenMathLib(ls api.LuaState) int {
	ls.NewLib(mathLib)
	ls.PushNumber(math.Pi)
	ls.SetField(-2, "pi")
	ls.PushNumber(math.Inf(1))
	ls.SetField(-2, "huge")
	ls.PushInteger(math.MaxInt64)
	ls.SetField(-2, "maxinteger")
	ls.PushInteger(math.MinInt64)
	ls.SetField(-2, "mininteger")
	return 1
}

func mathRandom(ls api.LuaState) int {
	var low, up int64
	switch ls.GetTop() { /* check number of arguments */
	case 0: /* no arguments */
		ls.PushNumber(rand.Float64()) /* Number between 0 and 1 */
		return 1
	case 1: /* only upper limit */
		low = 1
		up = ls.CheckInteger(1)
	case 2: /* lower and upper limits */
		low = ls.CheckInteger(1)
		up = ls.CheckInteger(2)
	default:
		return ls.Error2("wrong number of arguments")
	}

	/* random integer in the interval [low, up] */
	ls.ArgCheck(low <= up, 1, "interval is empty")
	ls.ArgCheck(low >= 0 || up <= math.MaxInt64+low, 1,
		"interval too large")
	if up-low == math.MaxInt64 {
		ls.PushInteger(low + rand.Int63())
	} else {
		ls.PushInteger(low + rand.Int63n(up-low+1))
	}
	return 1
}

func mathRandomSeed(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	rand.Seed(int64(x))
	return 0
}

func mathMax(ls api.LuaState) int {
	n := ls.GetTop() /* number of arguments */
	imax := 1        /* index of current maximum value */
	ls.ArgCheck(n >= 1, 1, "value expected")
	for i := 2; i <= n; i++ {
		if ls.Compare(imax, i, api.LUA_OPLT) {
			imax = i
		}
	}
	ls.PushValue(imax)
	return 1
}

func mathMin(ls api.LuaState) int {
	n := ls.GetTop() /* number of arguments */
	imin := 1        /* index of current minimum value */
	ls.ArgCheck(n >= 1, 1, "value expected")
	for i := 2; i <= n; i++ {
		if ls.Compare(i, imin, api.LUA_OPLT) {
			imin = i
		}
	}
	ls.PushValue(imin)
	return 1
}

func mathExp(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(math.Exp(x))
	return 1
}

func mathLog(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	var res float64

	if ls.IsNoneOrNil(2) {
		res = math.Log(x)
	} else {
		base := ls.ToNumber(2)
		if base == 2 {
			res = math.Log2(x)
		} else if base == 10 {
			res = math.Log10(x)
		} else {
			res = math.Log(x) / math.Log(base)
		}
	}

	ls.PushNumber(res)
	return 1
}

func mathDeg(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(x * 180 / math.Pi)
	return 1
}

func mathRad(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(x * math.Pi / 180)
	return 1
}

func mathSin(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(math.Sin(x))
	return 1
}

func mathCos(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(math.Cos(x))
	return 1
}

func mathTan(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(math.Tan(x))
	return 1
}

func mathAsin(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(math.Asin(x))
	return 1
}

func mathAcos(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(math.Acos(x))
	return 1
}

func mathAtan(ls api.LuaState) int {
	y := ls.CheckNumber(1)
	x := ls.OptNumber(2, 1.0)
	ls.PushNumber(math.Atan2(y, x))
	return 1
}

/* rounding functions */

func mathCeil(ls api.LuaState) int {
	if ls.IsInteger(1) {
		ls.SetTop(1) /* integer is its own ceil */
	} else {
		x := ls.CheckNumber(1)
		_pushNumInt(ls, math.Ceil(x))
	}
	return 1
}

func mathFloor(ls api.LuaState) int {
	if ls.IsInteger(1) {
		ls.SetTop(1) /* integer is its own floor */
	} else {
		x := ls.CheckNumber(1)
		_pushNumInt(ls, math.Floor(x))
	}
	return 1
}

func mathFmod(ls api.LuaState) int {
	if ls.IsInteger(1) && ls.IsInteger(2) {
		d := ls.ToInteger(2)
		if uint64(d)+1 <= 1 { /* special cases: -1 or 0 */
			ls.ArgCheck(d != 0, 2, "zero")
			ls.PushInteger(0) /* avoid overflow with 0x80000... / -1 */
		} else {
			ls.PushInteger(ls.ToInteger(1) % d)
		}
	} else {
		x := ls.CheckNumber(1)
		y := ls.CheckNumber(2)
		ls.PushNumber(x - math.Trunc(x/y)*y)
	}

	return 1
}

func mathModf(ls api.LuaState) int {
	if ls.IsInteger(1) {
		ls.SetTop(1)     /* number is its own integer part */
		ls.PushNumber(0) /* no fractional part */
	} else {
		x := ls.CheckNumber(1)
		i, f := math.Modf(x)
		_pushNumInt(ls, i)
		if math.IsInf(x, 0) {
			ls.PushNumber(0)
		} else {
			ls.PushNumber(f)
		}
	}

	return 2
}


func mathAbs(ls api.LuaState) int {
	if ls.IsInteger(1) {
		x := ls.ToInteger(1)
		if x < 0 {
			ls.PushInteger(-x)
		}
	} else {
		x := ls.CheckNumber(1)
		ls.PushNumber(math.Abs(x))
	}
	return 1
}

func mathSqrt(ls api.LuaState) int {
	x := ls.CheckNumber(1)
	ls.PushNumber(math.Sqrt(x))
	return 1
}

func mathUlt(ls api.LuaState) int {
	m := ls.CheckInteger(1)
	n := ls.CheckInteger(2)
	ls.PushBoolean(uint64(m) < uint64(n))
	return 1
}

func mathToInt(ls api.LuaState) int {
	if i, ok := ls.ToIntegerX(1); ok {
		ls.PushInteger(i)
	} else {
		ls.CheckAny(1)
		ls.PushNil() /* value is not convertible to integer */
	}
	return 1
}

func mathType(ls api.LuaState) int {
	if ls.Type(1) == api.LUA_TNUMBER {
		if ls.IsInteger(1) {
			ls.PushString("integer")
		} else {
			ls.PushString("float")
		}
	} else {
		ls.CheckAny(1)
		ls.PushNil()
	}
	return 1
}

func _pushNumInt(ls api.LuaState, d float64) {
	if i, ok := number.FloatToInteger(d); ok { /* does 'd' fit in an integer? */
		ls.PushInteger(i) /* result is integer */
	} else {
		ls.PushNumber(d) /* result is float */
	}
}

