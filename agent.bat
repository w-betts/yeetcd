@echo off
REM Agent wrapper script for Windows.
REM
REM Maps subcommands to agent names and launches opencode with the appropriate agent.
REM Does NOT pass any prompt - user enters prompt in TUI.
REM
REM Usage:
REM     agent.bat spec    - Launch spec agent (4-phase structured workflow)
REM     agent.bat vibe    - Launch vibe agent (direct implementation workflow)

if "%~1"=="" (
    echo Error: No subcommand provided.
    call :print_usage
    exit /b 1
)

if "%~1"=="spec" goto :launch_spec
if "%~1"=="vibe" goto :launch_vibe

echo Error: Invalid subcommand '%~1'.
call :print_usage
exit /b 1

:launch_spec
opencode --agent spec
exit /b %ERRORLEVEL%

:launch_vibe
opencode --agent vibe
exit /b %ERRORLEVEL%

:print_usage
echo Usage: agent.bat ^<subcommand^>
echo.
echo Subcommands:
echo   spec   - Launch spec agent (4-phase structured workflow)
echo   vibe   - Launch vibe agent (direct implementation workflow)
echo.
echo Examples:
echo   agent.bat spec    - Start structured workflow for complex features
echo   agent.bat vibe    - Start direct workflow for rapid iteration
goto :eof
