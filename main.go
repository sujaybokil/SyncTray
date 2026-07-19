package main

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/getlantern/systray"
)

const defaultWebUI = "http://127.0.0.1:8384"

var errApplicationExiting = errors.New("application is exiting")

var (
	syncthingMu         sync.Mutex
	syncthingCmd        *exec.Cmd
	syncthingGeneration uint64
	applicationExiting  bool
	logFile             *os.File
	webUIURL            string
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
// Currently supports one line: webui=http://127.0.0.1:8384
func loadConfig(exeDir string) {
	data, err := os.ReadFile(filepath.Join(exeDir, "synctray.conf"))
	if err != nil {
		webUIURL = defaultWebUI
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("Could not read synctray.conf: %v", err)
		}
		return
	}
	webUIURL = parseWebUIConfig(string(data))
	if webUIURL != defaultWebUI {
		log.Printf("Web UI URL loaded from config: %s", webUIURL)
	}
}

func parseWebUIConfig(data string) string {
	webUIURL := defaultWebUI
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "webui=") {
			val := strings.TrimPrefix(line, "webui=")
			if val != "" {
				webUIURL = val
			}
		}
	}
	return webUIURL
}

func loadIcon() []byte {
	exeDir, err := executableDir()
	if err != nil {
		log.Printf("Could not determine icon directory: %v", err)
		return makeIcon()
	}
	iconPath := filepath.Join(exeDir, "icon.ico")
	data, err := os.ReadFile(iconPath)
	if err != nil {
		log.Printf("Could not load icon.ico: %v", err)
		return makeIcon() // fall back to generated icon
	}
	return data
}

func onReady() {
	systray.SetIcon(loadIcon())
	systray.SetTooltip("Syncthing")

	mStatus := systray.AddMenuItem("● Starting...", "Syncthing status")
	mStatus.Disable()
	systray.AddSeparator()
	mOpenUI := systray.AddMenuItem("Open Web UI", "Open Syncthing in browser")
	systray.AddSeparator()
	mRestart := systray.AddMenuItem("Restart Syncthing", "Kill and restart Syncthing")
	mQuit := systray.AddMenuItem("Quit", "Stop Syncthing and exit")

	go startAndWatchSyncthing(mStatus)

	for {
		select {
		case <-mOpenUI.ClickedCh:
			openBrowser(webUIURL)
		case <-mRestart.ClickedCh:
			mStatus.SetTitle("● Restarting...")
			stopSyncthing()
			time.Sleep(1 * time.Second)
			startAndWatchSyncthing(mStatus)
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func onExit() {
	syncthingMu.Lock()
	applicationExiting = true
	syncthingMu.Unlock()
	stopSyncthing()
}

func startAndWatchSyncthing(mStatus *systray.MenuItem) {
	cmd, generation, err := startSyncthing()
	if err != nil {
		if !errors.Is(err, errApplicationExiting) {
			log.Printf("Failed to start syncthing: %v", err)
			mStatus.SetTitle("✖ Failed to start")
		}
		return
	}

	go watchSyncthing(cmd, generation, mStatus)

	// Give syncthing a moment to start up before marking it running.
	time.Sleep(2 * time.Second)
	if isCurrentSyncthing(cmd, generation) {
		mStatus.SetTitle("● Running")
	}
}

func startSyncthing() (*exec.Cmd, uint64, error) {
	// Look for syncthing.exe next to our own exe first, then in PATH
	exeDir, err := executableDir()
	if err != nil {
		return nil, 0, err
	}
	candidates := []string{
		filepath.Join(exeDir, "syncthing.exe"),
		"syncthing.exe",
		"syncthing",
	}

	var stPath string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			stPath = c
			break
		}
		if p, err := exec.LookPath(c); err == nil {
			stPath = p
			break
		}
	}

	if stPath == "" {
		return nil, 0, exec.ErrNotFound
	}

	log.Printf("Starting syncthing: %s", stPath)
	cmd := exec.Command(stPath, "--no-browser", "--no-restart", "--no-upgrade")
	cmd.Dir = exeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Redirect syncthing output to our log
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		return nil, 0, err
	}

	syncthingMu.Lock()
	if applicationExiting {
		syncthingMu.Unlock()
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Could not stop Syncthing while exiting: %v", err)
		} else {
			cmd.Wait()
		}
		return nil, 0, errApplicationExiting
	}
	syncthingGeneration++
	generation := syncthingGeneration
	syncthingCmd = cmd
	syncthingMu.Unlock()

	return cmd, generation, nil
}

func stopSyncthing() {
	syncthingMu.Lock()
	cmd := syncthingCmd
	syncthingCmd = nil
	syncthingGeneration++
	syncthingMu.Unlock()

	if cmd != nil && cmd.Process != nil {
		log.Println("Stopping syncthing")
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Could not stop syncthing: %v", err)
		}
	}
}

func watchSyncthing(cmd *exec.Cmd, generation uint64, mStatus *systray.MenuItem) {
	if err := cmd.Wait(); err != nil {
		log.Printf("Syncthing exited: %v", err)
	}

	syncthingMu.Lock()
	isCurrent := syncthingCmd == cmd && syncthingGeneration == generation
	if isCurrent {
		syncthingCmd = nil
	}
	syncthingMu.Unlock()

	if isCurrent {
		mStatus.SetTitle("✖ Stopped")
	}
}

func isCurrentSyncthing(cmd *exec.Cmd, generation uint64) bool {
	syncthingMu.Lock()
	defer syncthingMu.Unlock()
	return syncthingCmd == cmd && syncthingGeneration == generation
}

func openBrowser(url string) {
	if err := exec.Command("explorer", url).Start(); err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}

// makeIcon generates a simple teal square PNG icon at runtime.
// Replace this with //go:embed myicon.ico and real icon bytes for production.
func makeIcon() []byte {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	teal := color.RGBA{R: 0x00, G: 0x88, B: 0x88, A: 0xff}
	light := color.RGBA{R: 0x00, G: 0xbb, B: 0xbb, A: 0xff}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Simple rounded-rect-ish look: corners are transparent
			cx, cy := x-size/2, y-size/2
			if cx < 0 {
				cx = -cx
			}
			if cy < 0 {
				cy = -cy
			}
			if cx+cy > size/2+4 {
				continue // transparent corner
			}
			if x < 4 || y < 4 {
				img.Set(x, y, light)
			} else {
				img.Set(x, y, teal)
			}
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
