# Changelog

All notable changes to SyncTray are documented here.

## [1.0.4] - 2026-08-02

### SyncTray 1.0

- Starts Syncthing silently at Windows logon, with a scheduled-task startup and a Startup-folder fallback.
- Finds Syncthing from `PATH`, supports Scoop shims and their arguments, and accepts an explicit executable path.
- Supports Syncthing v1 and v2 launch commands.
- Provides tray actions for the web UI, optional sync folder, log, settings, start/restart, stop, and quit.
- Shows running, starting, stopped, and failed states with distinct icons and Windows notifications for failures and unexpected exits.
- Installs per user without administrator rights and preserves user settings during upgrades.
- Adds documented build, test, packaging, dependency, and release procedures.

### Quality

- Expands coverage for configuration parsing, version handling, settings updates, and Windows command-line parsing edge cases.
- Adds an MIT license, security reporting guidance, and automated dependency update checks.
