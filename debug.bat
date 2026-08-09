@echo off
setlocal

echo Building debug executable with symbols and a console...
go build -o synctray-debug.exe .
if errorlevel 1 goto :fail

echo.
echo Starting debug executable. Press Ctrl+C to stop it.
echo Use "debug.bat check" for a non-destructive diagnostic report.
set SYNCTRAY_DEBUG=1
synctray-debug.exe %*
goto :end

:fail
echo DEBUG BUILD FAILED. Make sure Go is installed: https://go.dev/dl/
exit /b 1

:end
pause
