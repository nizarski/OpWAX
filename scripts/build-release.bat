@echo off
setlocal EnableDelayedExpansion
call "%~dp0_init.bat"
call "%~dp0build-paths.bat"

echo Building release binaries...
call "%~dp0build.bat"
if errorlevel 1 exit /b 1

if not defined OPWAX_VER set "PKG_DIR=releases\OpWAX-portable"
if defined OPWAX_VER set "PKG_DIR=releases\OpWAX-%OPWAX_VER%"
set "PKG_ZIP=releases\OpWAX-portable.zip"
if defined OPWAX_VER set "PKG_ZIP=releases\OpWAX-%OPWAX_VER%.zip"

if exist "%PKG_DIR%" rmdir /s /q "%PKG_DIR%"
mkdir "%PKG_DIR%" 2>nul
mkdir "%PKG_DIR%\configs" 2>nul
mkdir "%PKG_DIR%\docs" 2>nul

copy /y "%GUI_EXE%" "%PKG_DIR%\" >nul 2>&1
copy /y "%CLI_EXE%" "%PKG_DIR%\" >nul 2>&1
copy /y configs\default.json "%PKG_DIR%\configs\" >nul
copy /y LICENSE "%PKG_DIR%\" >nul
copy /y AUTHORS "%PKG_DIR%\" >nul
copy /y README.md "%PKG_DIR%\" >nul
copy /y CHANGELOG.md "%PKG_DIR%\" >nul
copy /y docs\ARTIFACTS.md "%PKG_DIR%\docs\" >nul
copy /y docs\BUILD.md "%PKG_DIR%\docs\" >nul
mkdir "%PKG_DIR%\assets" 2>nul
copy /y assets\README.md "%PKG_DIR%\assets-README.md" >nul 2>&1
copy /y assets\opwax.png "%PKG_DIR%\assets\" >nul 2>&1
copy /y assets\opwax.ico "%PKG_DIR%\assets\" >nul 2>&1

(
echo OpWAX Portable Package
echo Built by nizarski
echo ========================
echo.
echo 1. Extract anywhere ^(USB, Desktop, etc.^)
echo 2. Run OpWAX.exe as Administrator ^(UAC manifest embedded^)
echo    Or CLI: OpWAX-cli.exe -config configs\default.json -preflight
echo.
echo Config: edit configs\default.json or use GUI Config tab.
echo Runtime data ^(second-pass task config^): %%ProgramData%%\OpWAX\
echo.
echo See LICENSE for terms.
) > "%PKG_DIR%\START_HERE.txt"

if exist "%PKG_ZIP%" del /f /q "%PKG_ZIP%"

powershell -NoProfile -Command "Compress-Archive -Path '%PKG_DIR%\*' -DestinationPath '%PKG_ZIP%' -Force"
if errorlevel 1 (
    echo [WARN] Zip failed - folder ready at %PKG_DIR%
    exit /b 0
)

echo.
echo Release ready:
echo   Dev binaries: %BIN_DIR%\
echo   Package:      %PKG_DIR%\
echo   Zip:          %PKG_ZIP%
echo.
echo Optional EV signing: set SIGN_* env vars and run scripts\sign.bat
endlocal
