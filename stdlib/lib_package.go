package stdlib

import (
	"luago/api"
	"os"
	"strings"
)

const LUA_LOADED_TABLE = "_LOADED"
const LUA_PRELOAD_TABLE = "_PRELOAD"

const (
	LUA_DIRSEP    = string(os.PathSeparator)
	LUA_PATH_SEP  = ";"
	LUA_PATH_MARK = "?"
	LUA_EXEC_DIR  = "!"
	LUA_IGMARK    = "-"
)

var llFuncs = map[string]api.GoFunction{
	"require": pkgRequire,
}

var pkgFuncs = map[string]api.GoFunction{
	"searchpath": pkgSearchPath,
	/* placeholders */
	"preload":   nil,
	"cpath":     nil,
	"path":      nil,
	"searchers": nil,
	"loaded":    nil,
}

func OpenPackageLib(ls api.LuaState) int {
	ls.NewLib(pkgFuncs) /* create 'package' table */
	createSearchersTable(ls)
	/* set paths */
	ls.PushString("./?.lua;./?/init.lua")
	ls.SetField(-2, "path")
	/* store config information */
	ls.PushString(LUA_DIRSEP + "\n" + LUA_PATH_SEP + "\n" +
		LUA_PATH_MARK + "\n" + LUA_EXEC_DIR + "\n" + LUA_IGMARK + "\n")
	ls.SetField(-2, "config")
	/* set field 'loaded' */
	ls.GetSubTable(api.LUA_REGISTRYINDEX, LUA_LOADED_TABLE)
	ls.SetField(-2, "loaded")
	/* set field 'preload' */
	ls.GetSubTable(api.LUA_REGISTRYINDEX, LUA_PRELOAD_TABLE)
	ls.SetField(-2, "preload")
	ls.PushGlobalTable()
	ls.PushValue(-2)        /* set 'package' as upvalue for next lib */
	ls.SetFuncs(llFuncs, 1) /* open lib into global table */
	ls.Pop(1)               /* pop global table */
	return 1                /* return 'package' table */
}

func createSearchersTable(ls api.LuaState) {
	searchers := []api.GoFunction{
		preloadSearcher,
		luaSearcher,
	}
	/* create 'searchers' table */
	ls.CreateTable(len(searchers), 0)
	/* fill it with predefined searchers */
	for idx, searcher := range searchers {
		ls.PushValue(-2) /* set 'package' as upvalue for all searchers */
		ls.PushGoClosure(searcher, 1)
		ls.RawSetI(-2, int64(idx+1))
	}
	ls.SetField(-2, "searchers") /* put it in field 'searchers' */
}

func preloadSearcher(ls api.LuaState) int {
	name := ls.CheckString(1)
	ls.GetField(api.LUA_REGISTRYINDEX, "_PRELOAD")
	if ls.GetField(-1, name) == api.LUA_TNIL { /* not found? */
		ls.PushString("\n\tno field package.preload['" + name + "']")
	}
	return 1
}

func luaSearcher(ls api.LuaState) int {
	name := ls.CheckString(1)
	ls.GetField(api.LuaUpvalueIndex(1), "path")
	path, ok := ls.ToStringX(-1)
	if !ok {
		ls.Error2("'package.path' must be a string")
	}

	filename, errMsg := _searchPath(name, path, ".", LUA_DIRSEP)
	if errMsg != "" {
		ls.PushString(errMsg)
		return 1
	}

	if ls.LoadFile(filename) == api.LUA_OK { /* module loaded successfully? */
		ls.PushString(filename) /* will be 2nd argument to module */
		return 2                /* return open function and file name */
	} else {
		return ls.Error2("error loading module '%s' from file '%s':\n\t%s",
			ls.CheckString(1), filename, ls.CheckString(-1))
	}
}

func _searchPath(name, path, sep, dirSep string) (filename, errMsg string) {
	if sep != "" {
		name = strings.Replace(name, sep, dirSep, -1)
	}

	for _, filename := range strings.Split(path, LUA_PATH_SEP) {
		filename = strings.Replace(filename, LUA_PATH_MARK, name, -1)
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			return filename, ""
		}
		errMsg += "\n\tno file '" + filename + "'"
	}

	return "", errMsg
}

func pkgSearchPath(ls api.LuaState) int {
	name := ls.CheckString(1)
	path := ls.CheckString(2)
	sep := ls.OptString(3, ".")
	rep := ls.OptString(4, LUA_DIRSEP)
	if filename, errMsg := _searchPath(name, path, sep, rep); errMsg == "" {
		ls.PushString(filename)
		return 1
	} else {
		ls.PushNil()
		ls.PushString(errMsg)
		return 2
	}
}

func pkgRequire(ls api.LuaState) int {
	name := ls.CheckString(1)
	ls.SetTop(1) /* LOADED table will be at index 2 */
	ls.GetField(api.LUA_REGISTRYINDEX, LUA_LOADED_TABLE)
	ls.GetField(2, name)  /* LOADED[name] */
	if ls.ToBoolean(-1) { /* is it there? */
		return 1 /* package is already loaded */
	}
	/* else must load package */
	ls.Pop(1) /* remove 'getfield' result */
	_findLoader(ls, name)
	ls.PushString(name) /* pass name as argument to module loader */
	ls.Insert(-2)       /* name is 1st argument (before search data) */
	ls.Call(2, 1)       /* run loader to load module */
	if !ls.IsNil(-1) {  /* non-nil return? */
		ls.SetField(2, name) /* LOADED[name] = returned value */
	}
	if ls.GetField(2, name) == api.LUA_TNIL { /* module set no value? */
		ls.PushBoolean(true) /* use true as result */
		ls.PushValue(-1)     /* extra copy to be returned */
		ls.SetField(2, name) /* LOADED[name] = true */
	}
	return 1
}

func _findLoader(ls api.LuaState, name string) {
	/* push 'package.searchers' to index 3 in the stack */
	if ls.GetField(api.LuaUpvalueIndex(1), "searchers") != api.LUA_TTABLE {
		ls.Error2("'package.searchers' must be a table")
	}

	/* to build error message */
	errMsg := "module '" + name + "' not found:"

	/*  iterate over available searchers to find a loader */
	for i := int64(1); ; i++ {
		if ls.RawGetI(3, i) == api.LUA_TNIL { /* no more searchers? */
			ls.Pop(1)         /* remove nil */
			ls.Error2(errMsg) /* create error message */
		}

		ls.PushString(name)
		ls.Call(1, 2)          /* call it */
		if ls.IsFunction(-2) { /* did it find a loader? */
			return /* module loader found */
		} else if ls.IsString(-2) { /* searcher returned error message? */
			ls.Pop(1)                    /* remove extra return */
			errMsg += ls.CheckString(-1) /* concatenate error message */
		} else {
			ls.Pop(2) /* remove both returns */
		}
	}
}
