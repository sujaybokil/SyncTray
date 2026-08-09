package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDiagnosticReportsMissingSyncthing(t *testing.T) {
	t.Setenv("PATH", "")
	previousPath, previousURL, previousFolder := syncthingPath, webUIURL, syncFolder
	t.Cleanup(func() { syncthingPath, webUIURL, syncFolder = previousPath, previousURL, previousFolder })
	syncthingPath = filepath.Join(t.TempDir(), "missing.exe")
	webUIURL = "http://127.0.0.1:1"
	syncFolder = ""

	var output bytes.Buffer
	runDiagnostic(&output, t.TempDir())
	text := output.String()
	if !strings.Contains(text, "Syncthing: not found") {
		t.Fatalf("diagnostic output = %q", text)
	}
	if !strings.Contains(text, "Sync folder: not configured") {
		t.Fatalf("diagnostic output = %q", text)
	}
}

func TestSyncthingArgsFallsBackWhenVersionProbeFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.exe")
	want := []string{"serve", "--no-browser", "--no-restart", "--no-upgrade"}
	got := syncthingArgs(path, t.TempDir())
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("syncthingArgs() = %q, want %q", got, want)
	}
}

func TestConfigureLoggingWritesLogFile(t *testing.T) {
	previousOutput := log.Writer()
	previousFile := logFile
	t.Cleanup(func() {
		closeLogFile()
		logFile = previousFile
		log.SetOutput(previousOutput)
	})

	logPath := filepath.Join(t.TempDir(), "synctray.log")
	configureLogging(logPath)
	log.Print("diagnostic log test")
	closeLogFile()
	logFile = nil

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "diagnostic log test") {
		t.Errorf("log file = %q", data)
	}
}
