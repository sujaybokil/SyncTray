# SyncTray

SyncTray is a lightweight Windows system-tray companion for [Syncthing](https://syncthing.net/). It starts Syncthing silently at logon and keeps the controls you need one click away. It supports both Syncthing v1 and the v2 `serve` command layout.

## Install

1. Download `synctray-setup.exe` from the [latest release](https://github.com/sujaybokil/SyncTray/releases/latest).
2. Run the installer—no administrator rights are needed.
3. SyncTray finds `syncthing.exe` on your `PATH` whenever it starts. When that path is a Scoop shim, it launches the shim's declared executable directly while retaining its configured arguments. If Syncthing is not on `PATH`, set a full path manually as described below.
4. Search for **SyncTray** in the Windows Start menu to launch it, or sign out and back in to use the scheduled logon start.

The installer places SyncTray in `%LOCALAPPDATA%\SyncTray\` and creates a logon task with a 30-second delay. To upgrade, install a newer release over the existing installation. To remove it, use **Add or Remove Programs**.

## Tray controls

The tray icon always uses the SyncTray sync mark: blue while running, amber while starting or restarting, gray when stopped, and red when SyncTray cannot start Syncthing. The tray menu provides:

| Item | Action |
| --- | --- |
| Open Syncthing | Opens the Syncthing web interface. |
| Open Sync Folder | Opens the optional folder configured below. |
| Open Log | Opens `%LOCALAPPDATA%\SyncTray\synctray.log`. |
| Edit Settings | Opens `synctray.conf` in Notepad. Restart SyncTray after saving changes. |
| Start / Restart Syncthing | Starts Syncthing when stopped, or restarts it when running; disabled while an action is in progress. |
| Kill Syncthing | Stops all running `syncthing.exe` processes. |
| Quit SyncTray | Stops Syncthing and exits the tray application. |

SyncTray displays Windows notifications only when Syncthing cannot start or stops unexpectedly.
If Syncthing is already running outside SyncTray, the status reads **Already running** and the web interface remains available.

## Settings

Use **Edit Settings** in the tray menu, or create `%LOCALAPPDATA%\SyncTray\synctray.conf`. Settings use one `key=value` entry per line:

```ini
# Optional: change the Syncthing web UI address.
webui=http://127.0.0.1:8384

# Optional: show Open Sync Folder in the tray menu.
folder=C:\Users\YourName\Sync

# Optional override. Leave this empty to resolve syncthing.exe from PATH on
# every start (recommended); set a full path only when it is not on PATH.
syncthing=C:\Path\To\syncthing.exe
```

If neither `PATH` nor this setting resolves Syncthing, SyncTray shows **Syncthing not found** and prompts you to update it using **Edit Settings**.

## Releases and development

Install SyncTray from [GitHub Releases](https://github.com/sujaybokil/SyncTray/releases/latest); releases contain the ready-to-run installer. Maintainers create a release by pushing a version tag such as `v1.0.0`.

For local builds, installer packaging, and release workflow details, see [BUILDING.md](BUILDING.md).
