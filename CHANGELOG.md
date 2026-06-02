# Changelog

All notable changes to OpWAX are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed

- [docs/BUILD.md](docs/BUILD.md): GCC/MinGW setup guide for GUI (WinLibs, MSYS2, PATH, troubleshooting)
- Docs shortened; technical style across README and `docs/`
- Build scripts in `scripts\`; output in `releases\bin\` and `releases\OpWAX-portable*` (legacy `dist\` retired)
- Project attribution: **nizarski** (LICENSE, AUTHORS, GUI, CLI `-version`, manifests)

## [1.3.0] - 2026-05-30

### Added

- Portable release packaging (`scripts\build-release.bat` → portable zip; now under `releases\`)
- `LICENSE` (MIT)
- `assets/` icon placement guide; optional `.ico` embed at build time
- Admin UAC manifest embed for **both** `OpWAX.exe` and `OpWAX-cli.exe`
- `scripts\embed-resources.bat` - auto-installs `rsrc` when missing
- `scripts\sign.bat` - EV Authenticode signing template (signtool)
- Execution gap artifacts (Tier A): Search index, MUICache/RecentApps/AppCompatFlags, Syscache, shell caches, targeted EVTX, EventTranscript, WER
- Deep traces (Tier B): servicing logs, print spooler, developer traces, DO cache, SmartScreen
- Bootstrap collector kill: WSearch, DiagTrack, DPS, SysMain (with EventLog/Prefetch/USN)
- Focused cleanup mode (pause Explorer during run)
- Post-run verification and optional second-pass scheduled task after reboot
- Central-copy preflight warnings (WEF, EDR, cloud, BitLocker note)

### Changed

- Project identity **OpWAX** (module `github.com/opwax/opwax`, binaries `OpWAX.exe` / `OpWAX-cli.exe`, assets `opwax.ico` / `opwax.png`)
- Version bumped to 1.3.0
- Documentation aligned to **1-pass zero-fill** secure delete (was incorrectly documented as 3-pass)
- README: correct CLI path (`cmd/opwax-cli`), tab name **Config**, portable distribution docs
- `scripts\build.bat` warns when admin manifest is not embedded

## [1.2.0] - 2026-05-30

### Added

- Modern Win11 artifacts: Recall, PCA, Notepad, Office TrustRecords, Sysinternals, Outlook/Teams
- Enhanced event log stack disable + full `Winevt\Logs` purge
- Advanced options GUI panels; preflight for batch-2 features
- Auto-elevation, export report, cancel during run, manifest diff

### Changed

- GUI/UX polish; run-order bootstrap phase; MFT scrub runs last

## [1.0.0] - Initial

- Seven cleanup modules, Fyne GUI, dry-run, preflight, scheduler, native Go MFT free-space scrub
