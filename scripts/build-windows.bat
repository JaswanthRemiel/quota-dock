@echo off
REM QuotaDock Windows Build Script
echo ========================================================
echo Building QuotaDock for Windows (Daemon + System Tray)
echo ========================================================

cd /d "%~dp0..\daemon"
echo [1/2] Compiling Go Daemon for Windows (x64)...
set GOOS=windows
set GOARCH=amd64
go build -o daemon.exe .
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to compile daemon.exe
    exit /b %ERRORLEVEL%
)
echo [OK] daemon.exe compiled successfully.

cd /d "%~dp0..\windows\QuotaDock.Windows"
echo [2/2] Building Windows WPF System Tray Application...
where dotnet >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    dotnet build -c Release
    if %ERRORLEVEL% NEQ 0 (
        echo [WARNING] dotnet build returned non-zero code.
    ) else (
        echo [OK] QuotaDock.Windows built successfully.
    )
) else (
    echo [INFO] dotnet SDK not found in PATH. Skipping WPF build (Go daemon ready).
)

echo ========================================================
echo Build complete! Run scripts\run-windows.ps1 to launch.
echo ========================================================
