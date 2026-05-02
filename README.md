# SyncTray

A personal Windows system tray wrapper for [Syncthing](https://syncthing.net/) that I built for my own use. It starts Syncthing silently at login and lives quietly in the system tray.

## Install

1. Download `synctray-setup.exe` from the [latest release](https://github.com/sujaybokil/SyncTray/releases/latest)
2. Run `synctray-setup.exe` — no admin rights needed
3. Place `syncthing.exe` from [syncthing.net](https://syncthing.net/downloads/) into `%LOCALAPPDATA%\SyncTray\`
4. Either log off and back on, or launch immediately from the Start menu

> SyncTray installs to `%LOCALAPPDATA%\SyncTray` and registers a logon task with a 30 second delay automatically.

## Uninstall

Use **Add or Remove Programs**, search for SyncTray and uninstall. This removes all files and the logon task.

---

## Tray menu

| Item | Action |
| --- | --- |
| ● Running | Status indicator |
| Open Web UI | Opens the Syncthing web UI in your browser |
| Restart Syncthing | Kills and restarts the process |
| Quit | Stops Syncthing and exits |

Output is logged to `%LOCALAPPDATA%\SyncTray\synctray.log`.

---

## Custom Web UI URL (optional)

By default SyncTray opens `http://127.0.0.1:8384` which is Syncthing's default address. You only need to change this if you've configured Syncthing to run on a different port or address.

**1. Open the install folder**

Paste this into File Explorer's address bar and press Enter:
```
%LOCALAPPDATA%\SyncTray\
```

**2. Create a file named `synctray.conf`**

Make sure it is saved as `synctray.conf` and not `synctray.conf.txt` — in File Explorer go to View → Show → File name extensions to verify.

**3. Add your URL**

Open the file in Notepad and add a single line:
```
webui=http://127.0.0.1:8384
```

**4. Restart SyncTray**

Right-click the tray icon → **Quit**, then launch `synctray.exe` again. The new URL will be used when you click **Open Web UI**.

---

## Build from source

### Requirements

- [Go 1.21+](https://go.dev/dl/) — `scoop install go`
- [Inno Setup 6](https://jrsoftware.org/isdl.php) — `scoop install innosetup` (only needed to build the installer)

### Build the exe

```bat
go mod tidy
go build -ldflags="-H windowsgui -s -w" -o synctray.exe .
```

Or just run `build.bat`.

### Build the installer

```bat
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DAppVersion=1.0.0 installer.iss
```

Outputs `Output\synctray-setup.exe`.
