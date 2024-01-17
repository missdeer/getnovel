#!/bin/bash

OS=$(uname -s)

case "$OS" in
"Darwin")
	PLAT="macosx"
	;;
"Linux")
	PLAT="linux"
	;;
"MINGW"* | "MSYS_NT"*)
	PLAT="mingw"
	;;
"FreeBSD" | "NetBSD" | "OpenBSD" | "DragonFly")
	PLAT="freebsd"
	;;
esac
echo $PLAT

find . -name 'lua*' -type d | while read dir; do
	echo $dir $PLAT
	cd $dir
	make clean
	make $PLAT
	cd ..
done
