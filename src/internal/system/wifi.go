//go:build linux

package system

import (
	"context"

	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type WiFiState struct {
	Available bool   `json:"available"`
	Enabled   bool   `json:"enabled"`
	Blocked   bool   `json:"blocked"`
	Connected bool   `json:"connected"`
	SSID      string `json:"ssid"`
	Error     string `json:"error"`
}

type WiFiMonitor struct {
	requests chan bool
	Updates  chan WiFiState
	cancel   context.CancelFunc
}

func readSys(path string) string { b, _ := os.ReadFile(path); return strings.TrimSpace(string(b)) }
func wlanRadio() string {
	paths, _ := filepath.Glob("/sys/class/rfkill/*")
	for _, p := range paths {
		if readSys(filepath.Join(p, "type")) == "wlan" {
			return p
		}
	}
	return ""
}
func wlanInterface() string {
	paths, _ := filepath.Glob("/sys/class/net/*/wireless")
	if len(paths) == 0 {
		return ""
	}
	return filepath.Base(filepath.Dir(paths[0]))
}
func wpaCommand(ctx context.Context, iface string, command string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	b, err := exec.CommandContext(ctx, "/usr/sbin/wpa_cli", "-i", iface, command).CombinedOutput()
	return string(b), err
}
func readWiFi(ctx context.Context) WiFiState {
	radio := wlanRadio()
	iface := wlanInterface()
	s := WiFiState{Available: radio != "" && iface != ""}
	if !s.Available {
		return s
	}
	soft, hard := readSys(filepath.Join(radio, "soft")), readSys(filepath.Join(radio, "hard"))
	if soft == "" || hard == "" {
		s.Error = "Cannot read Wi-Fi radio state"
		return s
	}
	s.Blocked = hard == "1"
	s.Enabled = soft == "0" && !s.Blocked
	if !s.Enabled {
		return s
	}
	output, err := wpaCommand(ctx, iface, "status")
	if err != nil {
		s.Error = "Connection status unavailable"
		return s
	}
	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "ssid":
			s.SSID = value
		case "wpa_state":
			s.Connected = value == "COMPLETED"
		}
	}
	return s
}
func setWiFi(ctx context.Context, enabled bool) error {
	radio := wlanRadio()
	if radio == "" {
		return fmt.Errorf("Wi-Fi radio unavailable")
	}
	if enabled && readSys(filepath.Join(radio, "hard")) == "1" {
		return fmt.Errorf("Wi-Fi is blocked by hardware")
	}
	value := "1"
	if enabled {
		value = "0"
	}
	if err := os.WriteFile(filepath.Join(radio, "soft"), []byte(value), 0644); err != nil {
		return fmt.Errorf("Cannot change Wi-Fi radio: %w", err)
	}
	if enabled {
		// The saved profile and DHCP setup remain in place across a radio block.
		// rfkill unblock notifies wpa_supplicant; reconnect also requests association.
		if iface := wlanInterface(); iface != "" {
			_, _ = wpaCommand(ctx, iface, "reconnect")
		}
	}
	return nil
}
func StartWiFi() *WiFiMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	m := &WiFiMonitor{make(chan bool, 1), make(chan WiFiState, 1), cancel}
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		var last WiFiState
		var lastError string
		publish := func() {
			s := readWiFi(ctx)
			if lastError != "" {
				s.Error = lastError
			}
			if s != last {
				log.Printf("Wi-Fi: enabled=%t connected=%t blocked=%t error=%q", s.Enabled, s.Connected, s.Blocked, s.Error)
				last = s
			}
			select {
			case <-m.Updates:
			default:
			}
			select {
			case m.Updates <- s:
			default:
			}
		}
		publish()
		for {
			select {
			case <-ctx.Done():
				return
			case enabled := <-m.requests:
				lastError = ""
				if err := setWiFi(ctx, enabled); err != nil {
					lastError = err.Error()
				}
				publish()
			case <-ticker.C:
				publish()
			}
		}
	}()
	return m
}
func (m *WiFiMonitor) Request(enabled bool) bool {
	select {
	case m.requests <- enabled:
		return true
	default:
		return false
	}
}

func (m *WiFiMonitor) Close() { m.cancel() }
