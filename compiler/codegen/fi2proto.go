package codegen

import "luago/binchunk"

func toProto(fi *funcInfo) *binchunk.Prototype {
	proto := &binchunk.Prototype{
		LineDefined:     uint32(fi.line),
		LastLineDefined: uint32(fi.lastLine),
		NumParams:       byte(fi.numParams),
		MaxStackSize:    byte(fi.maxRegs),
		Code:            fi.insts,
		Constants:       getConstants(fi),
		Upvalues:        getUpvalues(fi),
		Protos:          toProtos(fi.subFuncs),
		LineInfo:        fi.lineNums,
		LocVars:         getLocVars(fi),
		UpvalueNames:    getUpvalueNames(fi),
	}

	if fi.line == 0 {
		proto.LastLineDefined = 0
	}
	if proto.MaxStackSize < 2 {
		proto.MaxStackSize = 2 // todo
	}
	if fi.isVararg {
		proto.IsVararg = 1 // todo
	}

	return proto
}

func getLocVars(fi *funcInfo) []binchunk.LocVar {
	locVars := make([]binchunk.LocVar, len(fi.locVars))
	for i, locVar := range fi.locVars {
		locVars[i] = binchunk.LocVar{
			VarName: locVar.name,
			StartPC: uint32(locVar.startPC),
			EndPC:   uint32(locVar.endPC),
		}
	}
	return locVars
}

func getUpvalues(fi *funcInfo) []binchunk.Upvalue {
	upvals := make([]binchunk.Upvalue, len(fi.upvalues))
	for _, uv := range fi.upvalues {
		if uv.locVarSlot >= 0 { // instack
			upvals[uv.index] = binchunk.Upvalue{
				Instack: 1,
				Idx:     byte(uv.locVarSlot),
			}
		} else {
			upvals[uv.index] = binchunk.Upvalue{
				Instack: 0,
				Idx:     byte(uv.upvalIndex),
			}
		}
	}
	return upvals
}

func toProtos(fis []*funcInfo) []*binchunk.Prototype {
	protos := make([]*binchunk.Prototype, len(fis))
	for i, fi := range fis {
		protos[i] = toProto(fi)
	}
	return protos
}

func getConstants(fi *funcInfo) []interface{} {
	consts := make([]interface{}, len(fi.constants))
	for k, idx := range fi.constants {
		consts[idx] = k
	}
	return consts
}

func getUpvalueNames(fi *funcInfo) []string {
	names := make([]string, len(fi.upvalues))
	for name, uv := range fi.upvalues {
		names[uv.index] = name
	}
	return names
}
