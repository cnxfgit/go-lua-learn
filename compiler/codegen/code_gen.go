package codegen

import (
	"luago/binchunk"
	"luago/compiler/ast"
)

func GenProto(chunk *ast.Block) *binchunk.Prototype {
	fd := &ast.FuncDefExp{
		LastLine: chunk.LastLine,
		IsVararg: true,
		Block:    chunk,
	}

	fi := newFuncInfo(nil, fd)
	fi.addLocVar("_ENV", 0)
	cgFuncDefExp(fi, fd, 0)
	return toProto(fi.subFuncs[0])
}
