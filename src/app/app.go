//go:build linux

package main

import (
	"fmt"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/crosshj/audiovis-util/src/internal/renderer"
	"github.com/crosshj/audiovis-util/src/internal/system"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func run(opt options) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	width, height := uint32(canvasW), uint32(canvasH)
	scale := float32(1)
	if opt.full {
		width *= 2
		height *= 2
		scale = 2
	}
	view, err := renderer.New(width, height)
	if err != nil {
		return err
	}
	defer view.Close()
	bridge := newBridge(view)
	defer bridge.close()
	must(view.BindCommand(bridge.command))
	eval := view.Eval
	window := view.Window
	eval("dashboard.reset()")
	if opt.full {
		eval("document.body.style.zoom='2'")
	}
	if opt.animate {
		eval("dashboard.animate(true)")
	}
	if opt.static {
		eval("document.body.dataset.static='1'")
	}
	bridge.publishClock(time.Now())
	charge, batteryState := system.BatteryStatus()
	bridge.emit("battery.status", map[string]any{"charge": charge, "status": batteryState})
	log.Printf("Battery: %d%% %s", charge, batteryState)
	recognizer := system.NewRecognizer(system.Point{X: 700, Y: 400})
	trackpad, err := system.StartTrackpad(canvasW, canvasH)
	if err != nil {
		return err
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	var usageStart syscall.Rusage
	must(syscall.Getrusage(syscall.RUSAGE_SELF, &usageStart))
	cpu := func(r syscall.Rusage) float64 {
		return float64(r.Utime.Sec+r.Stime.Sec) + float64(r.Utime.Usec+r.Stime.Usec)/1e6
	}
	started := time.Now()
	statusTime := started
	statusFrames := 0
	statusCPU := cpu(usageStart)
	lastStatus := -1
	frames, slow := 0, 0
	var renderTime, submitTime time.Duration
	tested := false
	captured := false
	running := true
	for running && (opt.seconds == 0 || time.Since(started) < time.Duration(opt.seconds)*time.Second) {
		select {
		case <-stop:
			running = false
			continue
		default:
		}
		frameStart := time.Now()
		bridge.drain()
		for draining := true; draining; {
			select {
			case e, ok := <-trackpad:
				if !ok {
					return fmt.Errorf("trackpad closed")
				}
				if opt.benchmark {
					continue
				}
				for _, gesture := range recognizer.Handle(e) {
					p := gesture.Position
					p.X = max(0, min(p.X, canvasW-1))
					p.Y = max(0, min(p.Y, canvasH-1))
					recognizer.SetPosition(p)
					x, y := int32(p.X*scale), int32(p.Y*scale)
					switch gesture.Kind {
					case system.PointerMoved:
						view.Mouse(0, x, y, 0)
					case system.PointerPressed:
						view.Mouse(1, x, y, 1)
					case system.PointerReleased:
						view.Mouse(2, x, y, 1)
					}
				}
			default:
				draining = false
			}
		}
		p := recognizer.Position()
		p.X = max(0, min(p.X, canvasW-1))
		p.Y = max(0, min(p.Y, canvasH-1))
		recognizer.SetPosition(p)
		var event sdl.Event
		for sdl.PollEvent(&event) {
			if opt.benchmark && event.Type != sdl.EVENT_QUIT {
				if event.Type == sdl.EVENT_KEY_DOWN && event.KeyboardEvent().Key == sdl.K_ESCAPE {
					running = false
				}
				continue
			}
			switch event.Type {
			case sdl.EVENT_QUIT:
				running = false
			case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
				e := event.KeyboardEvent()
				if e.Key == sdl.K_ESCAPE && e.Down {
					if eval("dashboard.escape()") == "true" {
						continue
					}
					running = false
					continue
				}
				kind := uint32(1)
				if e.Down {
					kind = 2
				}
				view.Key(kind, keyMods(e.Mod), virtualKey(e.Key), int32(e.Scancode), "", e.Repeat)
			case sdl.EVENT_TEXT_INPUT:
				view.Key(3, 0, 0, 0, event.TextInputEvent().Text, false)
			case sdl.EVENT_MOUSE_WHEEL:
				e := event.MouseWheelEvent()
				view.Scroll(int32(e.X*40), int32(e.Y*40))
			}
		}
		// Warp after pumping SDL events so its internal state agrees with the
		// authoritative position. Window coordinates can differ from GL pixels.
		windowW, windowH, err := window.Size()
		if err != nil {
			return err
		}
		mouseX, mouseY := p.X*float32(windowW)/canvasW, p.Y*float32(windowH)/canvasH
		window.WarpMouseIn(mouseX, mouseY)
		if opt.selfTest {
			_, actualX, actualY := sdl.GetMouseState()
			if actualX != mouseX || actualY != mouseY {
				return fmt.Errorf("SDL cursor mismatch: got %.2f,%.2f want %.2f,%.2f", actualX, actualY, mouseX, mouseY)
			}
		}
		age := time.Since(started)
		tick := int(age.Seconds())
		if tick != lastStatus && (!opt.static || lastStatus < 0) {
			var currentUsage syscall.Rusage
			must(syscall.Getrusage(syscall.RUSAGE_SELF, &currentUsage))
			elapsed := time.Since(statusTime).Seconds()
			liveFPS := float64(frames-statusFrames) / elapsed
			liveCPU := 100 * (cpu(currentUsage) - statusCPU) / elapsed
			eval(fmt.Sprintf("document.getElementById('status').textContent='GPU · %.1f FPS · %.1f%% CPU/core · Go calls: %d'", liveFPS, liveCPU, bridge.calls))
			statusTime = time.Now()
			statusFrames = frames
			statusCPU = cpu(currentUsage)
			if !opt.static {
				bridge.publishClock(time.Now())
				charge, batteryState := system.BatteryStatus()
				bridge.emit("battery.status", map[string]any{"charge": charge, "status": batteryState})
			}
			lastStatus = tick
		}
		if opt.selfTest && !tested && age > time.Second {
			if err := validate(view, scale, bridge.demoValues); err != nil {
				return err
			}
			tested = true
		}

		render, submit, err := view.Draw(age, opt.static, bridge.calls, p.X, p.Y)
		if err != nil {
			return err
		}
		renderTime += render
		submitTime += submit
		if opt.capture && !captured && age > 2*time.Second {
			view.Capture()
			captured = true
		}
		must(view.Swap())
		frames++
		if time.Since(frameStart) > 25*time.Millisecond {
			slow++
		}
	}
	if opt.selfTest && !tested {
		return fmt.Errorf("self-test did not run")
	}
	log.Printf("Dashboard state=%s", eval("dashboard.state()"))
	var usageEnd syscall.Rusage
	must(syscall.Getrusage(syscall.RUSAGE_SELF, &usageEnd))

	view.LogStats(frames, submitTime)
	log.Printf("RESOURCE cpu_one_core_pct=%.2f peak_rss_mib=%.2f", 100*(cpu(usageEnd)-cpu(usageStart))/time.Since(started).Seconds(), float64(usageEnd.Maxrss)/1024)
	log.Printf("RESULT seconds=%.3f frames=%d fps=%.2f frames_over_25ms=%d html_render_ms_per_frame=%.3f native_calls=%d", time.Since(started).Seconds(), frames, float64(frames)/time.Since(started).Seconds(), slow, renderTime.Seconds()*1000/float64(max(1, frames)), bridge.calls)
	return nil
}
