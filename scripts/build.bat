@echo off
setlocal EnableDelayedExpansion
call "%~dp0_init.bat"
call "%~dp0build-paths.bat"

where go >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go is not installed. Install from https://go.dev/dl/
    exit /b 1
)

call :find_gcc
if errorlevel 1 (
    echo.
    echo [WARN] gcc not found - GUI build skipped. CLI will still build.
    echo.
    set BUILD_GUI=0
) else (
    set BUILD_GUI=1
    set CGO_ENABLED=1
)

go mod tidy
if errorlevel 1 exit /b 1

echo.
echo Embedding admin manifest for CLI...
call "%~dp0embed-resources.bat" cmd\opwax-cli\opwax-cli.manifest cmd\opwax-cli\opwax-cli.syso
if errorlevel 1 (
    echo [WARN] CLI manifest embed failed - %CLI_EXE% may not auto-elevate.
    set CLI_MANIFEST=0
) else (
    set CLI_MANIFEST=1
)

echo Building CLI binary...
set CGO_ENABLED=0
go build -ldflags="-s -w" -o "%CLI_EXE%" ./cmd/opwax-cli
if errorlevel 1 exit /b 1
echo   OK: %CLI_EXE%

if "!BUILD_GUI!"=="1" (
    echo Embedding admin manifest for GUI...
    call "%~dp0embed-resources.bat" cmd\opwax\opwax.manifest cmd\opwax\opwax.syso
    if errorlevel 1 (
        echo [WARN] GUI manifest embed failed - install rsrc or add assets\opwax.ico
        set GUI_MANIFEST=0
    ) else (
        set GUI_MANIFEST=1
    )

    echo Building GUI binary...
    set CGO_ENABLED=1
    go build -ldflags="-H windowsgui -s -w" -o "%GUI_EXE%" ./cmd/opwax
    if errorlevel 1 (
        echo [ERROR] GUI build failed. CLI binary is ready: %CLI_EXE%
        exit /b 1
    )
    echo   OK: %GUI_EXE%
) else (
    echo Skipped %GUI_EXE% ^(GUI^) - install GCC and re-run scripts\build.bat
)

if "!CLI_MANIFEST!"=="0" echo [WARN] Re-run after: go install github.com/akavel/rsrc@latest

echo.
echo Binaries: %BIN_DIR%\
echo Optional: scripts\build-release.bat for portable zip in releases\
endlocal
exit /b 0

:find_gcc
where gcc >nul 2>&1
if not errorlevel 1 exit /b 0

if defined LOCALAPPDATA (
    for /d %%D in ("%LOCALAPPDATA%\Microsoft\WinGet\Packages\BrechtSanders.WinLibs*") do (
        if exist "%%~D\mingw64\bin\gcc.exe" (
            set "PATH=%%~D\mingw64\bin;%PATH%"
            exit /b 0
        )
        if exist "%%~D\bin\gcc.exe" (
            set "PATH=%%~D\bin;%PATH%"
            exit /b 0
        )
    )
)

for %%P in (
    "C:\Program Files\WinLibs\bin"
    "C:\Program Files\WinLibs\mingw64\bin"
    "C:\msys64\mingw64\bin"
    "C:\msys64\ucrt64\bin"
    "C:\TDM-GCC-64\bin"
    "C:\MinGW\bin"
    "C:\mingw-w64\bin"
) do (
    if exist %%P\gcc.exe (
        set "PATH=%%~P;%PATH%"
        exit /b 0
    )
)
exit /b 1
