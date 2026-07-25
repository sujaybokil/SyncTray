# Building SyncTray

This document is for contributors and maintainers. Most users should install the ready-made installer from [GitHub Releases](https://github.com/sujaybokil/SyncTray/releases/latest).

## Requirements

- Windows
- Go 1.21 or newer — for example, `scoop install go`
- Inno Setup 6 — only for installer builds; for example, `scoop install innosetup`

## Build the application

From the repository root:

```bat
go mod tidy
go build -ldflags="-H windowsgui -s -w" -o synctray.exe .
```

Or run `build.bat` to perform the same dependency refresh and build interactively.

Run the checks before submitting a change:

```bat
go test ./...
go vet ./...
```

## Build the installer

With `synctray.exe` present, run:

```bat
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DAppVersion=1.0.0 installer.iss
```

The installer is written to `Output\synctray-setup.exe`.

## Publish a release

Push an annotated or lightweight version tag matching `v*`, for example:

```bat
git tag v1.0.0
git push origin v1.0.0
```

The GitHub Actions workflow builds the Windows executable and installer, then publishes `synctray-setup.exe` to the corresponding GitHub Release.
