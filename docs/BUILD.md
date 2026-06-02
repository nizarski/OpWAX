# Build

| Output | Path |
|--------|------|
| CLI | `releases\bin\OpWAX-cli.exe` |
| GUI | `releases\bin\OpWAX.exe` |
| Portable zip | `releases\OpWAX-portable.zip` |

Scripts: `scripts\`. Index: [README.md](README.md).

## Prerequisites

| Component | CLI | GUI |
|-----------|-----|-----|
| Windows 10/11+ | Yes | Yes |
| [Go 1.22+](https://go.dev/dl/) | Yes | Yes |
| MinGW-w64 **gcc** on PATH | No | Yes (`CGO_ENABLED=1`, Fyne) |

## CLI only (no GCC)

```powershell
.\scripts\build-cli.bat
.\releases\bin\OpWAX-cli.exe -config configs\default.json -preflight
```

## GCC on Windows (required for GUI)

Fyne links C code. You need `gcc.exe` on PATH in the **same** shell you run `build.bat`.

### Option A - WinLibs (recommended)

```powershell
winget install -e --id BrechtSanders.WinLibs.POSIX.UCRT
```

Close and reopen PowerShell, then:

```powershell
gcc --version
where gcc
```

Expected: path containing `mingw64\bin\gcc.exe`. `build.bat` also searches WinGet’s package folder under `%LOCALAPPDATA%\Microsoft\WinGet\Packages\`.

If `gcc` is not found, add manually (adjust path to your install):

```powershell
$env:Path = "C:\Program Files\WinLibs\mingw64\bin;$env:Path"
# or WinGet layout, e.g.:
# $env:Path = "$env:LOCALAPPDATA\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT64.MSVCRT64\mingw64\bin;$env:Path"
```

Permanent (user PATH): Settings → System → About → Advanced system settings → Environment Variables → edit **Path** → add the `...\mingw64\bin` folder.

### Option B - MSYS2

```powershell
winget install -e --id MSYS2.MSYS2
```

Open **UCRT64** or **MINGW64** from Start Menu, then:

```bash
pacman -S --needed mingw-w64-ucrt-x86_64-gcc
```

Use that MSYS2 shell, **or** add to Windows PATH:

- `C:\msys64\ucrt64\bin` (UCRT64), or  
- `C:\msys64\mingw64\bin` (MINGW64)

```powershell
gcc --version
```

### Build GUI + CLI

```powershell
cd <repo-root>
.\scripts\build.bat
```

Success: `releases\bin\OpWAX.exe` and `OpWAX-cli.exe`.  
If you see `gcc not found` / `GUI build skipped`, fix PATH and run again in a **new** terminal.

Manual (same flags as `build.bat`):

```powershell
$env:CGO_ENABLED = "1"
go build -ldflags="-H windowsgui -s -w" -o releases\bin\OpWAX.exe ./cmd/opwax
```

CLI without CGO:

```powershell
$env:CGO_ENABLED = "0"
go build -ldflags="-s -w" -o releases\bin\OpWAX-cli.exe ./cmd/opwax-cli
```

### GCC troubleshooting

| Symptom | Fix |
|---------|-----|
| `gcc: command not found` | Install WinLibs/MSYS2; reopen terminal; `where gcc` |
| `build.bat` skips GUI | GCC not on PATH in that session; add `mingw64\bin` |
| `cgo: C compiler "gcc" not found` | Same as above; verify `gcc --version` before `go build` |
| Wrong architecture | Use **x86_64** MinGW, not 32-bit MinGW |
| MSYS2 `gcc` works in MSYS2 only | Add `ucrt64\bin` or `mingw64\bin` to Windows PATH |

## Manifest / icon

`scripts\embed-resources.bat` runs `rsrc` (auto `go install github.com/akavel/rsrc@latest`). Embeds admin manifest; adds `assets\opwax.ico` if present.

## Portable release

```powershell
.\scripts\build-release.bat
```

Zip layout: `OpWAX.exe` beside `configs\` (not under `releases\bin\`).

## Signing

```powershell
set SIGN_CERT=Publisher Name
.\scripts\sign.bat
signtool verify /pa /v releases\bin\OpWAX.exe
```

Requires Windows SDK `signtool`.

## Other errors

| Error | Fix |
|-------|-----|
| `rsrc not found` | `go install github.com/akavel/rsrc@latest`; add `%USERPROFILE%\go\bin` to PATH |
| UAC not shown | Rebuild; check `cmd\opwax\opwax.syso` exists |
| `signtool not found` | Install Windows 10/11 SDK |
