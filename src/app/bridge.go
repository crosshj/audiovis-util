//go:build linux

package main

import (
	"encoding/json"
	"github.com/crosshj/audiovis-util/src/internal/renderer"
	"github.com/crosshj/audiovis-util/src/internal/system"
	"log"
	"math"
	"time"
)

// bridge is the only layer that translates application commands and system data
// into browser messages. System services never construct JavaScript.
type bridge struct {
	view       *renderer.Renderer
	clock      *system.ClockSetter
	wifi       *system.WiFiMonitor
	demoValues map[string]float64
	calls      int
}

func newBridge(view *renderer.Renderer) *bridge {
	return &bridge{view: view, clock: system.StartClockSetter(), wifi: system.StartWiFi(), demoValues: map[string]float64{}}
}
func (b *bridge) close() { b.clock.Close(); b.wifi.Close() }
func (b *bridge) command(name string, value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return -1
	}
	switch name {
	case "clock.set":
		if b.clock.Request(value) {
			return 1
		}
		return -1
	case "wifi.setEnabled":
		if (value == 0 || value == 1) && b.wifi.Request(value == 1) {
			return value
		}
		return -1
	case "audio.volume", "playback.seek":
		value = max(0, min(100, value))
	case "playback.playing", "tools.tone", "tools.record", "tools.analyzer", "audio.mute":
		if value != 0 && value != 1 {
			return -1
		}
	case "audio.output", "audio.input":
		if value < 0 || value > 2 || math.Trunc(value) != value {
			return -1
		}
	default:
		return -1
	}
	b.demoValues[name] = value
	b.calls++
	return value
}
func (b *bridge) emit(topic string, value any) {
	payload, err := json.Marshal(value)
	must(err)
	name, err := json.Marshal(topic)
	must(err)
	b.view.Eval("App.receive(" + string(name) + "," + string(payload) + ")")
}
func (b *bridge) publishClock(now time.Time) { b.emit("clock.tick", system.CurrentClock(now)) }
func (b *bridge) drain() {
	select {
	case message := <-b.clock.Results:
		log.Printf("Date/time save completed: error=%q", message)
		b.emit("clock.saved", message)
		b.publishClock(time.Now())
	default:
	}
	select {
	case state := <-b.wifi.Updates:
		b.emit("wifi.status", state)
	default:
	}
}
