print(_VERSION) --> Lua 5.3
print(_G._VERSION) --> Lua 5.3
print(_G) --> 0x7fce7e402710
print(_G._G) --> 0x7fce7e402710
print(print) --> 0x1073e2b90
print(_G.print) --> 0x1073e2b90

print(os.clock())