package state

func (ls *luaState) Len(idx int) {
	val := ls.stack.get(idx)
	if s, ok := val.(string); ok {
		ls.stack.push(int64(len(s)))
	} else if result, ok := callMetamethod(val, val, "__len", ls); ok {
		ls.stack.push(result)
	} else if t, ok := val.(*luaTable); ok {
		ls.stack.push(int64(t.len()))
	} else {
		panic("length error!")
	}
}

func (ls *luaState) Concat(n int) {
	if n == 0 {
		ls.stack.push("")
	} else if n >= 2 {
		for i := 1; i < n; i++ {
			if ls.IsString(-1) && ls.IsString(-2) {
				s2 := ls.ToString(-1)
				s1 := ls.ToString(-2)
				ls.stack.pop()
				ls.stack.pop()
				ls.stack.push(s1 + s2)
				continue
			}
			b := ls.stack.pop()
			a := ls.stack.pop()
			if result, ok := callMetamethod(a, b, "__concat", ls); ok {
				ls.stack.push(result)
				continue
			}
			panic("concatenation error!")
		}
	}

	// n == 1, do nothing
}

func (ls *luaState) Next(idx int) bool {
	val := ls.stack.get(idx)
	if t, ok := val.(*luaTable); ok {
		key := ls.stack.pop()
		if nextKey := t.nextKey(key); nextKey != nil {
			ls.stack.push(nextKey)
			ls.stack.push(t.get(nextKey))
			return true
		}
		return false
	}
	panic("table expected")
}

func (ls *luaState) Error() int {
	err := ls.stack.pop()
	panic(err)
}