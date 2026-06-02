# OpWAX

Windows privacy hygiene: disable logging, then remove forensic artifacts. Admin required. Authorized systems only.

**v1.3.0** · nizarski · [CHANGELOG](CHANGELOG.md)

## Features

Disable-then-delete · Fyne GUI + CLI · dry-run · preflight · manifest diff · Task Scheduler · portable zip

## Docs

| | |
|--|--|
| [docs/README.md](docs/README.md) | Index |
| [docs/BUILD.md](docs/BUILD.md) | **Build + GCC setup for GUI** |
| [docs/ARTIFACTS.md](docs/ARTIFACTS.md) | Artifact map |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Structure, flags |

## Requirements

- Windows 10 / 11 / Server 2019+
- Go 1.22+
- **GCC (MinGW-w64)** - GUI only; see [docs/BUILD.md](docs/BUILD.md#gcc-on-windows-required-for-gui)

## Build

```powershell
.\scripts\build-cli.bat          # CLI only, no GCC
.\scripts\build.bat              # CLI + GUI (needs gcc on PATH)
.\scripts\build-release.bat      # releases\OpWAX-portable.zip
```

Output: `releases\bin\OpWAX-cli.exe`, `releases\bin\OpWAX.exe`

## Run

GUI: `releases\bin\OpWAX.exe` (UAC)  
Config: `configs\default.json`

```powershell
.\releases\bin\OpWAX-cli.exe -version
.\releases\bin\OpWAX-cli.exe -config configs\default.json -preflight
.\releases\bin\OpWAX-cli.exe -config configs\default.json -dry-run
.\releases\bin\OpWAX-cli.exe -config configs\default.json
```

GUI tabs: Targets · Modules · Options · Preview · Config · Manifest · Schedule

## Safety

Does not wipe live `$MFT`/`$LogFile`, delete user accounts, inject LSASS, or kill critical processes. Reboot may be required for pagefile, hiberfil, locked EVTX.

## License

[LICENSE](LICENSE) · [CONTRIBUTING](CONTRIBUTING.md)
