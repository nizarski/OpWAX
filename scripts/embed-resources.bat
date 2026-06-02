@echo off
setlocal EnableDelayedExpansion
REM Usage: embed-resources.bat <manifest.xml> <output.syso>
REM Embeds admin manifest and optional assets/opwax.ico via rsrc.

set "MANIFEST=%~1"
set "OUT=%~2"

if "%MANIFEST%"=="" (
    echo [ERROR] embed-resources.bat: manifest path required
    exit /b 1
)
if "%OUT%"=="" (
    echo [ERROR] embed-resources.bat: output .syso path required
    exit /b 1
)
if not exist "%MANIFEST%" (
    echo [ERROR] Manifest not found: %MANIFEST%
    exit /b 1
)

if defined GOPATH (
    set "PATH=%GOPATH%\bin;%PATH%"
) else (
    set "PATH=%USERPROFILE%\go\bin;%PATH%"
)

where rsrc >nul 2>&1
if errorlevel 1 (
    echo Installing rsrc...
    go install github.com/akavel/rsrc@latest
    if errorlevel 1 (
        echo [ERROR] Failed to install rsrc. Run: go install github.com/akavel/rsrc@latest
        exit /b 1
    )
    where rsrc >nul 2>&1
    if errorlevel 1 (
        echo [ERROR] rsrc not on PATH after install. Add %%USERPROFILE%%\go\bin to PATH.
        exit /b 1
    )
)

set "ROOT=%~dp0.."
pushd "%ROOT%"

if exist "assets\opwax.ico" (
    echo   rsrc: manifest + assets\opwax.ico -^> %OUT%
    rsrc -manifest "%MANIFEST%" -ico assets\opwax.ico -o "%OUT%"
) else (
    echo   rsrc: manifest only ^(add assets\opwax.ico for custom icon^) -^> %OUT%
    rsrc -manifest "%MANIFEST%" -o "%OUT%"
)

set "RC=!errorlevel!"
popd
exit /b !RC!
