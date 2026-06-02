# System manifest

JSON snapshot of artifact paths on this host. Static reference: [ARTIFACTS.md](ARTIFACTS.md).

**GUI:** Manifest tab → Generate → Save  
**CLI:**

```powershell
.\releases\bin\OpWAX-cli.exe -manifest out.json
.\releases\bin\OpWAX-cli.exe -manifest -
```

Admin required. Top-level fields: `generated_at`, `hostname`, `os_version`, `users`, `drives`, `artifacts[]`, `event_logs`.
