@echo off
REM Open the human-facing documentation in a browser

setlocal

set SCRIPT_DIR=%~dp0
set DOC_PATH=%SCRIPT_DIR%documentation\human\index.html

if not exist "%DOC_PATH%" (
    echo Error: Documentation not found at %DOC_PATH%
    echo Please run the documentation agent first to generate the documentation.
    exit /b 1
)

start "" "%DOC_PATH%"

endlocal
