package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	fileconfig "syncthing-tray/internal/config"
)

const defaultWebUI = fileconfig.DefaultWebUI

type config struct {
	webUIURL      string
	syncFolder    string
	syncthingPath string
}

// loadConfig reads synctray.conf next to the executable for optional settings.
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
	settings := fileconfig.Parse(data)
	return config{
		webUIURL:      settings.WebUIURL,
		syncFolder:    settings.SyncFolder,
		syncthingPath: settings.SyncthingPath,
	}
}

func parseWebUIConfig(data string) string {
	return parseConfig(data).webUIURL
}

func withConfigValue(data, key, value string) string {
	return fileconfig.WithValue(data, key, value)
}
