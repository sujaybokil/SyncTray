package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/go-toast/toast"
)

type trayStatus int

const (
	statusStarting trayStatus = iota
	statusRunning
	statusRestarting
	statusStopped
	statusFailed
)

type trayStatusUpdate struct {
	state   trayStatus
	status  string
	tooltip string
}

var (
	startMenuItem *systray.MenuItem
	killMenuItem  *systray.MenuItem
	statusUpdates chan trayStatusUpdate
	trayDone      chan struct{}
	trayDoneOnce  sync.Once
)

func onReady() {
	statusUpdates = make(chan trayStatusUpdate, 16)
	trayDone = make(chan struct{})

	mApp := systray.AddMenuItem("SyncTray", "Syncthing tray companion")
	mApp.Disable()
	mStatus := systray.AddMenuItem("Status: Starting...", "Current Syncthing status")
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
	applyStatus(mStatus, trayStatusUpdate{state: statusStarting, status: "Starting...", tooltip: "Syncthing is starting"})

	go startAndWatchSyncthing()
	go monitorSyncthing()

	for {
		select {
		case update := <-statusUpdates:
			applyStatus(mStatus, update)
		case <-mStatus.ClickedCh:
			// Enabled so Windows preserves the status dot's color; intentionally inert.
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
				setStatus(statusRestarting, "Restarting...", "Syncthing is restarting")
				notify("Restarting Syncthing", "SyncTray is restarting Syncthing.")
				killAllSyncthing()
				time.Sleep(time.Second)
			} else {
				setStatus(statusStarting, "Starting...", "Syncthing is starting")
				notify("Starting Syncthing", "SyncTray is starting Syncthing.")
			}
			startAndWatchSyncthing()
		case <-mKill.ClickedCh:
			mRestart.Disable()
			mKill.Disable()
			setStatus(statusRestarting, "Stopping...", "Syncthing is stopping")
			killAllSyncthing()
			time.Sleep(time.Second)
			if isWebUIAvailable(webUIURL) {
				setStatus(statusRunning, "Already running", "Syncthing is already running")
			} else {
				setStatus(statusStopped, "Stopped", "Syncthing is stopped")
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

func setStatus(state trayStatus, status, tooltip string) {
	select {
	case statusUpdates <- trayStatusUpdate{state: state, status: status, tooltip: tooltip}:
	case <-trayDone:
	}
}

func applyStatus(mStatus *systray.MenuItem, update trayStatusUpdate) {
	mStatus.SetIcon(statusDotIcon(update.state))
	mStatus.SetTitle("Status: " + update.status)
	systray.SetTooltip("SyncTray — " + update.tooltip)
	systray.SetIcon(makeIcon(update.state))
	if update.state == statusRunning {
		setSyncthingActions(true)
	} else if update.state == statusStopped || update.state == statusFailed {
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
		notification := toast.Notification{AppID: "SyncTray", Title: title, Message: message, Audio: toast.Default}
		if err := notification.Push(); err != nil {
			log.Printf("Could not show notification: %v", err)
		}
	}()
}
