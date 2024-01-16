// This package provides access to the excellent lua language interpreter from go code.
//
// Access to most of the functions in lua.h and lauxlib.h is provided as well as additional convenience functions to publish Go objects and functions to lua code.
//
// The documentation of this package is no substitute for the official lua documentation and in many instances methods are described only with the name of their C equivalent
package lua

/*
#cgo !lua52,!lua53,!lua54 CFLAGS: -I ${SRCDIR}/lua51
#cgo lua52 CFLAGS: -I ${SRCDIR}/lua52
#cgo lua53 CFLAGS: -I ${SRCDIR}/lua53
#cgo lua54 CFLAGS: -I ${SRCDIR}/lua54
#cgo llua LDFLAGS: -llua

#cgo luaa LDFLAGS: -llua -lm -ldl
#cgo luajit LDFLAGS: -lluajit-5.1
#cgo lluadash5.1 LDFLAGS: -llua-5.1

#cgo linux,!lua52,!lua53,!lua54,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua5.1
#cgo linux,lua52,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua5.2
#cgo linux,lua53,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua5.3
#cgo linux,lua54,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua5.4 -lm

#cgo darwin,!lua52,!lua53,!lua54,!llua,!luaa,!luajit,!lluadash5.1 pkg-config: lua5.1
#cgo darwin,lua52,!llua,!luaa,!luajit,!lluadash5.1 pkg-config: lua5.2
#cgo darwin,lua53,!llua,!luaa,!luajit,!lluadash5.1 pkg-config: lua5.3
#cgo darwin,lua54,!llua,!luaa,!luajit,!lluadash5.1 pkg-config: lua5.4 m

#cgo freebsd,!lua52,!lua53,!lua54,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua-5.1
#cgo freebsd,lua52,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua-5.2
#cgo freebsd,lua53,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua-5.3
#cgo freebsd,lua54,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua-5.4 -lm

#cgo windows,!lua52,!lua53,!lua54,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -L${SRCDIR} -llua -lmingwex -lmingw32
#cgo windows,lua52,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua52
#cgo windows,lua53,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua53
#cgo windows,lua54,!llua,!luaa,!luajit,!lluadash5.1 LDFLAGS: -llua54

#include <lua.h>
#include <stdlib.h>

#include "golua.h"

*/
import "C"
import (
	"fmt"
	"unsafe"
)

type LuaStackEntry struct {
	Name        string
	Source      string
	ShortSource string
	CurrentLine int
}

func newState(L *C.lua_State) *State {
	newstate := &State{L, 0, make([]interface{}, 0, 8), make([]uint, 0, 8), nil, nil}
	registerGoState(newstate)
	C.clua_setgostate(L, C.size_t(newstate.Index))
	C.clua_initstate(L)
	return newstate
}

func (L *State) addFreeIndex(i uint) {
	freelen := len(L.freeIndices)
	//reallocate if necessary
	if freelen+1 > cap(L.freeIndices) {
		newSlice := make([]uint, freelen, cap(L.freeIndices)*2)
		copy(newSlice, L.freeIndices)
		L.freeIndices = newSlice
	}
	//reslice
	L.freeIndices = L.freeIndices[0 : freelen+1]
	L.freeIndices[freelen] = i
}

func (L *State) getFreeIndex() (index uint, ok bool) {
	freelen := len(L.freeIndices)
	//if there exist entries in the freelist
	if freelen > 0 {
		i := L.freeIndices[freelen-1] //get index
		//fmt.Printf("Free indices before: %v\n", L.freeIndices)
		L.freeIndices = L.freeIndices[0 : freelen-1] //'pop' index from list
		//fmt.Printf("Free indices after: %v\n", L.freeIndices)
		return i, true
	}
	return 0, false
}

//returns the registered function id
func (L *State) register(f interface{}) uint {
	//fmt.Printf("Registering %v\n")
	index, ok := L.getFreeIndex()
	//fmt.Printf("\tfreeindex: index = %v, ok = %v\n", index, ok)
	//if not ok, then we need to add new index by extending the slice
	if !ok {
		index = uint(len(L.registry))
		//reallocate backing array if necessary
		if index+1 > uint(cap(L.registry)) {
			newcap := cap(L.registry) * 2
			if index+1 > uint(newcap) {
				newcap = int(index + 1)
			}
			newSlice := make([]interface{}, index, newcap)
			copy(newSlice, L.registry)
			L.registry = newSlice
		}
		//reslice
		L.registry = L.registry[0 : index+1]
	}
	//fmt.Printf("\tregistering %d %v\n", index, f)
	L.registry[index] = f
	return index
}

func (L *State) unregister(fid uint) {
	//fmt.Printf("Unregistering %d (len: %d, value: %v)\n", fid, len(L.registry), L.registry[fid])
	if (fid < uint(len(L.registry))) && (L.registry[fid] != nil) {
		L.registry[fid] = nil
		L.addFreeIndex(fid)
	}
}

// Like lua_pushcfunction pushes onto the stack a go function as user data
func (L *State) PushGoFunction(f LuaGoFunction) {
	fid := L.register(f)
	C.clua_pushgofunction(L.s, C.uint(fid))
}

// PushGoClosure pushes a lua.LuaGoFunction to the stack wrapped in a Closure.
// this permits the go function to reflect lua type 'function' when checking with type()
// this implements behaviour akin to lua_pushcfunction() in lua C API.
func (L *State) PushGoClosure(f LuaGoFunction) {
	L.PushGoFunction(f)      // leaves Go function userdata on stack
	C.clua_pushcallback(L.s) // wraps the userdata object with a closure making it into a function
}

// Sets a metamethod to execute a go function
//
// The code:
//
// 	L.LGetMetaTable(tableName)
// 	L.SetMetaMethod(methodName, function)
//
// is the logical equivalent of:
//
// 	L.LGetMetaTable(tableName)
// 	L.PushGoFunction(function)
// 	L.SetField(-2, methodName)
//
// except this wouldn't work because pushing a go function results in user data not a cfunction
func (L *State) SetMetaMethod(methodName string, f LuaGoFunction) {
	L.PushGoFunction(f)      // leaves Go function userdata on stack
	C.clua_pushcallback(L.s) // wraps the userdata object with a closure making it into a function
	L.SetField(-2, methodName)
}

// Pushes a Go struct onto the stack as user data.
//
// The user data will be rigged so that lua code can access and change to public members of simple types directly
func (L *State) PushGoStruct(iface interface{}) {
	iid := L.register(iface)
	C.clua_pushgostruct(L.s, C.uint(iid))
}

// Push a pointer onto the stack as user data.
//
// This function doesn't save a reference to the interface, it is the responsibility of the caller of this function to insure that the interface outlasts the lifetime of the lua object that this function creates.
func (L *State) PushLightUserdata(ud *interface{}) {
	//push
	C.lua_pushlightuserdata(L.s, unsafe.Pointer(ud))
}

// Creates a new user data object of specified size and returns it
func (L *State) NewUserdata(size uintptr) unsafe.Pointer {
	return unsafe.Pointer(C.lua_newuserdata(L.s, C.size_t(size)))
}

// Sets the AtPanic function, returns the old one
//
// BUG(everyone_involved): passing nil causes serious problems
func (L *State) AtPanic(panicf LuaGoFunction) (oldpanicf LuaGoFunction) {
	fid := uint(0)
	if panicf != nil {
		fid = L.register(panicf)
	}
	oldres := interface{}(C.clua_atpanic(L.s, C.uint(fid)))
	switch i := oldres.(type) {
	case C.uint:
		f := L.registry[uint(i)].(LuaGoFunction)
		//free registry entry
		L.unregister(uint(i))
		return f
	case C.lua_CFunction:
		return func(L1 *State) int {
			return int(C.clua_callluacfunc(L1.s, i))
		}
	}
	//generally we only get here if the panicf got set to something like nil
	//potentially dangerous because we may silently fail
	return nil
}

func (L *State) callEx(nargs, nresults int, catch bool) (err error) {
	if catch {
		defer func() {
			if err2 := recover(); err2 != nil {
				if _, ok := err2.(error); ok {
					err = err2.(error)
				}
				return
			}
		}()
	}

	L.GetGlobal(C.GOLUA_DEFAULT_MSGHANDLER)
	// We must record where we put the error handler in the stack otherwise it will be impossible to remove after the pcall when nresults == LUA_MULTRET
	erridx := L.GetTop() - nargs - 1
	L.Insert(erridx)
	r := L.pcall(nargs, nresults, erridx)
	L.Remove(erridx)
	if r != 0 {
		err = &LuaError{r, L.ToString(-1), L.StackTrace()}
		if !catch {
			panic(err)
		}
	}
	return
}

// lua_call
func (L *State) Call(nargs, nresults int) (err error) {
	return L.callEx(nargs, nresults, true)
}

// Like lua_call but panics on errors
func (L *State) MustCall(nargs, nresults int) {
	L.callEx(nargs, nresults, false)
}

// lua_checkstack
func (L *State) CheckStack(extra int) bool {
	return C.lua_checkstack(L.s, C.int(extra)) != 0
}

// lua_close
func (L *State) Close() {
	C.lua_close(L.s)
	unregisterGoState(L)
}

// lua_concat
func (L *State) Concat(n int) {
	C.lua_concat(L.s, C.int(n))
}

// lua_createtable
func (L *State) CreateTable(narr int, nrec int) {
	C.lua_createtable(L.s, C.int(narr), C.int(nrec))
}

// lua_getfield
func (L *State) GetField(index int, k string) {
	Ck := C.CString(k)
	defer C.free(unsafe.Pointer(Ck))
	C.lua_getfield(L.s, C.int(index), Ck)
}

// lua_getmetatable
func (L *State) GetMetaTable(index int) bool {
	return C.lua_getmetatable(L.s, C.int(index)) != 0
}

// lua_gettable
func (L *State) GetTable(index int) { C.lua_gettable(L.s, C.int(index)) }

// lua_gettop
func (L *State) GetTop() int { return int(C.lua_gettop(L.s)) }

// Returns true if lua_type == LUA_TBOOLEAN
func (L *State) IsBoolean(index int) bool {
	return LuaValType(C.lua_type(L.s, C.int(index))) == LUA_TBOOLEAN
}

// Returns true if the value at index is a LuaGoFunction
func (L *State) IsGoFunction(index int) bool {
	return C.clua_isgofunction(L.s, C.int(index)) != 0
}

// Returns true if the value at index is user data pushed with PushGoStruct
func (L *State) IsGoStruct(index int) bool {
	return C.clua_isgostruct(L.s, C.int(index)) != 0
}

// Returns true if the value at index is user data pushed with PushGoFunction
func (L *State) IsFunction(index int) bool {
	return LuaValType(C.lua_type(L.s, C.int(index))) == LUA_TFUNCTION
}

// Returns true if the value at index is light user data
func (L *State) IsLightUserdata(index int) bool {
	return LuaValType(C.lua_type(L.s, C.int(index))) == LUA_TLIGHTUSERDATA
}

// lua_isnil
func (L *State) IsNil(index int) bool { return LuaValType(C.lua_type(L.s, C.int(index))) == LUA_TNIL }

// lua_isnone
func (L *State) IsNone(index int) bool { return LuaValType(C.lua_type(L.s, C.int(index))) == LUA_TNONE }

// lua_isnoneornil
func (L *State) IsNoneOrNil(index int) bool { return int(C.lua_type(L.s, C.int(index))) <= 0 }

// lua_isnumber
func (L *State) IsNumber(index int) bool { return C.lua_isnumber(L.s, C.int(index)) == 1 }

// lua_isstring
func (L *State) IsString(index int) bool { return C.lua_isstring(L.s, C.int(index)) == 1 }

// lua_istable
func (L *State) IsTable(index int) bool {
	return LuaValType(C.lua_type(L.s, C.int(index))) == LUA_TTABLE
}

// lua_isthread
func (L *State) IsThread(index int) bool {
	return LuaValType(C.lua_type(L.s, C.int(index))) == LUA_TTHREAD
}

// lua_isuserdata
func (L *State) IsUserdata(index int) bool { return C.lua_isuserdata(L.s, C.int(index)) == 1 }

// Creates a new lua interpreter state with the given allocation function
func NewStateAlloc(f Alloc) *State {
	ls := C.clua_newstate(unsafe.Pointer(&f))
	L := newState(ls)
	L.allocfn = &f
	return L
}

// lua_newtable
func (L *State) NewTable() {
	C.lua_createtable(L.s, 0, 0)
}

// lua_newthread
func (L *State) NewThread() *State {
	//TODO: call newState with result from C.lua_newthread and return it
	//TODO: should have same lists as parent
	//		but may complicate gc
	s := C.lua_newthread(L.s)
	return &State{s, 0, nil, nil, nil, nil}
}

// lua_next
func (L *State) Next(index int) int {
	return int(C.lua_next(L.s, C.int(index)))
}

// lua_objlen
// lua_pop
func (L *State) Pop(n int) {
	//Why is this implemented this way? I don't get it...
	//C.lua_pop(L.s, C.int(n));
	C.lua_settop(L.s, C.int(-n-1))
}

// lua_pushboolean
func (L *State) PushBoolean(b bool) {
	var bint int
	if b {
		bint = 1
	} else {
		bint = 0
	}
	C.lua_pushboolean(L.s, C.int(bint))
}

// lua_pushstring
func (L *State) PushString(str string) {
	Cstr := C.CString(str)
	defer C.free(unsafe.Pointer(Cstr))
	C.lua_pushlstring(L.s, Cstr, C.size_t(len(str)))
}

func (L *State) PushBytes(b []byte) {
	C.lua_pushlstring(L.s, (*C.char)(unsafe.Pointer(&b[0])), C.size_t(len(b)))
}

// lua_pushinteger
func (L *State) PushInteger(n int64) {
	C.lua_pushinteger(L.s, C.lua_Integer(n))
}

// lua_pushnil
func (L *State) PushNil() {
	C.lua_pushnil(L.s)
}

// lua_pushnumber
func (L *State) PushNumber(n float64) {
	C.lua_pushnumber(L.s, C.lua_Number(n))
}

// lua_pushthread
func (L *State) PushThread() (isMain bool) {
	return C.lua_pushthread(L.s) != 0
}

// lua_pushvalue
func (L *State) PushValue(index int) {
	C.lua_pushvalue(L.s, C.int(index))
}

// lua_rawequal
func (L *State) RawEqual(index1 int, index2 int) bool {
	return C.lua_rawequal(L.s, C.int(index1), C.int(index2)) != 0
}

// lua_rawget
func (L *State) RawGet(index int) {
	C.lua_rawget(L.s, C.int(index))
}

// lua_rawset
func (L *State) RawSet(index int) {
	C.lua_rawset(L.s, C.int(index))
}

// Registers a Go function as a global variable
func (L *State) Register(name string, f LuaGoFunction) {
	L.PushGoFunction(f)
	L.SetGlobal(name)
}

// lua_setallocf
func (L *State) SetAllocf(f Alloc) {
	L.allocfn = &f
	C.clua_setallocf(L.s, unsafe.Pointer(L.allocfn))
}

// lua_setfield
func (L *State) SetField(index int, k string) {
	Ck := C.CString(k)
	defer C.free(unsafe.Pointer(Ck))
	C.lua_setfield(L.s, C.int(index), Ck)
}

// lua_setmetatable
func (L *State) SetMetaTable(index int) {
	C.lua_setmetatable(L.s, C.int(index))
}

// lua_settable
func (L *State) SetTable(index int) {
	C.lua_settable(L.s, C.int(index))
}

// lua_settop
func (L *State) SetTop(index int) {
	C.lua_settop(L.s, C.int(index))
}

// lua_status
func (L *State) Status() int {
	return int(C.lua_status(L.s))
}

// lua_toboolean
func (L *State) ToBoolean(index int) bool {
	return C.lua_toboolean(L.s, C.int(index)) != 0
}

// Returns the value at index as a Go function (it must be something pushed with PushGoFunction)
func (L *State) ToGoFunction(index int) (f LuaGoFunction) {
	if !L.IsGoFunction(index) {
		return nil
	}
	fid := C.clua_togofunction(L.s, C.int(index))
	if fid < 0 {
		return nil
	}
	return L.registry[fid].(LuaGoFunction)
}

// Returns the value at index as a Go Struct (it must be something pushed with PushGoStruct)
func (L *State) ToGoStruct(index int) (f interface{}) {
	if !L.IsGoStruct(index) {
		return nil
	}
	fid := C.clua_togostruct(L.s, C.int(index))
	if fid < 0 {
		return nil
	}
	return L.registry[fid]
}

// lua_tostring
func (L *State) ToString(index int) string {
	var size C.size_t
	r := C.lua_tolstring(L.s, C.int(index), &size)
	return C.GoStringN(r, C.int(size))
}

func (L *State) ToBytes(index int) []byte {
	var size C.size_t
	b := C.lua_tolstring(L.s, C.int(index), &size)
	return C.GoBytes(unsafe.Pointer(b), C.int(size))
}

// lua_topointer
func (L *State) ToPointer(index int) uintptr {
	return uintptr(C.lua_topointer(L.s, C.int(index)))
}

// lua_tothread
func (L *State) ToThread(index int) *State {
	//TODO: find a way to link lua_State* to existing *State, return that
	return &State{}
}

// lua_touserdata
func (L *State) ToUserdata(index int) unsafe.Pointer {
	return unsafe.Pointer(C.lua_touserdata(L.s, C.int(index)))
}

// lua_type
func (L *State) Type(index int) LuaValType {
	return LuaValType(C.lua_type(L.s, C.int(index)))
}

// lua_typename
func (L *State) Typename(tp int) string {
	return C.GoString(C.lua_typename(L.s, C.int(tp)))
}

// lua_xmove
func XMove(from *State, to *State, n int) {
	C.lua_xmove(from.s, to.s, C.int(n))
}

// Restricted library opens

// Calls luaopen_base
func (L *State) OpenBase() {
	C.clua_openbase(L.s)
}

// Calls luaopen_io
func (L *State) OpenIO() {
	C.clua_openio(L.s)
}

// Calls luaopen_math
func (L *State) OpenMath() {
	C.clua_openmath(L.s)
}

// Calls luaopen_package
func (L *State) OpenPackage() {
	C.clua_openpackage(L.s)
}

// Calls luaopen_string
func (L *State) OpenString() {
	C.clua_openstring(L.s)
}

// Calls luaopen_table
func (L *State) OpenTable() {
	C.clua_opentable(L.s)
}

// Calls luaopen_os
func (L *State) OpenOS() {
	C.clua_openos(L.s)
}

// Sets the lua hook (lua_sethook).
// This and SetExecutionLimit are mutual exclusive
func (L *State) SetHook(f HookFunction, instrNumber int) {
	L.hookFn = f
	C.clua_sethook(L.s, C.int(instrNumber))
}

// Sets the maximum number of operations to execute at instrNumber, after this the execution ends
// This and SetHook are mutual exclusive
func (L *State) SetExecutionLimit(instrNumber int) {
	L.SetHook(func(l *State) {
		l.RaiseError(ExecutionQuantumExceeded)
	}, instrNumber)
}

// Returns the current stack trace
func (L *State) StackTrace() []LuaStackEntry {
	r := []LuaStackEntry{}
	var d C.lua_Debug
	Sln := C.CString("Sln")
	defer C.free(unsafe.Pointer(Sln))

	for depth := 0; C.lua_getstack(L.s, C.int(depth), &d) > 0; depth++ {
		C.lua_getinfo(L.s, Sln, &d)
		ssb := make([]byte, C.LUA_IDSIZE)
		for i := 0; i < C.LUA_IDSIZE; i++ {
			ssb[i] = byte(d.short_src[i])
			if ssb[i] == 0 {
				ssb = ssb[:i]
				break
			}
		}
		ss := string(ssb)

		r = append(r, LuaStackEntry{C.GoString(d.name), C.GoString(d.source), ss, int(d.currentline)})
	}

	return r
}

func (L *State) RaiseError(msg string) {
	st := L.StackTrace()
	prefix := ""
	if len(st) >= 2 {
		prefix = fmt.Sprintf("%s:%d: ", st[1].ShortSource, st[1].CurrentLine)
	}
	panic(&LuaError{0, prefix + msg, st})
}

func (L *State) NewError(msg string) *LuaError {
	return &LuaError{0, msg, L.StackTrace()}
}

func (L *State) GetState() *C.lua_State {
	return L.s
}
