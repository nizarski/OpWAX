# OpWAX artifact manifest

Artifact map: disable future logging, then delete existing data. Per-host discovery: [MANIFEST.md](MANIFEST.md).

Target OS: Windows 10, Windows 11, Windows Server 2019/2022 (paths resolved at runtime).

---

## Module: Volatile Memory

### 1. Active Processes & Threads (RAM / Kernel)

| Field | Value |
|-------|-------|
| Location | Volatile RAM, `EPROCESS` / `ETHREAD` structures |
| User-mode exposure | Task Manager, `Get-Process`, WMI `Win32_Process` |
| Disable future | Not applicable - processes are runtime state |
| Clean existing | Flush non-essential caches only; **do not terminate system processes** |
| Actions | Clear DNS client cache (`ipconfig /flushdns`), flush ARP cache (`arp -d *`), reset TCP statistics where safe |

### 2. Network Connections & Listening Ports

| Field | Value |
|-------|-------|
| Location | Volatile RAM / TCP/IP stack |
| Disable future | Disable `Tcpip` connection logging via registry (limited effect in user-mode) |
| Clean existing | `netsh int ip reset` avoided (disruptive); flush DNS, clear `C:\Windows\System32\LogFiles\Firewall\*.log` if present |

### 3. Cleartext Credentials & Encryption Keys (LSASS)

| Field | Value |
|-------|-------|
| Location | LSASS process memory (`lsass.exe`) |
| Disable future | Enable Credential Guard / LSA protection (optional, requires reboot - skipped in live mode) |
| Clean existing | User-mode: cmdkey, CredEnumerate/CredDelete, vaultcmd, klist purge, WDigest disable |
| Live LSASS RAM | **Reboot required** for full memory scrub - injection avoided (system crash risk) |
| Safety | **No LSASS injection or memory scraping** |

### 4. Injected Code / Process VAD

| Field | Value |
|-------|-------|
| Location | Process VAD trees |
| Disable future | Not applicable in user-mode |
| Clean existing | Skipped - requires kernel access; documented only |

### 4a. Kernel Callbacks & Rootkit Hooks

| Field | Value |
|-------|-------|
| Location | SSDT / kernel callback arrays, filter drivers |
| Disable future | Not applicable in user-mode |
| Clean existing | **Skipped** - requires kernel debugging or memory forensics (Volatility, etc.) |

### 4b. Timestomping ($STANDARD_INFORMATION vs $FILE_NAME)

| Field | Value |
|-------|-------|
| Location | `$MFT` file records (dual timestamps) |
| Disable future | Not disableable |
| Clean existing | **Not detected or repaired by OpWAX** - dry-run notes that live `$MFT` timestamp mismatch scans are out of scope |
| Safety | OpWAX does **not** parse live `$MFT` for anti-forensics evidence |

---

## Modern Windows 11 & AI Artifacts

### 31. Windows Recall (CoreAI Platform)

| Field | Value |
|-------|-------|
| Path | `%LocalAppData%\CoreAIPlatform\Capture\` (SQLite capture DB when enabled) |
| Disable future | Recall / AI snapshot policies via registry |
| Clean existing | Delete capture directory contents |

### 32. PCA Launch Dictionary

| Field | Value |
|-------|-------|
| Path | `C:\Windows\appcompat\Programs\PcaAppLaunchDic.txt` (+ related `Pca*` files) |
| Disable future | PcaSvc disabled by Execution module |
| Clean existing | Secure-delete PCA trace files |

### 33. Notepad Tab / Draft Cache

| Field | Value |
|-------|-------|
| Path | `%LocalAppData%\Packages\Microsoft.WindowsNotepad_*\LocalState\TabState\` |
| Disable future | N/A |
| Clean existing | Remove tab state folder |

---

## Email & Communication

### 34. Outlook OST/PST

| Field | Value |
|-------|-------|
| Path | `%LocalAppData%\Microsoft\Outlook\*.ost`, `*.pst` |
| Disable future | Optional cache-only vs full offline store delete (user choice) |
| Clean existing | `cache` or `delete_ost_pst` mode |

### 35. Microsoft Teams Cache

| Field | Value |
|-------|-------|
| Path | `%AppData%\Microsoft\Teams\`, `%LocalAppData%\Packages\*Teams*\` |
| Disable future | N/A |
| Clean existing | `cache` or `full` profile wipe (user choice) |

---

## Registry: Office & Sysinternals

### 36. Office Trusted Documents

| Field | Value |
|-------|-------|
| Hive | `HKCU\Software\Microsoft\Office\<ver>\<App>\Security\Trusted Documents\TrustRecords` |
| Clean existing | Delete TrustRecords keys per Office app |

### 37. Sysinternals EULA Approvals

| Field | Value |
|-------|-------|
| Hive | `HKCU\Software\Sysinternals\<ToolName>` |
| Clean existing | Delete entire Sysinternals key tree |

---

## Event Logs (Enhanced)

When **System Logs** module is enabled:

1. Clear audit policy  
2. Stop/disable **EventLog**, **Wecsvc**, **WdiSystemHost**  
3. Disable all channels via `wevtutil`  
4. Clear channels, then **purge entire** `C:\Windows\System32\Winevt\Logs\`  

**Note:** Security log clearing may itself be logged (Event ID **1102**) - preflight warns that log wiping is detectable.

### Post-Run Verification

When `post_run_verification` is enabled (default), OpWAX rescans key paths after cleanup and reports gaps:

| Category | Examples checked |
|----------|------------------|
| execution | Prefetch folder, Amcache.hve, PowerShell history |
| logs | Winevt\\Logs, jump lists, Timeline DB |
| ntfs | USN journal still active on target drives |
| modern | Recall capture, PCA logs, Notepad tab cache |
| network | SRUDB.dat, Outlook OST/PST (when delete mode) |

Gaps usually mean locked files (reboot needed) or a module was disabled.

---

## Module: Registry Hives

### 5. Hardware & USB Connection History

| Field | Value |
|-------|-------|
| Hive | `HKLM\SYSTEM` → `CurrentControlSet\Enum\USBSTOR`, `USB`, `SCSI`, `IDE`, `STORAGE` |
| Disable future | Registry key permissions deny write to `Enum` (reverted on some updates - also set `DisableUSBStorage` policy optional) |
| Clean existing | Delete subkeys under `Enum\USBSTOR`, `Enum\USB`, mounted devices in `MountedDevices` (USB entries only, preserve boot drives) |

### 6. Installed Applications & Autostart

| Field | Value |
|-------|-------|
| Hive | `HKLM\SOFTWARE`, `HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run` |
| Disable future | Not removed - would break apps; clean **MRU / Recent** keys only |
| Clean existing | Clear `RecentDocs`, `RunMRU`, `TypedPaths`, `WordWheelQuery`, `Explorer\RecentDocs`, `ComDlg32\OpenSavePidlMRU`, `ComDlg32\LastVisitedPidlMRU` |

### 7. Local User Accounts & Password Hashes (SAM)

| Field | Value |
|-------|-------|
| File | `C:\Windows\System32\config\SAM` |
| Disable future | Cannot disable SAM - required by OS |
| Clean existing | Clear `HKLM\SAM\...\Cache` (cached domain creds) if accessible; **do not delete user accounts** |

### 8. User Shellbags & Recent Docs (NTUSER.DAT)

| Field | Value |
|-------|-------|
| File | `C:\Users\<User>\NTUSER.DAT` |
| Keys | `Software\Microsoft\Windows\CurrentVersion\Explorer\Shell Bags`, `BagMRU`, `RecentDocs`, `StreamMRU`, `TypedPaths` |
| Disable future | Set `NoRecentDocsHistory=1`, `ClearRecentDocsOnExit=1` under Explorer policy keys |
| Clean existing | Delete ShellBags, BagMRU, RecentDocs, Open/Save MRU keys |

### 9. User Class & Shellbags (UsrClass.dat)

| Field | Value |
|-------|-------|
| File | `C:\Users\<User>\AppData\Local\Microsoft\Windows\UsrClass.dat` |
| Keys | `Local Settings\Software\Microsoft\Windows\Shell\BagMRU`, `Bags` |
| Disable future | Same Explorer policies as NTUSER |
| Clean existing | Delete BagMRU and Bags keys |

---

## Module: NTFS Metadata

### 10. Master File Table ($MFT)

| Field | Value |
|-------|-------|
| Location | `\\.\C:` volume `$MFT` |
| Disable future | Not disableable - inherent to NTFS |
| Clean existing | **Free MFT records**: `sdelete -c` or `cipher /w:` after file deletions. Does **not** wipe live $MFT (would destroy FS) |
| Tool | Sysinternals SDelete (preferred) or built-in `cipher /w` |

### 11. Transactional Log ($LogFile)

| Field | Value |
|-------|-------|
| Location | `$LogFile` |
| Disable future | Not disableable |
| Clean existing | **System drive**: schedule `chkdsk /F` on reboot. **Other drives**: dismount + `chkdsk /F` live |
| Safety | Microsoft-supported log replay - not a raw $LogFile delete |

### 12. USN Journal ($Extend\$UsnJrnl)

| Field | Value |
|-------|-------|
| Location | `$Extend\$UsnJrnl:$J` |
| Disable future | `fsutil usn deletejournal /D <drive>:` then disable via `fsutil behavior set disablelastaccess 1` |
| Clean existing | `fsutil usn deletejournal /D <drive>:` per selected volume |

### 13. Zone Identifier ADS

| Field | Value |
|-------|-------|
| Location | `<file>:Zone.Identifier` alternate data stream |
| Disable future | Disable Attachment Manager: `HKLM\...\Attachments\SaveZoneInformation=1` (don't save zones) |
| Clean existing | Recursive scan on selected drives; strip `:Zone.Identifier` ADS from all files |

---

## Module: Program Execution

### 14. Prefetch Files

| Field | Value |
|-------|-------|
| Path | `C:\Windows\Prefetch\*.pf` |
| Disable future | `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management\PrefetchParameters\EnablePrefetcher=0`, `EnableSuperfetch=0`; stop & disable `SysMain` service |
| Clean existing | Delete all `*.pf` after disabling |

### 15. SuperFetch / SysMain Database

| Field | Value |
|-------|-------|
| Path | `C:\Windows\Prefetch\Ag*.db`, `C:\Windows\Prefetch\*.db` |
| Disable future | Same as Prefetch (SysMain disabled) |
| Clean existing | Delete `Ag*.db` and related DB files |

### 16. AppCompatCache (Shimcache)

| Field | Value |
|-------|-------|
| Key | `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\AppCompatCache` |
| Disable future | Delete `AppCompatCache` value; set `DisableAppCompat=1` (custom, may not exist on all builds) |
| Clean existing | Clear `AppCompatCache` binary value |

### 17. Amcache.hve

| Field | Value |
|-------|-------|
| Path | `C:\Windows\appcompat\Programs\Amcache.hve`, `Amcache.hve.LOG*` |
| Disable future | Stop `PcaSvc` (Program Compatibility Assistant), set startup Disabled |
| Clean existing | Take ownership, delete hive files after stopping PcaSvc |

### 18. BAM / DAM (Background Activity Moderator)

| Field | Value |
|-------|-------|
| Keys | `HKLM\SYSTEM\CurrentControlSet\Services\bam\State\UserSettings`, `...\dam\State\UserSettings` |
| Disable future | Disable `bam` and `dam` services via registry `Start=4` |
| Clean existing | Delete all subkeys under `UserSettings` |

---

## Module: User Activity & System Logs

### 19. Windows Event Logs (.evtx)

| Field | Value |
|-------|-------|
| Path | `C:\Windows\System32\Winevt\Logs\*.evtx` |
| Disable future | Stop `EventLog` service, set `Start=4`; disable audit policies via `auditpol /clear /y`; `wevtutil sl` each log `/e:false` |
| Clean existing | `wevtutil cl <log>` for all channels, then delete residual `.evtx` if service stopped |

### 20. PowerShell Operational Log

| Field | Value |
|-------|-------|
| Path | `Microsoft-Windows-PowerShell%4Operational.evtx` |
| Disable future | Covered by global event log disable + `HKCU\...\PowerShell\ScriptBlockLogging=0`, `EnableScriptBlockLogging=0` |
| Clean existing | Clear channel + disable Script Block Logging |

### 21. Jump Lists (AutomaticDestinations)

| Field | Value |
|-------|-------|
| Path | `%AppData%\Microsoft\Windows\Recent\AutomaticDestinations\*`, `CustomDestinations\*` |
| Disable future | `NoRecentDocsHistory=1` |
| Clean existing | Delete all files in both folders |

### 22. Recent Shortcuts (.lnk)

| Field | Value |
|-------|-------|
| Path | `%AppData%\Microsoft\Windows\Recent\*.lnk` |
| Disable future | Same as jump lists |
| Clean existing | Delete all `.lnk` in Recent folder |

---

## Module: Persistence & Storage Overflows

### 23. Pagefile.sys

| Field | Value |
|-------|-------|
| Path | `<Drive>:\pagefile.sys` |
| Disable future | `HKLM\...\Memory Management\ClearPageFileAtShutdown=1`, `PagingFiles` empty to disable pagefile |
| Clean existing | Set clear-at-shutdown; attempt `wmic pagefileset delete` + recreate disabled; schedule deletion on next boot via `PendingFileRenameOperations` if locked |

### 24. MEMORY.DMP

| Field | Value |
|-------|-------|
| Path | `C:\Windows\MEMORY.DMP`, `C:\Windows\Minidump\*` |
| Disable future | `CrashDumpEnabled=0`, `LogEvent=0` under `CrashControl` |
| Clean existing | Secure-delete dump files |

### 25. hiberfil.sys

| Field | Value |
|-------|-------|
| Path | `C:\hiberfil.sys` |
| Disable future | `powercfg /hibernate off` |
| Clean existing | Removed automatically when hibernate disabled (may need admin + not locked) |

### 26. Recycle Bin

| Field | Value |
|-------|-------|
| Path | `C:\$Recycle.Bin\<SID>\$R*`, `$I*` |
| Disable future | Not disableable |
| Clean existing | 1-pass zero-fill `$R*` content, delete `$I*` metadata |

---

## Module: Network & Browser

### 27. SRUM (Network Data Usage)

| Field | Value |
|-------|-------|
| Path | `C:\Windows\System32\sru\SRUDB.dat`, `SRUDB.dat.LOG*`, `SRU*.log` |
| Disable future | Disable `DiagTrack` (Connected User Experiences) and `DPS` (Diagnostic Policy Service) |
| Clean existing | Stop services, secure-delete SRU database files |

### 28. WLAN Profiles

| Field | Value |
|-------|-------|
| Path | `C:\ProgramData\Microsoft\Wlansvc\Profiles\Interfaces\{GUID}\*.xml` |
| Disable future | User choice: keep current / delete all |
| Clean existing | Delete profile XMLs (user choice: all or except connected) |

### 29. Chromium Browsers (Edge / Chrome)

| Field | Value |
|-------|-------|
| Paths | `%LocalAppData%\Microsoft\Edge\User Data\Default\History`, `%LocalAppData%\Google\Chrome\User Data\Default\History`, `Downloads`, `Cookies`, `Web Data`, `Cache` |
| Disable future | N/A (browser-level) |
| Clean existing | Delete history DBs and caches; kill browser processes if running then retry |

### 30. Firefox

| Field | Value |
|-------|-------|
| Path | `%AppData%\Mozilla\Firefox\Profiles\*\places.sqlite`, `cookies.sqlite`, `cache2` |
| Disable future | N/A |
| Clean existing | Delete SQLite DBs and cache; kill `firefox.exe` if running |

---

## Disable-Before-Delete Execution Order

```
0. Bootstrap: EventLog + Prefetch/SysMain/PcaSvc + USN reset + WSearch/DiagTrack/DPS/SysMain collector kill; optional Explorer pause
1. Disable all modules
2. Clean all modules (Tier A/B gap artifacts included)
3. Secure: MFT free-record scrub (optional, last)
4. Verify + manifest diff; schedule one-time second pass if reboot/gaps
5. Optional reboot for locked files / pagefile / hiberfil
```

### Central copies (preflight warning)

Local cleanup **cannot** remove WEF/SIEM forwards, EDR cloud telemetry, Entra/M365 logs, Intune records, OneDrive versions, or backup snapshots. Preflight always warns when any module is enabled.

---

## Execution gap artifacts (Tier A)

### 38. Windows Search Index

| Field | Value |
|-------|-------|
| Path | `%ProgramData%\Microsoft\Search\Data\Applications\Windows\Windows.edb` |
| Disable | Stop/disable `WSearch`; indexing policy |
| Clean | Delete index DB + related files |

### 39. MUICache / RecentApps / AppCompatFlags

| Field | Value |
|-------|-------|
| Hive | UsrClass `MuiCache`; NTUSER `RecentApps`, `AppCompatFlags\Layers`, `Persisted` |
| Clean | Delete keys per target user |

### 40. Syscache.hve

| Field | Value |
|-------|-------|
| Path | `C:\Windows\appcompat\Programs\Syscache.hve`, `RecentFileCache.bcf` |
| Clean | Secure delete after PcaSvc stopped |

### 41. Shell Icon / Thumb Cache

| Field | Value |
|-------|-------|
| Path | `%LocalAppData%\Microsoft\Windows\Explorer\iconcache_*.db`, `thumbcache_*.db` |
| Clean | Secure delete |

### 42. Targeted Execution EVTX

| Field | Value |
|-------|-------|
| Channels | TaskScheduler/Operational, Application-Experience Program-Telemetry/Inventory/Compatibility-Assistant |
| Clean | Disable + clear before full log purge |

### 43. EventTranscript.db

| Field | Value |
|-------|-------|
| Path | `%ProgramData%\Microsoft\Diagnosis\EventTranscript\EventTranscript.db` |
| Clean | Delete SQLite DB |

### 44. Windows Error Reporting (WER)

| Field | Value |
|-------|-------|
| Path | `%ProgramData%\Microsoft\Windows\WER`, user `%LocalAppData%\...\WER` |
| Clean | Clear report folders |

---

## Deep Traces (Tier B - optional)

| ID | Artifact | Path / action |
|----|----------|---------------|
| 45 | Servicing logs | `Panther`, `Logs\CBS`, `Logs\WindowsUpdate` |
| 46 | Print spooler | `System32\spool\PRINTERS` |
| 47 | Developer traces | `.gitconfig`, npm logs, VS Code/Cursor History |
| 48 | Delivery Optimization | `ProgramData\...\DeliveryOptimization\Cache` |
| 49 | SmartScreen cache | `%LocalAppData%\...\Safety\edge\remote` |

---

## Safety Boundaries (will NOT do)

- Wipe `$MFT` or `$LogFile` (filesystem destruction)
- Delete SAM user accounts or corrupt SYSTEM hive boot keys
- Terminate critical system processes
- LSASS memory injection
- Remove `MountedDevices` entries for boot volumes
