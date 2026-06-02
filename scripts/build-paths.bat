@echo off
REM Output paths under releases\ (call after _init.bat).
set "BIN_DIR=releases\bin"
set "PKG_DIR=releases\OpWAX-portable"
set "PKG_ZIP=releases\OpWAX-portable.zip"
set "CLI_EXE=%BIN_DIR%\OpWAX-cli.exe"
set "GUI_EXE=%BIN_DIR%\OpWAX.exe"
if not exist releases mkdir releases
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"
