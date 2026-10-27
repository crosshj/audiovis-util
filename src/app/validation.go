//go:build linux

package main

import (
	"fmt"
	"github.com/crosshj/audiovis-util/src/internal/renderer"
	"github.com/crosshj/audiovis-util/src/internal/system"
	"log"
)

func validate(view *renderer.Renderer, scale float32, demoValues map[string]float64) error {
	eval := view.Eval
	point := func(id string, fraction float64) (int32, int32) {
		eval(fmt.Sprintf("window.testRect=document.getElementById(%q).getBoundingClientRect()", id))
		var x, y float32
		fmt.Sscanf(eval(fmt.Sprintf("(testRect.left+testRect.width*%f)+' '+(testRect.top+testRect.height/2)", fraction)), "%f %f", &x, &y)
		return int32(x * scale), int32(y * scale)
	}
	click := func(id string) {
		x, y := point(id, .5)
		view.Mouse(0, x, y, 0)
		view.Mouse(1, x, y, 1)
		view.Mouse(2, x, y, 1)
	}
	for _, slider := range []struct {
		id   string
		code string
	}{{"volume", "audio.volume"}, {"seek", "playback.seek"}} {
		x, y := point(slider.id, .25)
		view.Mouse(0, x, y, 0)
		view.Mouse(1, x, y, 1)
		x, y = point(slider.id, .75)
		view.Mouse(0, x, y, 1)
		view.Mouse(2, x, y, 1)
		view.Update()
		if demoValues[slider.code] != 75 {
			return fmt.Errorf("slider %s got %v want 75; JS=%s", slider.id, demoValues[slider.code], eval("dashboard.state()"))
		}
	}
	click("play")
	if demoValues["playback.playing"] != 1 {
		return fmt.Errorf("play callback failed")
	}
	click("play")
	click("output-device")
	if eval("document.getElementById('output-device').classList.contains('open')") != "true" {
		return fmt.Errorf("dropdown failed to open")
	}
	eval("document.querySelector('#output-device .option:nth-child(2)').id='test-option'")
	click("test-option")
	if demoValues["audio.output"] != 1 {
		return fmt.Errorf("dropdown callback failed")
	}
	eval("document.querySelector('#output-device .selection').textContent='Internal Speakers';App.command('audio.output',0);dashboard.reset()")
	if eval("document.querySelectorAll('input,select,button,textarea').length") != "0" {
		return fmt.Errorf("unexpected native form control")
	}
	log.Printf("PASS: custom slider drags -> Go (volume/seek=75), play/pause, dropdown selection; no native form controls; fonts=%s", eval("document.fonts.check('18px Inter')"))

	click("battery")
	if eval("document.getElementById('battery-popover').hidden") != "false" {
		return fmt.Errorf("battery popover failed to open")
	}
	if eval("(function(){var p=document.getElementById('battery-popover').getBoundingClientRect(),a=document.getElementById('battery').getBoundingClientRect();return p.left>=12&&p.right<=1268&&p.top>a.bottom})()") != "true" {
		return fmt.Errorf("popover anchor bounds failed")
	}
	if eval("dashboard.escape()") != "true" || eval("document.getElementById('battery-popover').hidden") != "true" {
		return fmt.Errorf("popover Escape failed")
	}
	click("battery")
	click("play")
	if eval("document.getElementById('battery-popover').hidden") != "true" {
		return fmt.Errorf("popover outside dismissal failed")
	}
	click("play")
	eval("dashboard.reset()")
	log.Print("PASS: battery popover opens, stays within viewport below anchor, closes on Escape and outside click")
	click("clock")
	if eval("document.getElementById('calendar-popover').hidden") != "false" {
		return fmt.Errorf("calendar did not open")
	}
	if eval("document.querySelectorAll('#calendar-days .today').length") != "1" {
		return fmt.Errorf("calendar today marker failed")
	}
	click("adjust-datetime")
	if eval("modals.isOpen()") != "true" {
		return fmt.Errorf("date/time modal did not open")
	}
	if eval("(function(){var r=document.querySelector('.modal-panel').getBoundingClientRect();return Math.abs(r.left+r.width/2-640)<1&&Math.abs(r.top+r.height/2-400)<1})()") != "true" {
		return fmt.Errorf("modal not centered")
	}
	eval("document.getElementById('date-year').focus()")
	view.Key(3, 0, 0, 0, "2024", false)
	if eval("document.getElementById('date-year').textContent") != "2024" {
		return fmt.Errorf("date editor keyboard input failed")
	}
	eval("document.getElementById('date-month').textContent='02';document.getElementById('date-day').textContent='30'")
	click("date-save")
	if eval("document.getElementById('date-error').textContent.length>0") != "true" {
		return fmt.Errorf("invalid date was accepted")
	}
	click("date-cancel")
	if eval("modals.isOpen()") != "false" {
		return fmt.Errorf("modal Cancel failed")
	}
	for _, invalid := range []float64{20260229000000, 20241301000000, 20241027240000} {
		if _, err := system.ParseClockValue(invalid); err == nil {
			return fmt.Errorf("backend accepted invalid date %v", invalid)
		}
	}
	if _, err := system.ParseClockValue(20240229010203); err != nil {
		return err
	}
	log.Print("PASS: current-month calendar, centered date modal, keyboard editing, invalid-date rejection, Cancel, backend leap-day validation; system time unchanged by self-test")

	return nil
}
