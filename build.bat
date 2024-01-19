@echo off
set PATH=%PATH%;H:\msys64\mingw64\bin
set CGO_ENABLED=1
for /F "tokens=*" %%R in ('git rev-parse --short HEAD') do set REV=%%R
for /F "tokens=*" %%A in ('date /T') do set TODAY=%%A
@echo on
go build -ldflags="-s -w -X main.sha1ver=%REV% -X 'main.buildTime=%TODAY%'" -tags lua51 -o getnovel-lua51.exe
go build -ldflags="-s -w -X main.sha1ver=%REV% -X 'main.buildTime=%TODAY%'" -tags lua52 -o getnovel-lua52.exe
go build -ldflags="-s -w -X main.sha1ver=%REV% -X 'main.buildTime=%TODAY%'" -tags lua53 -o getnovel-lua53.exe
go build -ldflags="-s -w -X main.sha1ver=%REV% -X 'main.buildTime=%TODAY%'" -tags lua54 -o getnovel-lua54.exe
go build -ldflags="-s -w -X main.sha1ver=%REV% -X 'main.buildTime=%TODAY%'" -tags luajit -o getnovel-luajit.exe
