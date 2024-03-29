package stdlib

//#include <time.h>
import "C"

import "os"
import "time"
import "luago/api"

var sysLib = map[string]api.GoFunction{
	"clock":     osClock,
	"difftime":  osDiffTime,
	"time":      osTime,
	"date":      osDate,
	"remove":    osRemove,
	"rename":    osRename,
	"tmpname":   osTmpName,
	"getenv":    osGetEnv,
	"execute":   osExecute,
	"exit":      osExit,
	"setlocale": osSetLocale,
}

func OpenOSLib(ls api.LuaState) int {
	ls.NewLib(sysLib)
	return 1
}

func osClock(ls api.LuaState) int {
	c := float64(C.clock()) / float64(C.CLOCKS_PER_SEC)
	ls.PushNumber(c)
	return 1
}

func osDiffTime(ls api.LuaState) int {
	t2 := ls.CheckInteger(1)
	t1 := ls.CheckInteger(2)
	ls.PushInteger(t2 - t1)
	return 1
}

func osTime(ls api.LuaState) int {
	if ls.IsNoneOrNil(1) { /* called without args? */
		t := time.Now().Unix() /* get current time */
		ls.PushInteger(t)
	} else {
		ls.CheckType(1, api.LUA_TTABLE)
		sec := _getField(ls, "sec", 0)
		min := _getField(ls, "min", 0)
		hour := _getField(ls, "hour", 12)
		day := _getField(ls, "day", -1)
		month := _getField(ls, "month", -1)
		year := _getField(ls, "year", -1)
		// todo: isdst
		t := time.Date(year, time.Month(month), day,
			hour, min, sec, 0, time.Local).Unix()
		ls.PushInteger(t)
	}
	return 1
}

// lua-5.3.4/src/loslib.c#getfield()
func _getField(ls api.LuaState, key string, dft int64) int {
	t := ls.GetField(-1, key) /* get field and its type */
	res, isNum := ls.ToIntegerX(-1)
	if !isNum { /* field is not an integer? */
		if t != api.LUA_TNIL { /* some other value? */
			return ls.Error2("field '%s' is not an integer", key)
		} else if dft < 0 { /* absent field; no default? */
			return ls.Error2("field '%s' missing in date table", key)
		}
		res = dft
	}
	ls.Pop(1)
	return int(res)
}

func osDate(ls api.LuaState) int {
	format := ls.OptString(1, "%c")
	var t time.Time
	if ls.IsInteger(2) {
		t = time.Unix(ls.ToInteger(2), 0)
	} else {
		t = time.Now()
	}

	if format != "" && format[0] == '!' { /* UTC? */
		format = format[1:] /* skip '!' */
		t = t.In(time.UTC)
	}

	if format == "*t" {
		ls.CreateTable(0, 9) /* 9 = number of fields */
		_setField(ls, "sec", t.Second())
		_setField(ls, "min", t.Minute())
		_setField(ls, "hour", t.Hour())
		_setField(ls, "day", t.Day())
		_setField(ls, "month", int(t.Month()))
		_setField(ls, "year", t.Year())
		_setField(ls, "wday", int(t.Weekday())+1)
		_setField(ls, "yday", t.YearDay())
	} else if format == "%c" {
		ls.PushString(t.Format(time.ANSIC))
	} else {
		ls.PushString(format) // TODO
	}

	return 1
}

func _setField(ls api.LuaState, key string, value int) {
	ls.PushInteger(int64(value))
	ls.SetField(-2, key)
}

func osRemove(ls api.LuaState) int {
	filename := ls.CheckString(1)
	if err := os.Remove(filename); err != nil {
		ls.PushNil()
		ls.PushString(err.Error())
		return 2
	} else {
		ls.PushBoolean(true)
		return 1
	}
}

func osRename(ls api.LuaState) int {
	oldName := ls.CheckString(1)
	newName := ls.CheckString(2)
	if err := os.Rename(oldName, newName); err != nil {
		ls.PushNil()
		ls.PushString(err.Error())
		return 2
	} else {
		ls.PushBoolean(true)
		return 1
	}
}

func osTmpName(ls api.LuaState) int {
	panic("todo: osTmpName!")
}

func osGetEnv(ls api.LuaState) int {
	key := ls.CheckString(1)
	if env := os.Getenv(key); env != "" {
		ls.PushString(env)
	} else {
		ls.PushNil()
	}
	return 1
}

func osExecute(ls api.LuaState) int {
	panic("todo: osExecute!")
}

func osExit(ls api.LuaState) int {
	if ls.IsBoolean(1) {
		if ls.ToBoolean(1) {
			os.Exit(0)
		} else {
			os.Exit(1) // todo
		}
	} else {
		code := ls.OptInteger(1, 1)
		os.Exit(int(code))
	}
	if ls.ToBoolean(2) {
		//ls.Close()
	}
	return 0
}

func osSetLocale(ls api.LuaState) int {
	panic("todo: osSetLocale!")
}
