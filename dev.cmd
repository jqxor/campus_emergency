@echo off
setlocal

REM Enterprise one-click dev runner (Windows):
REM - Auto-build missing backend executables
REM - Start all services and persist PIDs for stop.cmd

set ROOT=%~dp0
powershell -ExecutionPolicy Bypass -File "%ROOT%scripts\start-all.ps1"

echo.
echo Tips:
echo - Stop all:  stop.cmd
echo - Build all: build.cmd
echo.
pause
