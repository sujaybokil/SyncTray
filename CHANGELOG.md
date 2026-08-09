# Changelog

All notable changes to SyncTray are documented here.

## [1.0.7] - 2026-08-09

### Status indicator

- Adds a native Windows status marker in the tray menu, with a compact square
  that visually blends with the system menu background.
- Keeps the marker and tray icon color synchronized for running, starting,
  stopped, and failed states.

## [1.0.6] - 2026-08-09

### Maintainability and diagnostics

- Splits the application into focused entry-point, tray, process, configuration,
  and asset source files, with reusable internal configuration and Windows
  command-line packages.
- Adds a console-enabled debug launcher and a read-only `check` diagnostic for
  configuration, Syncthing discovery, version probing, and Web UI reachability.
- Adds a portable local smoke-test script with isolated Go build caching.

### Quality and release engineering

- Expands unit and race-test coverage for configuration, argument parsing,
  diagnostics, version-probe fallback, logging, icons, and Web UI checks.
- Packages and retains installers in CI, validates release tags and changelog
  entries, publishes checksums, and pins workflow dependencies.
- Documents local debugging, smoke testing, checksums, and the project’s
  personal-use, freely available status.

## [1.0.5] - 2026-08-02

### SyncTray 1.0

- Starts Syncthing silently at Windows logon, with a scheduled-task startup and a Startup-folder fallback.
- Finds Syncthing from `PATH`, supports Scoop shims and their arguments, and accepts an explicit executable path.
- Supports Syncthing v1 and v2 launch commands.
- Provides tray actions for the web UI, optional sync folder, log, settings, start/restart, stop, and quit.
- Shows running, starting, stopped, and failed states with distinct icons and Windows notifications for failures and unexpected exits.
- Displays the MIT license agreement, installs per user without administrator rights, safely replaces a running prior installation, and preserves user settings during upgrades.
- Adds documented build, test, packaging, dependency, and release procedures.

### Quality

- Expands coverage for configuration parsing, version handling, settings updates, and Windows command-line parsing edge cases.
- Adds an MIT license, security reporting guidance, and automated dependency update checks.
