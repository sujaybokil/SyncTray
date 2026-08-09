package main

import (
	"encoding/binary"
	"sync"
	"syscall"
)

var (
	statusDotIcons     map[trayStatus][]byte
	statusDotIconsOnce sync.Once
)

func statusDotIcon(status trayStatus) []byte {
	statusDotIconsOnce.Do(func() {
		statusDotIcons = map[trayStatus][]byte{
			statusStarting:   makeStatusDot(246, 185, 0),
			statusRestarting: makeStatusDot(246, 185, 0),
			statusRunning:    makeStatusDot(0, 139, 204),
			statusStopped:    makeStatusDot(128, 128, 128),
			statusFailed:     makeStatusDot(220, 53, 69),
		}
	})
	return statusDotIcons[status]
}

// makeStatusDot creates a 32px, 32-bit ICO file. Windows menu-item icons are
// loaded as IMAGE_ICON, so PNG data is not sufficient on this platform.
func makeStatusDot(red, green, blue byte) []byte {
	const size = 32
	const imageOffset = 22
	const dibSize = 40
	const pixelBytes = size * size * 4
	const maskStride = ((size + 31) / 32) * 4
	const maskBytes = size * maskStride
	icon := make([]byte, imageOffset+dibSize+pixelBytes+maskBytes)
	binary.LittleEndian.PutUint16(icon[2:], 1)
	binary.LittleEndian.PutUint16(icon[4:], 1)
	icon[6], icon[7] = size, size
	binary.LittleEndian.PutUint16(icon[10:], 1)
	binary.LittleEndian.PutUint16(icon[12:], 32)
	binary.LittleEndian.PutUint32(icon[14:], dibSize+pixelBytes+maskBytes)
	binary.LittleEndian.PutUint32(icon[18:], imageOffset)

	dib := icon[imageOffset:]
	binary.LittleEndian.PutUint32(dib[0:], dibSize)
	binary.LittleEndian.PutUint32(dib[4:], size)
	binary.LittleEndian.PutUint32(dib[8:], size*2)
	binary.LittleEndian.PutUint16(dib[12:], 1)
	binary.LittleEndian.PutUint16(dib[14:], 32)

	pixels := dib[dibSize : dibSize+pixelBytes]
	backgroundRed, backgroundGreen, backgroundBlue := menuBackgroundColor()
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			offset := ((size-1-y)*size + x) * 4
			pixels[offset] = backgroundBlue
			pixels[offset+1] = backgroundGreen
			pixels[offset+2] = backgroundRed
			pixels[offset+3] = 255
			if x < 6 || x > 25 || y < 6 || y > 25 {
				continue
			}
			pixels[offset] = blue
			pixels[offset+1] = green
			pixels[offset+2] = red
			pixels[offset+3] = 255
		}
	}
	return icon
}

func menuBackgroundColor() (red, green, blue byte) {
	const colorMenu = 4
	user32 := syscall.NewLazyDLL("user32.dll")
	colorRef, _, _ := user32.NewProc("GetSysColor").Call(colorMenu)
	return byte(colorRef), byte(colorRef >> 8), byte(colorRef >> 16)
}
