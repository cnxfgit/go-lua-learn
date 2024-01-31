-- 类型测试

local a, b, c
c = false       -- boolean
c = { 1, 2, 3 } -- table
c = "hello"     -- string
a = 3.14        -- number
b = a

print(type(nil))                     --> nil
print(type(true))                    --> boolean
print(type(3.14))                    --> number
print(type("Hello world"))           --> string
print(type({}))                      --> table
print(type(print))                   --> function
print(type(coroutine.create(print))) --> thread

-- 计算
print(5 // 3)            --> 1
print(-5 // 3)           --> -2
print(5 // -3.0)         --> -2.0
print(-5.0 // -3.0)      --> 1.0

print(5 % 3)             --> 2
print(-5 % 3)            --> 1
print(5 % -3.0)          --> -1.0
print(-5.0 % -3.0)       --> -2.0

print(100 / 10 / 2)      -- (100/10)/2 == 5.0
print(4 ^ 3 ^ 2)         -- 4^(3^2) == 262144.0

print(-1 >> 63)          --> 1
print(2 << -1)           --> 1
print("1" << 1.0)        --> 2

print(#"hello")          --> 5
print(#{ 7, 8, 9 })      --> 3

print("a" .. "b" .. "c") --> abc
print(1 .. 2 .. 3)       --> 123

-- for循环

local sum = 0
for i = 1, 100 do
    if i % 2 == 0 then
        sum = sum + i
    end
end
print("sum=" .. sum)

-- 表

local t = {} -- empty table
local p = { x = 100, y = 200 }

t[false] = nil; assert(t[false] == nil)
t["pi"] = 3.14; assert(t["pi"] == 3.14)
t[t] = "table"; assert(t[t] == "table")
t[10] = assert; assert(t[10] == assert)

local arr = { "a", "b", "c", nil, "e" }
assert(arr[1] == "a")
assert(arr[2] == "b")
assert(arr[3] == "c")
assert(arr[4] == nil)
assert(arr[5] == "e")

local seq = { "a", "b", "c", "d", "e" }
assert(#seq == 5)

local t = {}
t[6] = "foo"; assert(t[6.0] == "foo")
t[7.0] = "foo"; assert(t[7] == "foo")
t["8"] = "bar"; assert(t[8] == nil)
t[9] = "bar"; assert(t["9"] == nil)

local t = { 1, 2, 3 }
t[5] = 5
assert(#t == 3)
t[4] = 4
assert(#t == 5)


-- 函数
local function f(a, b, c)
    print(a, b, c)
end

f()              -->	nil	nil	nil
f(1, 2)          -->	1	2	nil
f(1, 2, 3, 4, 5) -->	1	2	3


local function f(a, ...)
    local b, c = ...
    local t = { a, ... }
    print(a, b, c, #t, ...)
end

f()           -->	nil	nil	nil	0
f(1, 2)       -->	1	2	nil	2	2
f(1, 2, 3, 4) -->	1	2	3	4	2	3	4


local function f()
    return 1, 2, 3
end

f()
a, b = f(); assert(a == 1 and b == 2)
a, b, c = f(); assert(a == 1 and b == 2 and c == 3)
a, b, c, d = f(); assert(a == 1 and b == 2 and c == 3 and d == nil)


function f()
    return 1, 2, 3
end

a, b = (f()); assert(a == 1 and b == nil)
a, b, c = 5, f(), 5; assert(a == 5 and b == 1 and c == 5)

function f() return 3, 2, 1 end

local function g() return 4, f() end

local function h(a, b, c, d) print(a, b, c, d) end

h(4, f())                       -->	4	3	2	1
h(g())                          -->	4	3	2	1
print(table.unpack({ 4, f() })) -->	4	3	2	1

-- 元表

print(getmetatable("foo")) --> table: 0x7f8aab4050c0
print(getmetatable("bar")) --> table: 0x7f8aab4050c0
print(getmetatable(nil))   --> nil
print(getmetatable(false)) --> nil
print(getmetatable(100))   --> nil
print(getmetatable({}))    --> nil
print(getmetatable(print)) --> nil

t = {}
local mt = {}
setmetatable(t, mt)
print(getmetatable(t) == mt)   --> true
print(getmetatable(200) == mt) --> true

local function vector(x, y)
    local v = { x = x, y = y }
    setmetatable(v, mt)
    return v
end

mt = {}
mt.__add = function(v1, v2)
    return vector(v1.x + v2.x, v1.y + v2.y)
end

local v1 = vector(1, 2)
local v2 = vector(3, 5)
local v3 = v1 + v2
print(v3.x, v3.y) --> 4	7

-- 迭代器

local function ipairs(t)
    local i = 0
    return function()
        i = i + 1
        if t[i] == nil then
            return nil, nil
        else
            return i, t[i]
        end
    end
end

t = { 10, 20, 30 }
local iter = ipairs(t)
while true do
    local i, v = iter()
    if i == nil then
        break
    end
    print(i, v)
end

t = { 10, 20, 30 }
for i, v in ipairs(t) do
    print(i, v)
end


function pairs(t)
    local k, v
    return function()
        k, v = next(t, k)
        return k, v
    end
end

t = { a = 10, b = 20, c = 30 }
for k, v in pairs(t) do
    print(k, v)
end


t = { a = 10, b = 20, c = 30 }
for k, v in next, t, nil do
    print(k, v)
end


function pairs(t)
    return next, t, nil
end

t = { a = 10, b = 20, c = 30 }
for k, v in pairs(t) do
    print(k, v)
end


local function inext(t, i)
    local nextIdx = i + 1
    local nextVal = t[nextIdx]
    if nextVal == nil then
        return nil
    else
        return nextIdx, nextVal
    end
end

t = { 10, 20, 30 }
for i, v in inext, t, 0 do
    print(i, v)
end

-- 异常

local function error(lock)
    local ok, err = pcall(function()
        sleep(2000)
    end)
    print("PCall: " .. err)
    print(ok)
end
error()

-- 字符串
print("hello") -- short comment
print("world") --> another short comment
print() --[[ long comment ]]
--[===[
  another
  long comment
]===]

print("hello, \z
       world!") --> hello, world!

a = 'alo\n123"'
print(a)
a = "alo\n123\""
print(a)
a = '\97lo\10\04923"'
print(a)
a = [[alo
123"]]
print(a)
a = [==[
alo
123"]==]
print(a)

-- 一元 二元表达式

local a = true and false or false or not true
local b = ((1 | 2) & 3) >> 1 << 1
local c = (3 + 2 - 1) * (5 % 2) // 2 / 2 ^ 2
local d = not not not not not false
local e = - - - - -1
local f = ~ ~ ~ ~ ~1
print(a, b, c, d, e, f)

-- 全局变量
print(_VERSION)    --> Lua 5.3
print(_G._VERSION) --> Lua 5.3
print(_G)          --> 0x7fce7e402710
print(_G._G)       --> 0x7fce7e402710
print(print)       --> 0x1073e2b90
print(_G.print)    --> 0x1073e2b90

print(select(1, "a", "b", "c"))		--> a    b    c
print(select(2, "a", "b", "c"))		--> b    c
print(select(3, "a", "b", "c"))		--> c
print(select(-1, "a", "b", "c"))	--> c
print(select("#", "a", "b", "c"))	--> 3

-- 作用域

local function f()
    local a, b = 1, 2; print(a, b)   -->	1	2
    local a, b = 3, 4; print(a, b)   -->	3	4
    do
        print(a, b)                  -->	3	4
        local a, b = 5, 6; print(a, b) -->	5	6
    end
    print(a, b)                      -->	3	4
end

f()

-- 标准库
print(math.type(100))          --> integer
print(math.type(3.14))         --> float
print(math.type("100"))        --> nil
print(math.tointeger(100.0))   --> 100
print(math.tointeger("100.0")) --> 100
print(math.tointeger(3.14))    --> nil


t = table.pack(1, 2, 3, 4, 5); print(table.unpack(t)) --> 1 2 3 4 5
table.move(t, 4, 5, 1);        print(table.unpack(t)) --> 4 5 3 4 5
table.insert(t, 3, 2);         print(table.unpack(t)) --> 4 5 2 3 4 5
table.remove(t, 2);            print(table.unpack(t)) --> 4 2 3 4 5
table.sort(t);                 print(table.unpack(t)) --> 2 3 4 4 5
print(table.concat(t, ","))                           --> 2,3,4,4,5


print(string.len("abc"))            --> 3
print(string.rep("a", 3, ","))      --> a,a,a
print(string.reverse("abc"))        --> cba
print(string.lower("ABC"))          --> abc
print(string.upper("abc"))          --> ABC
print(string.sub("abcdefg", 3, 5))  --> cde
print(string.byte("abcdefg", 3, 5)) --> 99 100 101
print(string.char(99, 100, 101))    --> cde

local s = "aBc"
print(s:len())       --> 2
print(s:rep(3, ",")) --> aBc,aBc,aBc
print(s:reverse())   --> cBa
print(s:upper())     --> ABC
print(s:lower())     --> abc
print(s:sub(1, 2))   --> aB
print(s:byte(1, 2))  --> 97 66

print(string.len("你好，世界！"))			--> 18
print(utf8.len("你好，世界！"))			--> 6
print("\u{4f60}\u{597d}")				--> 你好
print(utf8.offset("你好，世界！", 2))		--> 4
print(utf8.offset("你好，世界！", 5))		--> 13
print(utf8.codepoint("你好，世界！", 4))	--> 22909
print(utf8.codepoint("你好，世界！", 13))	--> 30028
for p, c in utf8.codes("你好，世界！") do
  print(p, c)
end

print(os.time()) --> 1518320879
print(os.time{year=2018, month=2, day=14,
  hour=12, min=30, sec=30}) --> 1518582630

print(os.date()) --> Sun Feb 11 11:49:28 2018
local t = os.date("*t", 1518582630)
print(t.year)  --> 2018
print(t.month) --> 02
print(t.day)   --> 14
print(t.hour)  --> 12
print(t.min)   --> 30
print(t.sec)   --> 30

-- 模块化

local mymod = require("mymod")

mymod.foo()
mymod.bar()

-- 协程

local function permgen(a, n)
    n = n or #a    -- default for 'n' is size of 'a'
    if n <= 1 then -- nothing to change?
        coroutine.yield(a)
    else
        for i = 1, n do
            -- put i-th element as the last one
            a[n], a[i] = a[i], a[n]
            -- generate all permutations of the other elements
            permgen(a, n - 1)
            -- restore i-th element
            a[n], a[i] = a[i], a[n]
        end
    end
end

local function permutations(a)
    local co = coroutine.create(function() permgen(a) end)
    return function() -- iterator
        local code, res = coroutine.resume(co)
        return res
    end
end

for p in permutations { "a", "b", "c" } do
    print(table.concat(p, ","))
end

-- 闭包

local Counter = function()
    local count = 0
    return function()
        count = count + 1
        return count
    end
end

local counter = Counter()
print(counter())
print(counter())
print(counter())


local counter1 = Counter()
print(counter1())
print(counter1())
print(counter1())

-- 函数语法糖

local t = {}
function t:test()
    print("t:test")
end

t:test()
