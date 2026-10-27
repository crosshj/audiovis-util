//go:build linux

package main

import (
	"flag"
	"log"
)

const canvasW, canvasH = 1280, 800

type options struct {
	benchmark, capture, full, selfTest, animate, static bool
	seconds                                             int
}

func main() {
	benchmark := flag.Bool("benchmark", false, "ignore physical input except Escape for repeatable samples")
	captureFrame := flag.Bool("capture", false, "save a framebuffer screenshot (adds a timing stall)")
	seconds := flag.Int("seconds", 60, "run duration, 0 until Escape")
	full := flag.Bool("full", false, "render HTML at 2560x1600")
	autotest := flag.Bool("self-test", false, "exercise custom controls and native callbacks automatically")
	animate := flag.Bool("animate-html", false, "animate spectrum bars")
	static := flag.Bool("static", false, "freeze animation and clock for reference capture")
	flag.Parse()
	if err := run(options{*benchmark, *captureFrame, *full, *autotest, *animate, *static, *seconds}); err != nil {
		log.Fatal(err)
	}
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
