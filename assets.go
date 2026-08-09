package main

import _ "embed"

var (
	//go:embed icon.ico
	runningIcon []byte
	//go:embed assets/icon-starting.ico
	startingIcon []byte
	//go:embed assets/icon-stopped.ico
	stoppedIcon []byte
	//go:embed assets/icon-failed.ico
	failedIcon []byte
)

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
