//go:build linux

package system

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Read Linux power_supply data directly; unavailable charge stays unknown (-1).
func BatteryStatus() (int, string) {
	entries, _ := filepath.Glob("/sys/class/power_supply/*")
	for _, dir := range entries {
		read := func(name string) string {
			b, _ := os.ReadFile(filepath.Join(dir, name))
			return strings.TrimSpace(string(b))
		}
		if read("type") != "Battery" || read("present") == "0" {
			continue
		}
		charge, err := strconv.Atoi(read("capacity"))
		if err != nil || charge < 0 || charge > 100 {
			return -1, "Unknown"
		}
		return charge, read("status")
	}
	return -1, "Unavailable"
}
