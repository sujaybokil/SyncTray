@echo off
echo [1/2] Fetching dependencies...
go mod tidy
if errorlevel 1 goto :fail

echo [2/2] Building...
go build -ldflags="-H windowsgui -s -w" -o synctray.exe .
if errorlevel 1 goto :fail

echo Done! synctray.exe created.
goto :end

:fail
echo BUILD FAILED. Make sure Go is installed: https://go.dev/dl/
exit /b 1

:end
pause
