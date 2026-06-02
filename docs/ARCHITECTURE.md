# Architecture

GUI (`cmd/opwax`) and CLI (`cmd/opwax-cli`) share orchestrator, `configs/default.json`, and modules.

```
cmd/opwax | cmd/opwax-cli → runner/gui → orchestrator → modules + preflight + verify
```

## Layout

`internal/`: `artifacts`, `config`, `gui`, `models`, `modules`, `orchestrator`, `preflight`, `runner`, `scheduler`, `system`, `util`, `verify`, `version`  
`scripts/` · `releases/bin/` · `configs/default.json`

## Phases

| # | Phase |
|---|--------|
| 0 | Preflight |
| 1 | Dry run |
| 2 | Disable (services, policies) |
| 3 | Clean (files, registry) |
| 4 | Secure (overwrite, LSASS caches, MFT scrub if enabled) |
| 5 | Finalize (reboot, verify, manifest diff) |

## Modules (`configs/default.json` → `modules`)

| Key | Focus |
|-----|--------|
| `volatile_memory` | DNS/ARP, creds, dumps |
| `registry_hives` | Hives, policies |
| `ntfs_metadata` | USN, pagefile, MFT scrub |
| `program_execution` | Prefetch, UserAssist, Recall, gaps |
| `system_logs` | EVTX, timeline |
| `persistence_storage` | Tasks, WMI, Run keys (default off) |
| `network_browser` | WLAN, browsers |

## CLI flags

| Flag | Action |
|------|--------|
| `-version` | Version banner, exit |
| `-config` | JSON profile |
| `-preflight` | Checks only; exit `2` if critical fail |
| `-dry-run` | List actions only |
| `-no-gui` | Headless (GUI binary) |
| `-manifest <path\|-`> | System manifest JSON |
| `-install-schedule` / `-uninstall-schedule` / `-schedule-status` | Task `OpWAX-PrivacyCleanup` |

## Build

[BUILD.md](BUILD.md) - GCC required for GUI.
