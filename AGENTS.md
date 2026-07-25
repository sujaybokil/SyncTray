# Repository Guidelines

## Project Structure & Module Organization

SyncTray is a small Go application for Windows. `main.go` contains the tray UI,
Syncthing process management, configuration loading, and icon fallback logic.
`go.mod` and `go.sum` define the Go module and dependencies. Keep application
code in package `main` unless a new concern is substantial enough to warrant a
separate package.

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
go vet ./...
```

The first two commands refresh dependencies and build the GUI executable;
`build.bat` performs the same build interactively. Run `go vet ./...` before
submitting Go changes. To package a release, install Inno Setup 6 and run
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
