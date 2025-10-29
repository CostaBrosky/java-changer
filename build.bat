@echo off
echo Building Java Version Switcher (jv)...
echo.

REM Download dependencies
echo [1/3] Downloading dependencies...
go mod download
if errorlevel 1 (
    echo Error downloading dependencies!
    exit /b 1
)

REM Build the executable
echo [2/3] Building executable...
go build -ldflags="-s -w" -o jv.exe .
if errorlevel 1 (
    echo Error building executable!
    exit /b 1
)

REM Success
echo [3/3] Build successful!
echo.
echo Executable created: jv.exe
echo.
echo To install system-wide, copy jv.exe to a directory in your PATH
echo Example: copy jv.exe C:\tools
echo.
echo Or add the current directory to your PATH environment variable.
echo.

pause
