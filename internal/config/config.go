// Package config parses and updates SyncTray's small key=value settings file.
package config

import "strings"

const DefaultWebUI = "http://127.0.0.1:8384"

type Settings struct {
	WebUIURL      string
	SyncFolder    string
	SyncthingPath string
}

func Parse(data string) Settings {
	settings := Settings{WebUIURL: DefaultWebUI}
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "webui":
			if value != "" {
				settings.WebUIURL = value
			}
		case "folder":
			if value != "" {
				settings.SyncFolder = value
			}
		case "syncthing":
			settings.SyncthingPath = value
		}
	}
	return settings
}

func WithValue(data, key, value string) string {
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
