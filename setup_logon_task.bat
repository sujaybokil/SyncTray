@echo off
setlocal

set TASK_NAME=SyncTray
set EXE_PATH=%~dp0synctray.exe

if not exist "%EXE_PATH%" (
    echo ERROR: synctray.exe not found at:
    echo   %EXE_PATH%
    echo Build it first with: go build -ldflags="-H windowsgui -s -w" -o synctray.exe .
    pause
    exit /b 1
)

echo Creating Task Scheduler logon task...
echo   Task : %TASK_NAME%
echo   Exe  : %EXE_PATH%
echo   User : %USERDOMAIN%\%USERNAME%
echo   Delay: 30 seconds after logon
echo.

schtasks /create ^
  /tn "%TASK_NAME%" ^
  /tr "\"%EXE_PATH%\"" ^
  /sc ONLOGON ^
  /ru "%USERDOMAIN%\%USERNAME%" ^
  /delay 0000:30 ^
  /it ^
  /f

if errorlevel 1 (
    echo.
    echo FAILED to create task.
    pause
    exit /b 1
)

echo.
echo Done! Task "%TASK_NAME%" will run 30s after next logon.
echo.
echo Useful commands:
echo   Run now  : schtasks /run /tn "%TASK_NAME%"
echo   Delete   : schtasks /delete /tn "%TASK_NAME%" /f
echo   Status   : schtasks /query /tn "%TASK_NAME%"
echo.
pause
