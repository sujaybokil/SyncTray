package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/getlantern/systray"
)

var (
	syncthingCmd *exec.Cmd
	logFile      *os.File
)

func main() {
	// Set up log file next to the exe
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	logPath := filepath.Join(exeDir, "synctray.log")
	logFile, _ = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	systray.Run(onReady, onExit)
}

func loadIcon() []byte {
	exePath, _ := os.Executable()
	iconPath := filepath.Join(filepath.Dir(exePath), "icon.ico")
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

	go func() {
		err := startSyncthing()
		if err != nil {
			log.Printf("Failed to start syncthing: %v", err)
			mStatus.SetTitle("✖ Failed to start")
			return
		}
		// Give syncthing a moment to start up
		time.Sleep(2 * time.Second)
		mStatus.SetTitle("● Running")

		// Watch the process
		go func() {
			if syncthingCmd != nil {
				syncthingCmd.Wait()
				mStatus.SetTitle("✖ Stopped")
			}
		}()
	}()

	for {
		select {
		case <-mOpenUI.ClickedCh:
			openBrowser("http://127.0.0.1:8384")
		case <-mRestart.ClickedCh:
			mStatus.SetTitle("● Restarting...")
			stopSyncthing()
			time.Sleep(1 * time.Second)
			err := startSyncthing()
			if err != nil {
				mStatus.SetTitle("✖ Failed to start")
			} else {
				time.Sleep(2 * time.Second)
				mStatus.SetTitle("● Running")
			}
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func onExit() {
	stopSyncthing()
}

func startSyncthing() error {
	// Look for syncthing.exe next to our own exe first, then in PATH
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
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
		return exec.ErrNotFound
	}

	log.Printf("Starting syncthing: %s", stPath)
	syncthingCmd = exec.Command(stPath, "--no-browser", "--no-restart", "--no-upgrade")
	syncthingCmd.Dir = exeDir
	syncthingCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Redirect syncthing output to our log
	if logFile != nil {
		syncthingCmd.Stdout = logFile
		syncthingCmd.Stderr = logFile
	}

	return syncthingCmd.Start()
}

func stopSyncthing() {
	if syncthingCmd != nil && syncthingCmd.Process != nil {
		log.Println("Stopping syncthing")
		syncthingCmd.Process.Kill()
		syncthingCmd = nil
	}
}

func openBrowser(url string) {
	exec.Command("cmd", "/c", "start", url).Start()
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
			if cx < 0 { cx = -cx }
			if cy < 0 { cy = -cy }
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
