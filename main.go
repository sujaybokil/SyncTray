package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/getlantern/systray"
)

var (
	logFile       *os.File
	webUIURL      string
	syncFolder    string
	syncthingPath string
)

func main() {
	// Store operational files beside the executable so the per-user installer
	// and portable builds use the same predictable location.
	exeDir, err := executableDir()
	if err != nil {
		log.Printf("Could not determine executable directory: %v", err)
		return
	}
	configureLogging(filepath.Join(exeDir, "synctray.log"))
	defer closeLogFile()

	loadConfig(exeDir)
	if len(os.Args) == 2 && os.Args[1] == "check" {
		runDiagnostic(os.Stdout, exeDir)
		return
	}
	systray.Run(onReady, onExit)
}

func configureLogging(logPath string) {
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Could not open log file %s: %v", logPath, err)
		return
	}
	if os.Getenv("SYNCTRAY_DEBUG") == "1" {
		log.SetOutput(io.MultiWriter(logFile, os.Stderr))
		return
	}
	log.SetOutput(logFile)
}

func closeLogFile() {
	if logFile != nil {
		logFile.Close()
	}
}

func runDiagnostic(out io.Writer, exeDir string) {
	fmt.Fprintln(out, "SyncTray diagnostic")
	fmt.Fprintf(out, "Configuration directory: %s\n", exeDir)
	fmt.Fprintf(out, "Web UI: %s\n", webUIURL)
	if syncFolder == "" {
		fmt.Fprintln(out, "Sync folder: not configured")
	} else {
		fmt.Fprintf(out, "Sync folder: %s\n", syncFolder)
	}

	launch, err := findSyncthingLaunch(exeDir)
	if err != nil {
		fmt.Fprintln(out, "Syncthing: not found")
		return
	}
	fmt.Fprintf(out, "Syncthing: %s\n", launch.path)
	if len(launch.args) > 0 {
		fmt.Fprintf(out, "Shim arguments: %q\n", launch.args)
	}

	version, err := syncthingVersion(launch.path, exeDir)
	if err != nil {
		fmt.Fprintf(out, "Version probe: failed (%v)\n", err)
	} else {
		fmt.Fprintf(out, "Version: %s\n", version)
	}
	if isWebUIAvailable(webUIURL) {
		fmt.Fprintln(out, "Web UI reachability: available")
	} else {
		fmt.Fprintln(out, "Web UI reachability: unavailable")
	}
}

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}
