//go:build linux

package system

import (
	"context"

	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"
)

type ClockState struct {
	Year   int    `json:"year"`
	Month  int    `json:"month"`
	Day    int    `json:"day"`
	Hour   int    `json:"hour"`
	Minute int    `json:"minute"`
	Second int    `json:"second"`
	Zone   string `json:"zone"`
}

func CurrentClock(now time.Time) ClockState {
	zone, _ := now.Zone()
	return ClockState{now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second(), zone}
}
func ParseClockValue(value float64) (time.Time, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value {
		return time.Time{}, fmt.Errorf("Invalid date or time")
	}
	text := fmt.Sprintf("%.0f", value)
	t, err := time.ParseInLocation("20060102150405", text, time.Local)
	if err != nil || len(text) != 14 || t.Year() < 2000 || t.Year() > 2099 || t.Format("20060102150405") != text {
		return time.Time{}, fmt.Errorf("Enter a valid date and time between 2000 and 2099")
	}
	return t, nil
}

// timedated updates CLOCK_REALTIME and the battery-backed RTC in one operation.
// No saved timestamp is replayed at boot: the RTC keeps advancing while off.
func saveSystemClock(ctx context.Context, t time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	b, err := exec.CommandContext(ctx, "/usr/bin/timedatectl", "--no-ask-password", "set-time", t.Format("2006-01-02 15:04:05")).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(b))
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("Could not save date/time: %s", message)
	}
	return nil
}

type ClockSetter struct {
	requests chan time.Time
	Results  chan string
	cancel   context.CancelFunc
}

func StartClockSetter() *ClockSetter {
	ctx, cancel := context.WithCancel(context.Background())
	s := &ClockSetter{make(chan time.Time, 1), make(chan string, 1), cancel}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-s.requests:
				message := ""
				if err := saveSystemClock(ctx, t); err != nil {
					message = err.Error()
				}
				select {
				case s.Results <- message:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return s
}
func (s *ClockSetter) Request(value float64) bool {
	t, err := ParseClockValue(value)
	if err != nil {
		return false
	}
	select {
	case s.requests <- t:
		return true
	default:
		return false
	}
}

func (s *ClockSetter) Close() { s.cancel() }
