package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"syncthing-tray/internal/windowscmd"
)

const (
	versionProbeTimeout = 10 * time.Second
	webUITimeout        = time.Second
)

var errApplicationExiting = errors.New("application is exiting")

type syncthingLaunch struct {
	path string
	args []string
}

var (
	syncthingMu         sync.Mutex
	syncthingCmd        *exec.Cmd
	syncthingGeneration uint64
	intentionalStops    = make(map[*exec.Cmd]bool)
	syncthingReady      = make(map[*exec.Cmd]bool)
	applicationExiting  bool
	webUIClient         = &http.Client{Timeout: webUITimeout}
)

func onExit() {
	syncthingMu.Lock()
	applicationExiting = true
	syncthingMu.Unlock()
	trayDoneOnce.Do(func() {
		if trayDone != nil {
			close(trayDone)
		}
	})
	stopSyncthing()
}

func startAndWatchSyncthing() {
	cmd, generation, _, err := startSyncthing()
	if err != nil {
		if !errors.Is(err, errApplicationExiting) {
			log.Printf("Failed to start syncthing: %v", err)
			if errors.Is(err, exec.ErrNotFound) {
				setStatus(statusFailed, "Syncthing not found", "Add Syncthing to PATH or set its path in Settings")
				notify("Syncthing not found", "Add syncthing.exe to PATH, or set syncthing= to its full path in Edit Settings.")
			} else {
				setStatus(statusFailed, "Unable to start", "Check the SyncTray log for details")
				notify("Syncthing could not start", "Check the SyncTray log for details.")
			}
		}
		return
	}
	go watchSyncthing(cmd, generation)
	time.Sleep(2 * time.Second)
	if markSyncthingReady(cmd, generation) {
		setStatus(statusRunning, "Running", "Syncthing is running")
	}
}

func startSyncthing() (*exec.Cmd, uint64, string, error) {
	exeDir, err := executableDir()
	if err != nil {
		return nil, 0, "", err
	}
	launch, err := findSyncthingLaunch(exeDir)
	if err != nil {
		return nil, 0, "", err
	}
	arguments := append(append([]string{}, launch.args...), syncthingArgs(launch.path, exeDir)...)
	log.Printf("Starting syncthing: %s", launch.path)
	cmd := exec.Command(launch.path, arguments...)
	cmd.Dir = exeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	if err := cmd.Start(); err != nil {
		return nil, 0, "", err
	}
	syncthingMu.Lock()
	if applicationExiting {
		syncthingMu.Unlock()
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Could not stop Syncthing while exiting: %v", err)
		} else {
			cmd.Wait()
		}
		return nil, 0, "", errApplicationExiting
	}
	syncthingGeneration++
	generation := syncthingGeneration
	syncthingCmd = cmd
	syncthingMu.Unlock()
	return cmd, generation, launch.path, nil
}

func findSyncthingLaunch(exeDir string) (syncthingLaunch, error) {
	candidates := []string{syncthingPath, "syncthing.exe", "syncthing", filepath.Join(exeDir, "syncthing.exe")}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			if path, err := filepath.Abs(candidate); err == nil {
				return launchFromPath(path), nil
			}
		}
		if path, err := exec.LookPath(candidate); err == nil {
			if path, err = filepath.Abs(path); err == nil {
				return launchFromPath(path), nil
			}
		}
	}
	return syncthingLaunch{}, exec.ErrNotFound
}

func launchFromPath(path string) syncthingLaunch {
	if target, args, ok := scoopShimTarget(path); ok {
		log.Printf("Resolved Scoop shim %s to %s", path, target)
		return syncthingLaunch{path: target, args: args}
	}
	return syncthingLaunch{path: path}
}
func scoopShimTarget(shimPath string) (string, []string, bool) {
	data, err := os.ReadFile(strings.TrimSuffix(shimPath, filepath.Ext(shimPath)) + ".shim")
	if err != nil {
		return "", nil, false
	}
	target, args, ok := parseScoopShim(string(data))
	if !ok {
		return "", nil, false
	}
	resolved, err := filepath.EvalSymlinks(target)
	if err != nil {
		return target, args, true
	}
	return resolved, args, true
}
func parseScoopShim(data string) (string, []string, bool) {
	var target string
	var args []string
	for _, line := range strings.Split(data, "\n") {
		key, value, found := strings.Cut(strings.TrimSpace(line), "=")
		if !found {
			continue
		}
		switch strings.TrimSpace(key) {
		case "path":
			target = strings.Trim(strings.TrimSpace(value), "\"")
		case "args":
			args = splitWindowsArgs(strings.TrimSpace(value))
		}
	}
	return target, args, target != ""
}
func splitWindowsArgs(value string) []string { return windowscmd.SplitArgs(value) }

func syncthingArgs(path, workDir string) []string {
	version, err := syncthingVersion(path, workDir)
	if err != nil {
		log.Printf("Could not determine Syncthing version; using v2 arguments: %v", err)
		return []string{"serve", "--no-browser", "--no-restart", "--no-upgrade"}
	}
	return syncthingArgsForVersion(version)
}

func syncthingVersion(path, workDir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionProbeTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, path, "--version")
	command.Dir = workDir
	version, err := command.Output()
	return strings.TrimSpace(string(version)), err
}
func syncthingArgsForVersion(version string) []string {
	args := []string{"--no-browser", "--no-restart", "--no-upgrade"}
	if !strings.Contains(version, " v1.") {
		return append([]string{"serve"}, args...)
	}
	return args
}

func stopSyncthing() {
	syncthingMu.Lock()
	cmd := syncthingCmd
	syncthingCmd = nil
	syncthingGeneration++
	if cmd != nil {
		intentionalStops[cmd] = true
	}
	syncthingMu.Unlock()
	if cmd != nil && cmd.Process != nil {
		log.Println("Stopping syncthing")
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Could not stop syncthing: %v", err)
		}
	}
}
func killAllSyncthing() {
	stopSyncthing()
	cmd := exec.Command("taskkill", "/f", "/im", "syncthing.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		log.Printf("Could not kill remaining Syncthing processes: %v", err)
	}
}

func watchSyncthing(cmd *exec.Cmd, generation uint64) {
	if err := cmd.Wait(); err != nil {
		log.Printf("Syncthing exited: %v", err)
	}
	syncthingMu.Lock()
	isCurrent := syncthingCmd == cmd && syncthingGeneration == generation
	intentionalStop := intentionalStops[cmd]
	delete(intentionalStops, cmd)
	wasReady := syncthingReady[cmd]
	delete(syncthingReady, cmd)
	if isCurrent {
		syncthingCmd = nil
	}
	syncthingMu.Unlock()
	if !isCurrent {
		return
	}
	if isWebUIAvailable(webUIURL) {
		setStatus(statusRunning, "Already running", "Syncthing is already running")
		return
	}
	if !wasReady {
		setStatus(statusFailed, "Unable to start", "Syncthing did not finish starting")
		notify("Syncthing could not start", "Check Syncthing's path and the SyncTray log for details.")
		return
	}
	setStatus(statusStopped, "Stopped", "Syncthing has stopped")
	if !intentionalStop {
		notify("Syncthing stopped", "Syncthing exited unexpectedly. Use Restart Syncthing to start it again.")
	}
}

func isWebUIAvailable(url string) bool {
	response, err := webUIClient.Get(url)
	if err != nil {
		return false
	}
	response.Body.Close()
	return true
}
func monitorSyncthing() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if isApplicationExiting() || hasCurrentSyncthing() {
			continue
		}
		if isWebUIAvailable(webUIURL) {
			setStatus(statusRunning, "Already running", "Syncthing is already running")
		} else {
			setStatus(statusStopped, "Stopped", "Syncthing is stopped")
		}
	}
}
func isApplicationExiting() bool {
	syncthingMu.Lock()
	defer syncthingMu.Unlock()
	return applicationExiting
}
func hasCurrentSyncthing() bool {
	syncthingMu.Lock()
	defer syncthingMu.Unlock()
	return syncthingCmd != nil
}
func markSyncthingReady(cmd *exec.Cmd, generation uint64) bool {
	syncthingMu.Lock()
	defer syncthingMu.Unlock()
	if syncthingCmd != cmd || syncthingGeneration != generation {
		return false
	}
	syncthingReady[cmd] = true
	return true
}
