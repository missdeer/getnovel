#!/bin/bash

OS=$(uname -s)

if [ "$OS" == "Darwin" ]; then
	ARCH=$(uname -m)
	if [ "$ARCH" == "arm64" ]; then
		# Compile for ARM64 and x86_64 and create a universal binary
		env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go clean
		env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" -o getnovel-arm64
		env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go clean
		env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" -o getnovel-amd64
		lipo -create -output getnovel getnovel-amd64 getnovel-arm64
		rm getnovel-arm64 getnovel-amd64
	elif [ "$ARCH" == "x86_64" ]; then
		# Compile only for x86_64
		env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go clean
		env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" -o getnovel
	fi
else
	# For other OSes
	env CGO_ENABLED=0 go clean
	env CGO_ENABLED=0 go build -ldflags="-s -w -X main.sha1ver=$(git rev-parse --short HEAD) -X 'main.buildTime=$(date)'" -o getnovel
fi
