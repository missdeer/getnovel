package main

import (
	"github.com/aarzilli/golua/lua"
	"github.com/missdeer/golib/ic"
	"golang.org/x/net/html/charset"
)

func ConvertEncoding(L *lua.State) int {
	fromEncoding := L.CheckString(1)
	toEncoding := L.CheckString(2)
	fromStr := L.CheckString(3)
	toStr := ic.ConvertString(fromEncoding, toEncoding, fromStr)
	L.PushString(toStr)
	return 1
}

func DetectContentCharset(L *lua.State) int {
	dataStr := L.CheckString(1)
	data := []byte(dataStr)
	if _, name, ok := charset.DetermineEncoding(data, ""); ok {
		L.PushString(name)
		return 1
	}
	L.PushString("utf-8")
	return 1
}

func registerLuaAPIs(L *lua.State) {
	// add string.convert(from, to, str) method
	L.GetGlobal("string")
	L.PushGoFunction(ConvertEncoding)
	L.SetField(-2, "convert")
	L.Pop(1)

	// add string.charsetdet(str) method
	L.GetGlobal("string")
	L.PushGoFunction(DetectContentCharset)
	L.SetField(-2, "charsetdet")
	L.Pop(1)
}
