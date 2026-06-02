# scripts/

Run from repo root (`_init.bat` sets `cd`).

| Script | Output |
|--------|--------|
| `build-cli.bat` | `releases\bin\OpWAX-cli.exe` |
| `build.bat` | CLI + GUI (needs `gcc`) |
| `build-release.bat` | `releases\OpWAX-portable.zip` |
| `sign.bat` | Signs `releases\bin\*.exe`, zip |
| `embed-resources.bat` | UAC manifest + optional `.ico` |

GCC setup: [docs/BUILD.md](../docs/BUILD.md#gcc-on-windows-required-for-gui)
