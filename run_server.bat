@echo off
echo ==============================================
echo  Markflow Studio - Server Restarter
echo ==============================================

echo [1/2] Stopping existing server processes...
taskkill /F /IM go.exe /IM main.exe 2>nul
if %ERRORLEVEL% equ 0 (
    echo Successfully stopped old server.
) else (
    echo No old server running.
)

echo.
echo [2/2] Starting new server on port 8888...
"C:\Program Files\Go\bin\go.exe" run cmd/server/main.go

pause
