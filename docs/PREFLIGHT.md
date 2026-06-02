# Pre-flight

Runs before dry-run/execute. CLI-only: `-preflight` (exit `0` ok, `2` critical).

```powershell
.\releases\bin\OpWAX-cli.exe -config configs\default.json -preflight
```

| Check | Severity |
|-------|----------|
| No users/drives | Critical |
| Browsers running | Warning |
| Event Log disable | Warning |
| Pagefile locked | Warning |
| Multi-user hives | Info |

## Manifest diff

`options.manifest_diff` (default `true`): before/after scan → REMOVED / REDUCED / STILL PRESENT.

```json
"manifest_diff": false
```
