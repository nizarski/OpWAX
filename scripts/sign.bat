@echo off
setlocal EnableDelayedExpansion
call "%~dp0_init.bat"
call "%~dp0build-paths.bat"

REM EV code signing - requires Windows SDK signtool + EV certificate on token/HSM.
REM Set before running:
REM   set SIGN_CERT=YourPublisherName
REM   set SIGN_FILE=C:\path\to\ev.pfx
REM   set SIGN_PASSWORD=...

if not exist "%GUI_EXE%" if not exist "%CLI_EXE%" (
    echo [ERROR] Build first: scripts\build.bat
    exit /b 1
)

where signtool >nul 2>&1
if errorlevel 1 (
    echo [ERROR] signtool not found. Install Windows SDK or Visual Studio Build Tools.
    exit /b 1
)

set "TS=http://timestamp.digicert.com"
set "FILES="
if exist "%GUI_EXE%" set "FILES=!FILES! %GUI_EXE%"
if exist "%CLI_EXE%" set "FILES=!FILES! %CLI_EXE%"
if exist "%PKG_ZIP%" set "FILES=!FILES! %PKG_ZIP%"

if defined SIGN_FILE (
    if not defined SIGN_PASSWORD (
        echo Enter PFX password:
        set /p SIGN_PASSWORD=
    )
    for %%F in (%FILES%) do (
        echo Signing %%~F ^(PFX^)...
        signtool sign /fd SHA256 /tr %TS% /td SHA256 /f "%SIGN_FILE%" /p "%SIGN_PASSWORD%" "%%F"
        if errorlevel 1 exit /b 1
    )
) else if defined SIGN_CERT (
    for %%F in (%FILES%) do (
        echo Signing %%~F ^(store: %SIGN_CERT%^)...
        signtool sign /fd SHA256 /tr %TS% /td SHA256 /n "%SIGN_CERT%" "%%F"
        if errorlevel 1 exit /b 1
    )
) else (
    echo [ERROR] Set SIGN_CERT=PublisherName or SIGN_FILE=path\to\ev.pfx
    exit /b 1
)

echo.
echo Verify:
for %%F in (%FILES%) do signtool verify /pa /v "%%F"
echo Done.
endlocal
