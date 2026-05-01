# syncthing-tray

A minimal Windows system tray wrapper for [Syncthing](https://syncthing.net/).

- Starts Syncthing automatically in the background
- Lives in the system tray with a right-click menu
- Launches at Windows logon via Task Scheduler

## Tray menu

| Item | Action |
|---|---|
| ● Running | Status indicator (disabled) |
| Open Web UI | Opens http://127.0.0.1:8384 in browser |
| Restart Syncthing | Kills and restarts the process |
| Quit | Stops Syncthing and exits the tray app |

Syncthing output is logged to `syncthing-tray.log` next to the exe.

## Requirements

- [Go 1.21+](https://go.dev/dl/) (only for building)
- [Syncthing](https://syncthing.net/downloads/) — `syncthing.exe`

## Build & install

```bat
:: 1. Build the exe (only needed once)
build.bat

:: 2. Register the logon task (only needed once)
setup-task.bat

:: 3. Optionally run it right now
schtasks /run /tn SyncthingTray
```

## File layout

Put everything in one folder:

```
C:\Users\You\Apps\syncthing\
  syncthing.exe          ← downloaded from syncthing.net
  syncthing-tray.exe     ← built by build.bat
  syncthing-tray.log     ← created automatically at runtime
```

`syncthing-tray.exe` first looks for `syncthing.exe` in its own directory,
then falls back to searching `PATH`.

## Custom icon

Replace the `makeIcon()` call in `main.go` with an embedded `.ico` file:

```go
//go:embed icon.ico
var iconBytes []byte

// then in onReady():
systray.SetIcon(iconBytes)
```

## Uninstall

```bat
schtasks /delete /tn SyncthingTray /f
```

Then delete the folder.
