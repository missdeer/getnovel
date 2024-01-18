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
