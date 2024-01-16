package main

import (
	"fmt"

	"github.com/aarzilli/golua/lua"
)

// dumpAndLoadTest: dump a function chunk to bytecodes, then load bytecodes and call function
func dumpAndLoadTest(L *lua.State) {
	loadret := L.LoadString(`print("msg from dump_and_load_test")`)
	if loadret != 0 {
		panic(fmt.Sprintf("LoadString error: %v", loadret))
	}
	dumpret := L.Dump()
	if dumpret != 0 {
		panic(fmt.Sprintf("Dump error: %v", dumpret))
	}

	isstring := L.IsString(-1)
	if !isstring {
		panic("stack top not a string")
	}
	bytecodes := L.ToBytes(-1)
	loadret = L.Load(bytecodes, "chunk_from_dump_and_load_test")
	if loadret != 0 {
		panic(fmt.Sprintf("Load error: %v", loadret))
	}
	err := L.Call(0, 0)
	if err != nil {
		panic(fmt.Sprintf("Call error: %v", err))
	}
}

func main() {
	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()

	dumpAndLoadTest(L)
}
