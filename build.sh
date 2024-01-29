#!/bin/bash

OS=$(uname -s)

# Function to compile for a given Lua tag
compile_lua() {
	local lua_tag=$1
	if [ "$OS" == "Darwin" ]; then
		ARCH=$(uname -m)
		if [ "$ARCH" == "arm64" ]; then
			# Compile for ARM64 and x86_64 and create a universal binary
			env CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go clean --tags ${lua_tag}
			env CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" --tags ${lua_tag} -o getnovel-${lua_tag}-arm64
			env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go clean --tags ${lua_tag}
			env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" --tags ${lua_tag} -o getnovel-${lua_tag}-amd64
			lipo -create -output getnovel-${lua_tag} getnovel-${lua_tag}-amd64 getnovel-${lua_tag}-arm64
			rm getnovel-${lua_tag}-arm64 getnovel-${lua_tag}-amd64
		elif [ "$ARCH" == "x86_64" ]; then
			# Compile only for x86_64
			env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go clean --tags ${lua_tag}
			env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" --tags ${lua_tag} -o getnovel-${lua_tag}
		fi
	else
		# For other OSes
		env CGO_ENABLED=1 go clean --tags ${lua_tag}
		env CGO_ENABLED=1 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" --tags ${lua_tag} -o getnovel-${lua_tag}
	fi
}

# Check if a tag is provided
if [ -n "$1" ]; then
	compile_lua "lua$1"
else
	for lua in lua51 lua52 lua53 lua54 luajit; do
		compile_lua ${lua}
	done
fi
