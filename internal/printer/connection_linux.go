//go:build linux

package printer

import (
	"fmt"
	"os"
	"path/filepath"
)

// Open opens a connection to the printer at the given device path.
func Open(devicePath string) (*Printer, error) {
	f, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("abrir impressora %s: %w", devicePath, err)
	}
	return &Printer{device: f, path: devicePath}, nil
}

// DetectPrinters returns available USB printer device paths on Linux.
func DetectPrinters() []string {
	matches, err := filepath.Glob("/dev/usb/lp*")
	if err != nil {
		return nil
	}
	return matches
}
