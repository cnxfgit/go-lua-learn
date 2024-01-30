print(123)

local function foo()
    print("foo")
end


local function bar()
    print("bar")
end

return {
    foo = foo,
    bar = bar
}
