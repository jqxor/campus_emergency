@echo off
setlocal
set ROOT=%~dp0
powershell -ExecutionPolicy Bypass -File "%ROOT%scripts\stop-all.ps1"
