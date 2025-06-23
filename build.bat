@echo off
where gcc >nul 2>&1
if %ERRORLEVEL% == 0 (
echo GCC is available.
) else (
set "PATH=%PATH%;D:\msys64\mingw64\bin"
)
set CGO_ENABLED=1
for /F "tokens=*" %%R in ('git rev-parse --short HEAD') do set REV=%%R
for /F "tokens=*" %%A in ('date /T') do set TODAY=%%A

if "%1"=="" goto buildall
if "%1"=="51" goto build51
if "%1"=="52" goto build52
if "%1"=="53" goto build53
if "%1"=="54" goto build54
if "%1"=="jit" goto buildjit
@goto end

:buildall
@call :build 51
@call :build 52
@call :build 53
@call :build 54
@call :build jit
@goto end

:build51
@call :build 51
@goto end

:build52
@call :build 52
@goto end

:build53
@call :build 53
@goto end

:build54
@call :build 54
@goto end

:buildjit
@call :build jit
@goto end

:build
@go clean -tags lua%1
@echo on
go build -ldflags="-s -w -X main.sha1ver=%REV% -X 'main.buildTime=%TODAY%'" -tags lua%1 -o getnovel-lua%1.exe
@goto :eof

:end