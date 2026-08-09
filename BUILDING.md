# Building SyncTray

This document is for contributors and maintainers. SyncTray is a personal-use
project that is freely available under the MIT License; most users should
install the ready-made installer from [GitHub Releases](https://github.com/sujaybokil/SyncTray/releases/latest).

## Requirements

- Windows
- Go 1.21 or newer — for example, `scoop install go`
- Inno Setup 6 — only for installer builds; CI uses the pinned Chocolatey package version 6.7.1
- Syncthing — only to exercise the installed application; it is not required to compile or test SyncTray

## Source layout

- `main.go` is the executable entry point and owns runtime file setup.
- `tray.go`, `syncthing.go`, `config.go`, and `assets.go` separate the UI,
  process lifecycle, configuration integration, and embedded icons.
- `internal/config` contains the settings-file parser and writer.
- `internal/windowscmd` contains Windows command-line parsing used for Scoop
  shim arguments.

Keep Windows UI and process coordination in the root package. Put reusable,
non-UI implementation details under `internal/` so they cannot become external
API by accident.

## Build the application

From the repository root:

```bat
go mod tidy
go build -trimpath -ldflags="-H windowsgui -s -w" -o synctray.exe .
```

Or run `build.bat` to perform the same dependency refresh and build interactively.

## Debug locally

Run `debug.bat` to build `synctray-debug.exe` without the GUI subsystem or
symbol stripping, enable console logging, and start it. Pass `check` as an
argument (`debug.bat check`) for a non-destructive diagnostic report covering
configuration, Syncthing discovery, version probing, and Web UI reachability.
Operational messages are written to both the console and `synctray.log` next to
the executable. Do not use the debug executable for releases.

The debug executable uses the same configuration and log location as the
release executable: the directory containing the executable. Copy it beside a
test build, or run it from an isolated directory, when investigating a user's
configuration without touching the installed application.

Run the repeatable smoke checks with:

```powershell
.\tools\smoke-test.ps1
.\tools\smoke-test.ps1 -RunDiagnostic
```

The optional diagnostic does not start the tray application or terminate any
processes; it reports the local Syncthing setup instead. The default smoke
checks require only Windows and Go. A local Syncthing installation is needed
only when using the optional diagnostic to confirm discovery and Web UI status.

Run the checks before submitting a change:

```bat
go test ./...
go test -race ./...
go vet ./...
```

## Build the installer

With `synctray.exe` present, run:

```bat
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DAppVersion=1.0.6 installer.iss
```

The installer is written to `Output\synctray-setup.exe`.

## Publish a release

Add a matching `## [MAJOR.MINOR.PATCH]` section to `CHANGELOG.md`, then push an
annotated or lightweight version tag matching `vMAJOR.MINOR.PATCH`, for example:

```bat
git tag v1.0.6
git push origin v1.0.6
```

Pull requests and pushes to `master` run tests (including the race detector), vet, build, package, and retain the installer as a CI artifact. A release tag must use `vMAJOR.MINOR.PATCH`; the release workflow repeats those checks, publishes the installer with a SHA-256 checksum, and uses that version's changelog section as the GitHub Release notes. Re-running a release workflow updates its assets and notes safely.

See [DEPENDENCIES.md](DEPENDENCIES.md) before updating or adding third-party code.
