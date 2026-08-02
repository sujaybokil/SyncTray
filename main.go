package main

import (
	_ "embed"
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

	"github.com/getlantern/systray"
	"github.com/go-toast/toast"
)

const defaultWebUI = "http://127.0.0.1:8384"

var (
	errApplicationExiting = errors.New("application is exiting")
)

type trayStatus int

const (
	statusStarting trayStatus = iota
	statusRunning
	statusRestarting
	statusStopped
	statusFailed
)

type config struct {
	webUIURL      string
	syncFolder    string
	syncthingPath string
}

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
	logFile             *os.File
	webUIURL            string
	syncFolder          string
	syncthingPath       string
	startMenuItem       *systray.MenuItem
	killMenuItem        *systray.MenuItem
	//go:embed icon.ico
	runningIcon []byte
	//go:embed assets/icon-starting.ico
	startingIcon []byte
	//go:embed assets/icon-stopped.ico
	stoppedIcon []byte
	//go:embed assets/icon-failed.ico
	failedIcon []byte
)

func main() {
	// Set up log file next to the exe
	exeDir, err := executableDir()
	if err != nil {
		log.Printf("Could not determine executable directory: %v", err)
		return
	}
	logPath := filepath.Join(exeDir, "synctray.log")
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Could not open log file %s: %v", logPath, err)
	}
	if logFile != nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	loadConfig(exeDir)
	systray.Run(onReady, onExit)
}

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

// loadConfig reads synctray.conf next to the exe for optional settings.
func loadConfig(exeDir string) {
	data, err := os.ReadFile(filepath.Join(exeDir, "synctray.conf"))
	if err != nil {
		webUIURL = defaultWebUI
		syncFolder = ""
		syncthingPath = ""
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("Could not read synctray.conf: %v", err)
		}
		return
	}
	settings := parseConfig(string(data))
	webUIURL = settings.webUIURL
	syncFolder = settings.syncFolder
	syncthingPath = settings.syncthingPath
	if webUIURL != defaultWebUI {
		log.Printf("Web UI URL loaded from config: %s", webUIURL)
	}
	if syncFolder != "" {
		log.Printf("Sync folder loaded from config: %s", syncFolder)
	}
}

func parseConfig(data string) config {
	settings := config{webUIURL: defaultWebUI}
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		key, val, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		val = strings.TrimSpace(val)
		switch strings.TrimSpace(key) {
		case "webui":
			if val != "" {
				settings.webUIURL = val
			}
		case "folder":
			if val != "" {
				settings.syncFolder = val
			}
		case "syncthing":
			settings.syncthingPath = val
		}
	}
	return settings
}

func parseWebUIConfig(data string) string {
	return parseConfig(data).webUIURL
}

func onReady() {
	systray.SetIcon(makeIcon(statusStarting))
	systray.SetTooltip("SyncTray — Syncthing is starting")

	mApp := systray.AddMenuItem("SyncTray", "Syncthing tray companion")
	mApp.Disable()
	mStatus := systray.AddMenuItem("Status: Starting...", "Current Syncthing status")
	mStatus.Disable()
	systray.AddSeparator()
	mOpenUI := systray.AddMenuItem("Open Syncthing", "Open the Syncthing web interface")
	var mOpenFolder *systray.MenuItem
	if syncFolder != "" {
		mOpenFolder = systray.AddMenuItem("Open Sync Folder", "Open your configured Syncthing folder")
	}
	mOpenLog := systray.AddMenuItem("Open Log", "Open the SyncTray log in Notepad")
	mSettings := systray.AddMenuItem("Edit Settings", "Edit synctray.conf; restart SyncTray to apply changes")
	mRestart := systray.AddMenuItem("Restart Syncthing", "Restart Syncthing when it is running")
	mKill := systray.AddMenuItem("Kill Syncthing", "Stop all running Syncthing processes")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit SyncTray", "Stop Syncthing and exit SyncTray")
	startMenuItem = mRestart
	killMenuItem = mKill
	setSyncthingActions(false)

	go startAndWatchSyncthing(mStatus, mRestart)
	go monitorSyncthing(mStatus)

	for {
		select {
		case <-mOpenUI.ClickedCh:
			openBrowser(webUIURL)
		case <-menuClick(mOpenFolder):
			openFolder(syncFolder)
		case <-mOpenLog.ClickedCh:
			openLog()
		case <-mSettings.ClickedCh:
			openSettings()
		case <-mRestart.ClickedCh:
			mRestart.Disable()
			mKill.Disable()
			if isWebUIAvailable(webUIURL) {
				setStatus(mStatus, statusRestarting, "Restarting...", "Syncthing is restarting")
				notify("Restarting Syncthing", "SyncTray is restarting Syncthing.")
				killAllSyncthing()
				time.Sleep(1 * time.Second)
			} else {
				setStatus(mStatus, statusStarting, "Starting...", "Syncthing is starting")
				notify("Starting Syncthing", "SyncTray is starting Syncthing.")
			}
			startAndWatchSyncthing(mStatus, mRestart)
		case <-mKill.ClickedCh:
			mRestart.Disable()
			mKill.Disable()
			setStatus(mStatus, statusRestarting, "Stopping...", "Syncthing is stopping")
			killAllSyncthing()
			time.Sleep(1 * time.Second)
			if isWebUIAvailable(webUIURL) {
				setStatus(mStatus, statusRunning, "Already running", "Syncthing is already running")
			} else {
				setStatus(mStatus, statusStopped, "Stopped", "Syncthing is stopped")
			}
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func menuClick(item *systray.MenuItem) <-chan struct{} {
	if item == nil {
		return nil
	}
	return item.ClickedCh
}

func setStatus(mStatus *systray.MenuItem, state trayStatus, status, tooltip string) {
	mStatus.SetTitle("Status: " + status)
	systray.SetTooltip("SyncTray — " + tooltip)
	if state == statusRunning {
		systray.SetIcon(makeIcon(state))
		setSyncthingActions(true)
		return
	}
	systray.SetIcon(makeIcon(state))
	if state == statusStopped || state == statusFailed {
		setSyncthingActions(false)
	}
}

func setSyncthingActions(running bool) {
	if startMenuItem == nil || killMenuItem == nil {
		return
	}
	if running {
		startMenuItem.SetTitle("Restart Syncthing")
		startMenuItem.Enable()
		killMenuItem.Enable()
		return
	}
	startMenuItem.SetTitle("Start Syncthing")
	startMenuItem.Enable()
	killMenuItem.Disable()
}

func onExit() {
	syncthingMu.Lock()
	applicationExiting = true
	syncthingMu.Unlock()
	stopSyncthing()
}

func startAndWatchSyncthing(mStatus, mRestart *systray.MenuItem) {
	defer mRestart.Enable()

	cmd, generation, _, err := startSyncthing()
	if err != nil {
		if !errors.Is(err, errApplicationExiting) {
			log.Printf("Failed to start syncthing: %v", err)
			if errors.Is(err, exec.ErrNotFound) {
				setStatus(mStatus, statusFailed, "Syncthing not found", "Add Syncthing to PATH or set its path in Settings")
				notify("Syncthing not found", "Add syncthing.exe to PATH, or set syncthing= to its full path in Edit Settings.")
			} else {
				setStatus(mStatus, statusFailed, "Unable to start", "Check the SyncTray log for details")
				notify("Syncthing could not start", "Check the SyncTray log for details.")
			}
		}
		return
	}

	go watchSyncthing(cmd, generation, mStatus)

	// Give syncthing a moment to start up before marking it running.
	time.Sleep(2 * time.Second)
	if markSyncthingReady(cmd, generation) {
		setStatus(mStatus, statusRunning, "Running", "Syncthing is running")
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

	// Redirect syncthing output to our log
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
	// A configured path is an explicit override. When it is empty, resolve PATH
	// on every launch so package-manager upgrades are picked up automatically.
	candidates := []string{syncthingPath, "syncthing.exe", "syncthing", filepath.Join(exeDir, "syncthing.exe")}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			path, err := filepath.Abs(candidate)
			if err != nil {
				continue
			}
			return launchFromPath(path), nil
		}
		if path, err := exec.LookPath(candidate); err == nil {
			path, err = filepath.Abs(path)
			if err != nil {
				continue
			}
			return launchFromPath(path), nil
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
	metadataPath := strings.TrimSuffix(shimPath, filepath.Ext(shimPath)) + ".shim"
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", nil, false
	}
	target, args, ok := parseScoopShim(string(data))
	if !ok {
		return "", nil, false
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return target, args, true
	}
	return resolvedTarget, args, true
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
	if target == "" {
		return "", nil, false
	}
	return target, args, true
}

func splitWindowsArgs(value string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	started := false
	for index := 0; index < len(value); {
		if (value[index] == ' ' || value[index] == '\t') && !inQuotes {
			if started {
				args = append(args, current.String())
				current.Reset()
				started = false
			}
			index++
			continue
		}

		backslashes := 0
		for index < len(value) && value[index] == '\\' {
			backslashes++
			index++
		}
		if index < len(value) && value[index] == '"' {
			current.WriteString(strings.Repeat("\\", backslashes/2))
			if backslashes%2 == 0 {
				inQuotes = !inQuotes
			} else {
				current.WriteByte('"')
			}
			started = true
			index++
			continue
		}
		if backslashes > 0 {
			current.WriteString(strings.Repeat("\\", backslashes))
			started = true
			continue
		}
		current.WriteByte(value[index])
		started = true
		index++
	}
	if started {
		args = append(args, current.String())
	}
	return args
}

func withConfigValue(data, key, value string) string {
	lines := strings.Split(strings.TrimRight(data, "\r\n"), "\n")
	for index, line := range lines {
		entryKey, _, found := strings.Cut(strings.TrimSpace(line), "=")
		if found && strings.TrimSpace(entryKey) == key {
			lines[index] = key + "=" + value
			return strings.Join(lines, "\n") + "\n"
		}
	}
	if len(lines) == 1 && lines[0] == "" {
		return key + "=" + value + "\n"
	}
	return strings.Join(append(lines, key+"="+value), "\n") + "\n"
}

func syncthingArgs(path, workDir string) []string {
	versionCommand := exec.Command(path, "--version")
	versionCommand.Dir = workDir
	version, err := versionCommand.Output()
	if err != nil {
		// Modern Syncthing releases use the explicit serve command. Prefer it
		// when a launcher on PATH cannot service the version probe.
		log.Printf("Could not determine Syncthing version; using v2 arguments: %v", err)
		return []string{"serve", "--no-browser", "--no-restart", "--no-upgrade"}
	}
	return syncthingArgsForVersion(string(version))
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

func watchSyncthing(cmd *exec.Cmd, generation uint64, mStatus *systray.MenuItem) {
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

	if isCurrent {
		if isWebUIAvailable(webUIURL) {
			setStatus(mStatus, statusRunning, "Already running", "Syncthing is already running")
			return
		}
		if !wasReady {
			setStatus(mStatus, statusFailed, "Unable to start", "Syncthing did not finish starting")
			notify("Syncthing could not start", "Check Syncthing's path and the SyncTray log for details.")
			return
		}
		setStatus(mStatus, statusStopped, "Stopped", "Syncthing has stopped")
		if !intentionalStop {
			notify("Syncthing stopped", "Syncthing exited unexpectedly. Use Restart Syncthing to start it again.")
		}
	}
}

func isWebUIAvailable(url string) bool {
	client := http.Client{Timeout: time.Second}
	response, err := client.Get(url)
	if err != nil {
		return false
	}
	response.Body.Close()
	return true
}

func monitorSyncthing(mStatus *systray.MenuItem) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if isApplicationExiting() || hasCurrentSyncthing() {
			continue
		}
		if isWebUIAvailable(webUIURL) {
			setStatus(mStatus, statusRunning, "Already running", "Syncthing is already running")
		} else {
			setStatus(mStatus, statusStopped, "Stopped", "Syncthing is stopped")
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

func isCurrentSyncthing(cmd *exec.Cmd, generation uint64) bool {
	syncthingMu.Lock()
	defer syncthingMu.Unlock()
	return syncthingCmd == cmd && syncthingGeneration == generation
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

func openBrowser(url string) {
	if err := exec.Command("explorer", url).Start(); err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}

func openFolder(folder string) {
	if err := exec.Command("explorer", folder).Start(); err != nil {
		log.Printf("Could not open sync folder: %v", err)
	}
}

func openLog() {
	exeDir, err := executableDir()
	if err != nil {
		log.Printf("Could not determine log path: %v", err)
		return
	}
	openInNotepad(filepath.Join(exeDir, "synctray.log"))
}

func openSettings() {
	exeDir, err := executableDir()
	if err != nil {
		log.Printf("Could not determine settings path: %v", err)
		return
	}

	configPath := filepath.Join(exeDir, "synctray.conf")
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		const defaultConfig = "# SyncTray settings\n# webui=http://127.0.0.1:8384\n# folder=C:\\Path\\To\\Your\\SyncFolder\n"
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			log.Printf("Could not create settings file: %v", err)
			return
		}
	}
	openInNotepad(configPath)
}

func openInNotepad(path string) {
	if err := exec.Command("notepad.exe", path).Start(); err != nil {
		log.Printf("Could not open %s: %v", path, err)
	}
}

func notify(title, message string) {
	go func() {
		notification := toast.Notification{
			AppID:   "SyncTray",
			Title:   title,
			Message: message,
			Audio:   toast.Default,
		}
		if err := notification.Push(); err != nil {
			log.Printf("Could not show notification: %v", err)
		}
	}()
}

func makeIcon(status trayStatus) []byte {
	switch status {
	case statusStarting, statusRestarting:
		return startingIcon
	case statusStopped:
		return stoppedIcon
	case statusFailed:
		return failedIcon
	default:
		return runningIcon
	}
}
