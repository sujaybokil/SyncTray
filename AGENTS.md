# Repository Guidelines

## Project Structure & Module Organization

SyncTray is a small Go application for Windows. The root package contains the
thin entry point (`main.go`), tray UI (`tray.go`), Syncthing lifecycle
management (`syncthing.go`), configuration integration (`config.go`), and
embedded assets (`assets.go`). Reusable, non-UI helpers belong in scoped
`internal/` packages: `internal/config` owns settings-file parsing and
`internal/windowscmd` owns Windows command-line parsing. `go.mod` and `go.sum`
define the module and dependencies.

Windows packaging lives at the repository root: `build.bat` builds the binary,
`installer.iss` defines the Inno Setup installer, and `setup-task.bat` creates
the optional logon task. `icon.ico` is the installer icon. User installation and
configuration are documented in `README.md`; local build and release details are
in `BUILDING.md`. Release automation is in `.github/workflows/release.yml`; tag
pushes matching `v*` build and publish the installer. Generated files such as
`synctray.exe`, `Output/`, and logs are ignored and must not be committed.

## Build, Test, and Development Commands

Use Go 1.21 or newer on Windows.

```bat
go mod tidy
go build -ldflags="-H windowsgui -s -w" -o synctray.exe .
build.bat
debug.bat
./tools/smoke-test.ps1
go test -race ./...
go vet ./...
```

The first two commands refresh dependencies and build the GUI executable;
`build.bat` performs the same build interactively. `debug.bat` creates a
console-enabled, symbol-rich executable and supports `debug.bat check` for a
non-destructive environment report. `tools/smoke-test.ps1` runs the automated
local smoke checks; add `-RunDiagnostic` when Syncthing discovery and Web UI
reachability should also be reported. The default tests must remain independent
of a locally installed Syncthing instance and user-specific paths. Run the race tests and `go vet ./...`
before submitting Go changes. To package a release, install Inno Setup 6 and run
`ISCC.exe /DAppVersion=1.0.0 installer.iss`; the installer appears in `Output/`.

## Coding Style & Naming Conventions

Format all Go changes with `gofmt`. Use tabs as produced by `gofmt`, concise
camelCase names for local variables and functions, and PascalCase only for
exported identifiers. Keep Windows-specific behavior explicit (for example,
`.exe` paths and `syscall.SysProcAttr`) and log operational failures rather than
silently discarding them. Store user-facing runtime files beside the executable,
consistent with `synctray.conf` and `synctray.log`.

## Testing Guidelines

There is currently no test framework or coverage target. Add focused Go tests as
`*_test.go` files alongside the code they cover, using the standard `testing`
package. Run `go test ./...` for all tests; manually verify tray actions and
Syncthing start/restart behavior on Windows when changing process or UI code.

## Commit & Pull Request Guidelines

Recent commits use short, imperative summaries such as `web ui fix` and
`updated installer`. Keep subjects brief and scoped to one change. Pull requests
should explain the user-visible effect, note validation performed, link relevant
issues, and include screenshots for tray, installer, or other UI changes.
