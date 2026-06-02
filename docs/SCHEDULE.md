# Schedule

Task name: `OpWAX-PrivacyCleanup` · highest privileges · `OpWAX-cli.exe -config <path>` (no GUI).

**GUI:** Schedule tab → mode/time → config path → Install

```powershell
.\releases\bin\OpWAX-cli.exe -install-schedule -config configs\default.json
.\releases\bin\OpWAX-cli.exe -schedule-status
.\releases\bin\OpWAX-cli.exe -uninstall-schedule
```

```json
"schedule": {
  "enabled": true,
  "mode": "daily",
  "time": "02:00",
  "day_of_week": 0,
  "day_of_month": 1,
  "config_path": "C:\\path\\config.json"
}
```

Modes: `at_logon` · `daily` · `weekly` · `monthly` (day 1–28) · `once`
