@echo off
setlocal
call "%~dp0_init.bat"
call "%~dp0build-paths.bat"

call "%~dp0embed-resources.bat" cmd\opwax-cli\opwax-cli.manifest cmd\opwax-cli\opwax-cli.syso
if errorlevel 1 (
    echo [WARN] Manifest embed failed - install rsrc: go install github.com/akavel/rsrc@latest
)

set CGO_ENABLED=0
go build -ldflags="-s -w" -o "%CLI_EXE%" ./cmd/opwax-cli
if errorlevel 1 exit /b 1
echo Built %CLI_EXE% ^(admin manifest embedded when rsrc available^)
endlocal
