package stdlib

import "fmt"
import "strings"
import "luago/api"

var strLib = map[string]api.GoFunction{
	"len":      strLen,
	"rep":      strRep,
	"reverse":  strReverse,
	"lower":    strLower,
	"upper":    strUpper,
	"sub":      strSub,
	"byte":     strByte,
	"char":     strChar,
	"dump":     strDump,
	"format":   strFormat,
	"packsize": strPackSize,
	"pack":     strPack,
	"unpack":   strUnpack,
	"find":     strFind,
	"match":    strMatch,
	"gsub":     strGsub,
	"gmatch":   strGmatch,
}

func OpenStringLib(ls api.LuaState) int {
	ls.NewLib(strLib)
	createMetatable(ls)
	return 1
}

func createMetatable(ls api.LuaState) {
	ls.CreateTable(0, 1)       /* table to be metatable for strings */
	ls.PushString("dummy")     /* dummy string */
	ls.PushValue(-2)           /* copy table */
	ls.SetMetatable(-2)        /* set table as metatable for strings */
	ls.Pop(1)                  /* pop dummy string */
	ls.PushValue(-2)           /* get string library */
	ls.SetField(-2, "__index") /* metatable.__index = string */
	ls.Pop(1)                  /* pop metatable */
}


func strLen(ls api.LuaState) int {
	s := ls.CheckString(1)
	ls.PushInteger(int64(len(s)))
	return 1
}


func strRep(ls api.LuaState) int {
	s := ls.CheckString(1)
	n := ls.CheckInteger(2)
	sep := ls.OptString(3, "")

	if n <= 0 {
		ls.PushString("")
	} else if n == 1 {
		ls.PushString(s)
	} else {
		a := make([]string, n)
		for i := 0; i < int(n); i++ {
			a[i] = s
		}
		ls.PushString(strings.Join(a, sep))
	}

	return 1
}

func strReverse(ls api.LuaState) int {
	s := ls.CheckString(1)

	if strLen := len(s); strLen > 1 {
		a := make([]byte, strLen)
		for i := 0; i < strLen; i++ {
			a[i] = s[strLen-1-i]
		}
		ls.PushString(string(a))
	}

	return 1
}

func strLower(ls api.LuaState) int {
	s := ls.CheckString(1)
	ls.PushString(strings.ToLower(s))
	return 1
}

func strUpper(ls api.LuaState) int {
	s := ls.CheckString(1)
	ls.PushString(strings.ToUpper(s))
	return 1
}


func strSub(ls api.LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	i := posRelat(ls.CheckInteger(2), sLen)
	j := posRelat(ls.OptInteger(3, -1), sLen)

	if i < 1 {
		i = 1
	}
	if j > sLen {
		j = sLen
	}

	if i <= j {
		ls.PushString(s[i-1 : j])
	} else {
		ls.PushString("")
	}

	return 1
}

func strByte(ls api.LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	i := posRelat(ls.OptInteger(2, 1), sLen)
	j := posRelat(ls.OptInteger(3, int64(i)), sLen)

	if i < 1 {
		i = 1
	}
	if j > sLen {
		j = sLen
	}

	if i > j {
		return 0 /* empty interval; return no values */
	}
	//if (j - i >= INT_MAX) { /* arithmetic overflow? */
	//  return ls.Error2("string slice too long")
	//}

	n := j - i + 1
	ls.CheckStack2(n, "string slice too long")

	for k := 0; k < n; k++ {
		ls.PushInteger(int64(s[i+k-1]))
	}
	return n
}

func strChar(ls api.LuaState) int {
	nArgs := ls.GetTop()

	s := make([]byte, nArgs)
	for i := 1; i <= nArgs; i++ {
		c := ls.CheckInteger(i)
		ls.ArgCheck(int64(byte(c)) == c, i, "value out of range")
		s[i-1] = byte(c)
	}

	ls.PushString(string(s))
	return 1
}


func strDump(ls api.LuaState) int {
	// strip := ls.ToBoolean(2)
	// ls.CheckType(1, LUA_TFUNCTION)
	// ls.SetTop(1)
	// ls.PushString(string(ls.Dump(strip)))
	// return 1
	panic("todo: strDump!")
}

func strPackSize(ls api.LuaState) int {
	fmt := ls.CheckString(1)
	if fmt == "j" {
		ls.PushInteger(8) // todo
	} else {
		panic("todo: strPackSize!")
	}
	return 1
}

func strPack(ls api.LuaState) int {
	panic("todo: strPack!")
}

func strUnpack(ls api.LuaState) int {
	panic("todo: strUnpack!")
}

func strFormat(ls api.LuaState) int {
	fmtStr := ls.CheckString(1)
	if len(fmtStr) <= 1 || strings.IndexByte(fmtStr, '%') < 0 {
		ls.PushString(fmtStr)
		return 1
	}

	argIdx := 1
	arr := parseFmtStr(fmtStr)
	for i, s := range arr {
		if s[0] == '%' {
			if s == "%%" {
				arr[i] = "%"
			} else {
				argIdx += 1
				arr[i] = _fmtArg(s, ls, argIdx)
			}
		}
	}

	ls.PushString(strings.Join(arr, ""))
	return 1
}

func _fmtArg(tag string, ls api.LuaState, argIdx int) string {
	switch tag[len(tag)-1] { // specifier
	case 'c': // character
		return string([]byte{byte(ls.ToInteger(argIdx))})
	case 'i':
		tag = tag[:len(tag)-1] + "d" // %i -> %d
		return fmt.Sprintf(tag, ls.ToInteger(argIdx))
	case 'd', 'o': // integer, octal
		return fmt.Sprintf(tag, ls.ToInteger(argIdx))
	case 'u': // unsigned integer
		tag = tag[:len(tag)-1] + "d" // %u -> %d
		return fmt.Sprintf(tag, uint(ls.ToInteger(argIdx)))
	case 'x', 'X': // hex integer
		return fmt.Sprintf(tag, uint(ls.ToInteger(argIdx)))
	case 'f': // float
		return fmt.Sprintf(tag, ls.ToNumber(argIdx))
	case 's', 'q': // string
		return fmt.Sprintf(tag, ls.ToString2(argIdx))
	default:
		panic("todo! tag=" + tag)
	}
}

func strFind(ls api.LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	pattern := ls.CheckString(2)
	init := posRelat(ls.OptInteger(3, 1), sLen)
	if init < 1 {
		init = 1
	} else if init > sLen+1 { /* start after string's end? */
		ls.PushNil()
		return 1
	}
	plain := ls.ToBoolean(4)

	start, end := find(s, pattern, init, plain)

	if start < 0 {
		ls.PushNil()
		return 1
	}
	ls.PushInteger(int64(start))
	ls.PushInteger(int64(end))
	return 2
}

func strMatch(ls api.LuaState) int {
	s := ls.CheckString(1)
	sLen := len(s)
	pattern := ls.CheckString(2)
	init := posRelat(ls.OptInteger(3, 1), sLen)
	if init < 1 {
		init = 1
	} else if init > sLen+1 { /* start after string's end? */
		ls.PushNil()
		return 1
	}

	captures := match(s, pattern, init)

	if captures == nil {
		ls.PushNil()
		return 1
	} else {
		for i := 0; i < len(captures); i += 2 {
			capture := s[captures[i]:captures[i+1]]
			ls.PushString(capture)
		}
		return len(captures) / 2
	}
}

func strGsub(ls api.LuaState) int {
	s := ls.CheckString(1)
	pattern := ls.CheckString(2)
	repl := ls.CheckString(3) // todo
	n := int(ls.OptInteger(4, -1))

	newStr, nMatches := gsub(s, pattern, repl, n)
	ls.PushString(newStr)
	ls.PushInteger(int64(nMatches))
	return 2
}

func strGmatch(ls api.LuaState) int {
	s := ls.CheckString(1)
	pattern := ls.CheckString(2)

	gmatchAux := func(ls api.LuaState) int {
		captures := match(s, pattern, 1)
		if captures != nil {
			for i := 0; i < len(captures); i += 2 {
				capture := s[captures[i]:captures[i+1]]
				ls.PushString(capture)
			}
			s = s[captures[len(captures)-1]:]
			return len(captures) / 2
		} else {
			return 0
		}
	}

	ls.PushGoFunction(gmatchAux)
	return 1
}

func posRelat(pos int64, _len int) int {
	_pos := int(pos)
	if _pos >= 0 {
		return _pos
	} else if -_pos > _len {
		return 0
	} else {
		return _len + _pos + 1
	}
}
