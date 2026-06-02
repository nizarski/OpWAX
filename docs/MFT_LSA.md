# MFT, $LogFile, LSASS

Native Go (`internal/util/`). No SDelete/klist/vaultcmd for core paths.

## MFT free space (`freescrub.go`)

SDelete-style: fill free MFT records + free clusters (1-pass zero), then delete temp files.  
Does **not** overwrite live `$MFT`.

`options.mft_free_space_scrub` - default **`false`** in `configs/default.json`.

## $LogFile (`mft_scrub.go`)

`chkdsk /F`: C: on reboot; other drives dismount + immediate run.

`options.logfile_reset_on_reboot`

## LSASS caches

| Target | Method |
|--------|--------|
| Kerberos | LSA APIs |
| Vault / cred dirs | Delete under `%LOCALAPPDATA%\Microsoft\` |
| cmdkey list | subprocess |
| CredEnumerate | advapi32 |

Full LSASS RAM → **reboot** (no injection).  
`options.lsass_scrub` (default `true`).

Locked deletes: `MoveFileExW` `MOVEFILE_DELAY_UNTIL_REBOOT`.

## Order

1. Artifact delete  
2. Free-space scrub (if enabled)  
3. chkdsk for `$LogFile` (if enabled)  
4. LSASS cache purge  
5. Reboot
