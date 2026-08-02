# Dependencies

SyncTray is deliberately small. The application has two direct Go dependencies, both pinned in `go.mod` and checksummed in `go.sum`.

| Dependency | Version | Purpose | License |
| --- | --- | --- | --- |
| [github.com/getlantern/systray](https://github.com/getlantern/systray) | v1.2.2 | Windows notification-area icon and menu | Apache-2.0 |
| [github.com/go-toast/toast](https://github.com/go-toast/toast) | v0.0.0-20190211030409-01e6764cf0a4 | Windows toast notifications | MIT |

Their transitive Go modules are recorded as indirect requirements in `go.mod`; use `go mod tidy` only with a reviewed dependency change and commit the resulting `go.mod` and `go.sum` together.

## External tools

| Tool | Required for | Distribution |
| --- | --- | --- |
| [Syncthing](https://syncthing.net/) | Running SyncTray's managed synchronization process | Installed separately by the user; not bundled or pinned by SyncTray |
| Go 1.21+ | Building and testing | Go toolchain |
| Inno Setup 6 | Building the Windows installer | Installer build only |

GitHub Dependabot checks Go modules and GitHub Actions monthly. Review updates for compatibility, licensing, and Windows behavior before merging.
