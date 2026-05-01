# SyncTray

A minimal Windows system tray wrapper for [Syncthing](https://syncthing.net/).  
Starts Syncthing silently at login and lives quietly in the system tray.

## Install from release

1. Download `synctray-setup.exe` from the [latest release](../../releases/latest)
2. Run it — no admin rights needed
3. It installs to `%LOCALAPPDATA%\SyncTray`, registers a logon task with a 30s delay, and optionally launches immediately

> **Syncthing is not bundled.** Download `syncthing.exe` from [syncthing.net](https://syncthing.net/downloads/) and place it in the same folder as `synctray.exe` (`%LOCALAPPDATA%\SyncTray\`).

## Build from source

### Requirements

- [Go 1.21+](https://go.dev/dl/) — `scoop install go`
- [Inno Setup 6](https://jrsoftware.org/isdl.php) — only needed to build the installer

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

### Register the logon task manually (without installer)

Run `setup-task.bat` once from the folder containing `synctray.exe`. Or manually:

```bat
schtasks /create /tn "SyncTray" /tr "\"C:\path\to\synctray.exe\"" /sc ONLOGON /delay 0000:30 /it /f
```

---

## Tray menu

| Item | Action |
|---|---|
| ● Running | Status indicator |
| Open Web UI | Opens http://127.0.0.1:8384 |
| Restart Syncthing | Kills and restarts the process |
| Quit | Stops Syncthing and exits |

Output is logged to `synctray.log` next to the exe.

## Uninstall

If installed via installer: use **Add or Remove Programs** — removes all files and the logon task.

If set up manually:
```bat
schtasks /delete /tn "SyncTray" /f
```
Then delete the folder.

---

## Publishing a release

Push a version tag — GitHub Actions builds and publishes automatically:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow builds `synctray.exe`, packages it with Inno Setup, and attaches `synctray-setup.exe` to the GitHub release. Go does not need to be installed on end-user machines.

## Repo structure

```
.github/workflows/release.yml   ← CI build & release
main.go                         ← tray app source
go.mod
installer.iss                   ← Inno Setup script
icon.ico                        ← tray + installer icon
build.bat                       ← local build helper
setup-task.bat                  ← manual logon task registration
.gitignore
```
