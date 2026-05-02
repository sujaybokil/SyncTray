# SyncTray

A personal Windows system tray wrapper for [Syncthing](https://syncthing.net/) that I built for my own use. It starts Syncthing silently at login and lives quietly in the system tray.

## Install

1. Download `synctray-x.x.x.zip` from the [latest release](../../releases/latest) and unzip it
2. Right-click `install.ps1` → **Run with PowerShell**
3. Place `syncthing.exe` from [syncthing.net](https://syncthing.net/downloads/) into `%LOCALAPPDATA%\SyncTray\`
4. Either log off and back on, or run immediately:
   ```powershell
   Start-ScheduledTask -TaskName "SyncTray"
   ```

> SyncTray installs to `%LOCALAPPDATA%\SyncTray` — no admin rights needed.

## Uninstall

1. Right-click `uninstall.ps1` → **Run with PowerShell**

This removes all files and the logon task.

---

## Tray menu

| Item | Action |
|---|---|
| ● Running | Status indicator |
| Open Web UI | Opens the Syncthing web UI in your browser |
| Restart Syncthing | Kills and restarts the process |
| Quit | Stops Syncthing and exits |

Output is logged to `%LOCALAPPDATA%\SyncTray\synctray.log`.

## Custom Web UI URL (optional)

By default SyncTray opens `http://127.0.0.1:8384` which is Syncthing's default address. You only need to change this if you've configured Syncthing to run on a different port or address.

**1. Open the install folder**

```
%LOCALAPPDATA%\SyncTray\
```

Paste that path into File Explorer's address bar and press Enter.

**2. Create a new text file named `synctray.conf`**

Make sure it is saved as `synctray.conf` and not `synctray.conf.txt` — in File Explorer, go to View → Show → File name extensions to verify.

**3. Add your URL**

Open the file in Notepad and add a single line:

```
webui=http://127.0.0.1:8384
```

Replace the URL with your own, for example:

```
webui=http://127.0.0.1:9999
```

**4. Restart SyncTray**

Right-click the tray icon → **Quit**, then launch `synctray.exe` again. The new URL will be used when you click **Open Web UI**.

---

## Build from source

### Requirements

- [Go 1.21+](https://go.dev/dl/) — `scoop install go`

### Build

```bat
go mod tidy
go build -ldflags="-H windowsgui -s -w" -o synctray.exe .
```

Or just run `build.bat`.

### Register the logon task manually (without install script)

```bat
schtasks /create /tn "SyncTray" /tr "\"C:\path\to\synctray.exe\"" /sc ONLOGON /delay 0000:30 /it /f
```
