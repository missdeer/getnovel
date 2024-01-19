#!/bin/bash

OS=$(uname -s)
CoreCount=1

case "$OS" in
"Darwin")
	PLAT="macosx"
	CoreCount=$(getconf _NPROCESSORS_ONLN)
	;;
"Linux")
	PLAT="linux"
	CoreCount=$(getconf _NPROCESSORS_ONLN)
	;;
"MINGW"* | "MSYS_NT"*)
	PLAT="mingw"
	CoreCount=$(nproc)
	;;
"FreeBSD" | "NetBSD" | "OpenBSD" | "DragonFly")
	PLAT="freebsd"
	CoreCount=$(getconf _NPROCESSORS_ONLN)
	;;
esac
echo $PLAT

find . -name 'lua*' -type d | while read dir; do
	echo $dir $PLAT
	cd $dir
	make clean
	if [ "$OS" == "Darwin" ]; then
		make MYCFLAGS="-arch x86_64 -arch arm64" MYLDFLAGS="-arch x86_64 -arch arm64" $PLAT -j $CoreCount
	else
		make $PLAT -j $CoreCount
	fi
	cd ..
done

if [ ! -d luajit ]; then
	git clone --depth 1 https://github.com/LuaJIT/LuaJIT.git luajit
fi
cd luajit
if [ "$OS" == "Darwin" ]; then
	env MACOSX_DEPLOYMENT_TARGET=12.0 make clean
	env MACOSX_DEPLOYMENT_TARGET=12.0 CFLAGS="-arch x86_64" LDFLAGS="-arch x86_64" make -j $CoreCount BUILDMODE=static
	arch=`uname -m`
	if [ "$arch" == "arm64" ]; then
		mv src/libluajit.a ./libluajit-amd64.a
		env MACOSX_DEPLOYMENT_TARGET=12.0 make clean
		env MACOSX_DEPLOYMENT_TARGET=12.0 CFLAGS="-arch arm64" LDFLAGS="-arch arm64" make -j $CoreCount BUILDMODE=static
		mv src/libluajit.a ./libluajit-arm64.a
		lipo -create -output libluajit.a libluajit-arm64.a libluajit-amd64.a
	else
		mv src/libluajit.a ./libluajit.a
	fi
else
	make -j $CoreCount BUILDMODE=static
	mv src/*.a ./libluajit.a
fi
cd ..
