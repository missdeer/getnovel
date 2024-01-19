#!/bin/bash
OS=$(uname -s)
if [ "$OS" == "Darwin" ]; then
	for lua in lua51 lua52 lua53 lua54 luajit; do
		env CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" --tags ${lua} -o getnovel-${lua}-arm64
		env CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" --tags ${lua} -o getnovel-${lua}-amd64
		lipo -create -output getnovel-${lua} getnovel-${lua}-amd64 getnovel-${lua}-arm64
		rm getnovel-${lua}-arm64 getnovel-${lua}-amd64
	done
else
	for lua in lua51 lua52 lua53 lua54 luajit; do
		env CGO_ENABLED=1 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" --tags ${lua} -o getnovel-${lua}
	done
fi
