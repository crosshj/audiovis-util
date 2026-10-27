package main

import (
	_ "embed"
	"fmt"
	"image"
	stdcolor "image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/crosshj/audio-utility/src/input"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed assets/fonts/Inter-Variable.ttf
var interFont []byte

const (
	canvasW = 1366
	canvasH = 768
	windowW = 1280
	windowH = 800

	defaultCursorSize = 60
	minCursorSize     = 32
	maxCursorSize     = 88
	cursorSizeStep    = 4
	cursorSliderX     = 920
	cursorSliderW     = 140
	volumeSliderX     = 160
	volumeSliderY     = 335
	volumeSliderW     = 320
	playbackSliderX   = 596
	playbackSliderY   = 670
	playbackSliderW   = 660
)

type dragControl uint8

const (
	dragNone dragControl = iota
	dragCursorSize
	dragVolume
	dragPlayback
)

type color struct{ r, g, b, a uint8 }

var (
	background = color{18, 22, 29, 255}
	panel      = color{24, 29, 37, 255}
	panelAlt   = color{29, 35, 44, 255}
	outline    = color{56, 65, 78, 255}
	grid       = color{37, 45, 56, 255}
	muted      = color{161, 173, 190, 255}
	white      = color{228, 234, 242, 255}
	blue       = color{77, 151, 242, 255}
	green      = color{112, 213, 129, 255}
	orange     = color{239, 144, 79, 255}
)

func main() {
	defer loadSDL()()
	defer sdl.Quit()
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatal(err)
	}
	typeface, err := opentype.Parse(interFont)
	if err != nil {
		log.Fatal(err)
	}

	width, height, flags := windowW, windowH, sdl.WINDOW_HIGH_PIXEL_DENSITY
	fullscreen := os.Getenv("SDL_VIDEODRIVER") == "kmsdrm"
	if fullscreen {
		sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengles2")
		width, height, flags = 0, 0, sdl.WINDOW_FULLSCREEN
	}
	window, renderer, err := sdl.CreateWindowAndRenderer("Audio Utility", width, height, flags)
	if err != nil {
		log.Fatal(err)
	}
	defer window.Destroy()
	defer renderer.Destroy()
	// The target adapter reads raw multitouch contacts and physical click
	// press/release events from evdev. The dev preview uses SDL mouse events.
	var trackpadEvents <-chan input.Event
	recognizer := input.NewRecognizer(input.Point{X: canvasW / 2, Y: canvasH / 2})
	if fullscreen {
		if err := sdl.HideCursor(); err != nil {
			log.Fatalf("could not hide SDL's absolute trackpad cursor: %v", err)
		}
		trackpadEvents, err = startTrackpadPointer()
		if err != nil {
			log.Fatalf("could not initialize relative trackpad input: %v", err)
		}
		log.Print("Gesture input ready: trackpad contacts and click buttons feed the UI gesture layer.")
	}
	cursorSize := defaultCursorSize
	var cursor *sdl.Cursor
	var cursorTexture *sdl.Texture
	if fullscreen {
		cursorTexture, err = createCursorTexture(renderer, cursorSize)
	} else {
		cursor, err = createMouseCursor(cursorSize)
		if err == nil {
			err = sdl.SetCursor(cursor)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if cursor != nil {
			cursor.Destroy()
		}
		if cursorTexture != nil {
			cursorTexture.Destroy()
		}
	}()
	log.Printf("Mouse cursor size: %d px (default)", cursorSize)
	volume := 80
	playbackProgress := float32(0)
	text := &textPainter{renderer: renderer, font: typeface, cache: make(map[textKey]*sdl.Texture)}
	defer text.Close()
	if w, h, err := renderer.RenderOutputSize(); err == nil {
		name, _ := renderer.Name()
		log.Printf("SDL renderer %q, render surface %dx%d pixels", name, w, h)
		if fullscreen && name != "opengles2" {
			log.Fatalf("expected SDL OpenGL ES renderer on KMSDRM, got %q", name)
		}
	}
	testDuration := time.Duration(0)
	if seconds, err := strconv.Atoi(os.Getenv("AUDIO_UTILITY_TEST_SECONDS")); err == nil && seconds > 0 {
		testDuration = time.Duration(seconds) * time.Second
		log.Printf("test mode: returning to console after %s", testDuration)
	}
	started := time.Now()
	activeDrag := dragNone
	setCursorSize := func(nextSize int) {
		if nextSize == cursorSize {
			return
		}
		if fullscreen {
			nextTexture, err := createCursorTexture(renderer, nextSize)
			if err != nil {
				log.Printf("Could not create %d px cursor: %v", nextSize, err)
				return
			}
			cursorTexture.Destroy()
			cursorTexture = nextTexture
		} else {
			nextCursor, err := createMouseCursor(nextSize)
			if err != nil {
				log.Printf("Could not create %d px cursor: %v", nextSize, err)
				return
			}
			if err := sdl.SetCursor(nextCursor); err != nil {
				nextCursor.Destroy()
				log.Printf("Could not set %d px cursor: %v", nextSize, err)
				return
			}
			cursor.Destroy()
			cursor = nextCursor
		}
		cursorSize = nextSize
		log.Printf("Mouse cursor size changed: %d px", cursorSize)
	}
	handleGesture := func(gesture input.Gesture) {
		switch gesture.Kind {
		case input.DragStarted:
			if gesture.Button != input.LeftButton {
				break
			}
			switch {
			case inCursorSlider(gesture.Position.X, gesture.Position.Y):
				activeDrag = dragCursorSize
				setCursorSize(cursorSizeAt(gesture.Position.X))
			case inVolumeSlider(gesture.Position.X, gesture.Position.Y):
				activeDrag = dragVolume
				volume = volumeAt(gesture.Position.X)
			case inPlaybackSlider(gesture.Position.X, gesture.Position.Y):
				activeDrag = dragPlayback
				playbackProgress = progressAt(gesture.Position.X, playbackSliderX, playbackSliderW)
			}
		case input.DragMoved:
			switch activeDrag {
			case dragCursorSize:
				setCursorSize(cursorSizeAt(gesture.Position.X))
			case dragVolume:
				volume = volumeAt(gesture.Position.X)
			case dragPlayback:
				playbackProgress = progressAt(gesture.Position.X, playbackSliderX, playbackSliderW)
			}
		case input.DragEnded:
			if gesture.Button == input.LeftButton {
				activeDrag = dragNone
			}
		}
	}

	for {
		var inputs []input.Event
		if fullscreen {
			draining := true
			for draining {
				select {
				case inputEvent := <-trackpadEvents:
					inputs = append(inputs, inputEvent)
				default:
					draining = false
				}
			}
		}
		var event sdl.Event
		quit := false
		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				quit = true
			case sdl.EVENT_MOUSE_BUTTON_DOWN:
				button := event.MouseButtonEvent()
				if !fullscreen {
					x, y := canvasCoordinates(window, button.X, button.Y)
					inputs = append(inputs, input.Event{Kind: input.AbsolutePointer, Position: input.Point{X: x, Y: y}})
					inputs = append(inputs, input.Event{Kind: input.ButtonDown, Button: input.Button(button.Button)})
				}
			case sdl.EVENT_MOUSE_BUTTON_UP:
				button := event.MouseButtonEvent()
				if !fullscreen {
					x, y := canvasCoordinates(window, button.X, button.Y)
					inputs = append(inputs, input.Event{Kind: input.AbsolutePointer, Position: input.Point{X: x, Y: y}})
					inputs = append(inputs, input.Event{Kind: input.ButtonUp, Button: input.Button(button.Button)})
				}
			case sdl.EVENT_MOUSE_MOTION:
				if !fullscreen {
					motion := event.MouseMotionEvent()
					x, y := canvasCoordinates(window, motion.X, motion.Y)
					inputs = append(inputs, input.Event{Kind: input.AbsolutePointer, Position: input.Point{X: x, Y: y}})
				}
			}
		}
		for _, inputEvent := range inputs {
			for _, gesture := range recognizer.Handle(inputEvent) {
				handleGesture(gesture)
			}
		}
		pointer := recognizer.Position()
		pointer.X = max(0, min(pointer.X, float32(canvasW-1)))
		pointer.Y = max(0, min(pointer.Y, float32(canvasH-1)))
		recognizer.SetPosition(pointer)
		if quit || (testDuration > 0 && time.Since(started) >= testDuration) {
			break
		}
		w, h, err := renderer.RenderOutputSize()
		if err != nil {
			log.Fatal(err)
		}
		draw(renderer, text, cursorSize, volume, playbackProgress, float32(w), float32(h))
		if fullscreen {
			scaleX, scaleY := float32(w)/float32(canvasW), float32(h)/float32(canvasH)
			dst := &sdl.FRect{
				X: pointer.X - 1/scaleX, Y: pointer.Y - 1/scaleY,
				W: float32(cursorSize) / scaleX, H: float32(cursorSize) / scaleY,
			}
			if err := renderer.RenderTexture(cursorTexture, nil, dst); err != nil {
				log.Fatal(err)
			}
		}
		if err := renderer.Present(); err != nil {
			log.Fatal(err)
		}
		sdl.Delay(16)
	}
	log.Print("SDL session ended cleanly.")
}

func draw(r *sdl.Renderer, text *textPainter, cursorSize, volume int, playbackProgress, w, h float32) {
	// The window is 1280x800 points. On the target Retina-sized display, its
	// high-density render surface is 2560x1600 pixels, matching the target Mac.
	scaleX, scaleY := w/float32(canvasW), h/float32(canvasH)
	if scaleX <= 0 || scaleY <= 0 {
		return
	}
	if err := r.SetScale(scaleX, scaleY); err != nil {
		log.Fatal(err)
	}

	set(r, background)
	_ = r.Clear()
	fill := func(c color, x, y, rw, rh float32) {
		rect(r, c, x, y, rw, rh)
	}
	label := func(x, y float32, value string, c color) {
		text.Draw(x, y, value, c)
	}
	card := func(x, y, cw, ch float32) {
		fill(panel, x, y, cw, ch)
		fill(outline, x, y, cw, 1)
		fill(outline, x, y+ch-1, cw, 1)
		fill(outline, x, y, 1, ch)
		fill(outline, x+cw-1, y, 1, ch)
	}

	// Slim system strip
	fill(panelAlt, 0, 0, canvasW, 42)
	fill(outline, 0, 41, canvasW, 1)
	label(850, 14, "CURSOR", muted)
	fill(outline, cursorSliderX, 20, cursorSliderW, 4)
	progress := float32(cursorSize-minCursorSize) / float32(maxCursorSize-minCursorSize)
	fill(blue, cursorSliderX, 20, cursorSliderW*progress, 4)
	knobX := float32(cursorSliderX) + cursorSliderW*progress - 6
	fill(white, knobX, 14, 12, 16)
	label(cursorSliderX-3, 30, "32", muted)
	label(cursorSliderX+cursorSliderW/2-6, 30, "60", muted)
	label(cursorSliderX+cursorSliderW-9, 30, "88", muted)
	label(1074, 14, fmt.Sprintf("%d px", cursorSize), white)
	label(1190, 17, "WIFI", white)
	label(1245, 17, "261026  22:57", white)

	// Product identity and subtitle
	// Stylized waveform mark made from simple bars, avoiding a machine-specific asset.
	fill(white, 95, 111, 4, 12)
	fill(white, 111, 96, 4, 42)
	fill(white, 127, 83, 4, 68)
	fill(white, 143, 101, 4, 34)
	fill(white, 159, 112, 4, 12)
	label(193, 98, "audio-utility", white)
	label(193, 135, "play  ·  record  ·  monitor  ·  explore", muted)

	// Output device and level
	card(84, 194, 464, 190)
	label(108, 216, "Output", white)
	card(105, 253, 422, 42)
	label(122, 269, "Internal Speakers", white)
	label(501, 269, "v", muted)
	label(108, 327, "VOL", muted)
	fill(outline, volumeSliderX, volumeSliderY, volumeSliderW, 6)
	volumeProgress := float32(volume) / 100
	fill(blue, volumeSliderX, volumeSliderY, volumeSliderW*volumeProgress, 6)
	fill(white, volumeSliderX+volumeSliderW*volumeProgress-11, volumeSliderY-8, 22, 22)
	label(502, 333, fmt.Sprintf("%d", volume), muted)

	// Input device and live level meter
	card(84, 402, 464, 184)
	label(108, 424, "Input", white)
	card(105, 461, 422, 42)
	label(122, 477, "Internal Microphone", white)
	label(501, 477, "v", muted)
	for i := 0; i < 38; i++ {
		x := float32(108 + i*11)
		c := color{57, 65, 77, 255}
		if i < 17 {
			c = green
		} else if i > 30 && i < 34 {
			c = orange
		}
		fill(c, x, 535, 6, 24)
	}

	// Audio tools
	card(84, 604, 464, 142)
	label(108, 625, "Audio Tools", white)
	tool := func(x float32, icon, title string) {
		card(x, 661, 130, 68)
		label(x+48, 671, icon, white)
		label(x+37, 703, title, white)
	}
	tool(105, "~", "Tone")
	tool(260, "O", "Record")
	tool(415, "|||", "Analyzer")

	// Spectrum analyzer panel
	card(568, 202, 714, 380)
	label(594, 224, "Spectrum", white)
	label(1156, 224, "LIVE  |  48 kHz", muted)
	plotX, plotY, plotW, plotH := float32(627), float32(264), float32(628), float32(258)
	for i := 0; i <= 6; i++ {
		y := plotY + float32(i)*plotH/6
		fill(grid, plotX, y, plotW, 1)
		label(590, y+4, []string{"0", "-12", "-24", "-36", "-48", "-60", ""}[i], muted)
	}
	for i := 0; i <= 10; i++ {
		x := plotX + float32(i)*plotW/10
		fill(grid, x, plotY, 1, plotH)
	}
	levels := []int{64, 90, 108, 92, 136, 158, 194, 171, 151, 128, 112, 96, 88, 78, 70, 84, 69, 59, 58, 60, 72, 62, 54, 48, 46, 54, 44, 40, 47, 38, 34, 42, 32, 28, 25, 24, 22, 20, 18, 17, 15, 13, 12, 10, 8, 7}
	barGap := float32(4)
	barW := (plotW - float32(len(levels))*barGap) / float32(len(levels))
	for i, v := range levels {
		x := plotX + float32(i)*(barW+barGap)
		bh := float32(v)
		for j := 0; j < 6; j++ {
			c := color{uint8(48 + j*4), uint8(111 + j*8), uint8(181 + j*9), 255}
			segment := bh / 6
			fill(c, x, plotY+plotH-bh+float32(j)*segment, barW, segment+1)
		}
	}
	for i, s := range []string{"20", "50", "100", "200", "500", "1k", "2k", "5k", "10k", "20k"} {
		x := plotX + float32(i)*plotW/9 - 6
		label(x, plotY+plotH+18, s, muted)
	}

	// Player controls
	card(568, 602, 714, 144)
	label(594, 624, "Now Playing", white)
	label(1134, 624, "00:00 / 00:00", muted)
	fill(outline, playbackSliderX, playbackSliderY, playbackSliderW, 8)
	fill(blue, playbackSliderX, playbackSliderY, playbackSliderW*playbackProgress, 8)
	fill(white, playbackSliderX+playbackSliderW*playbackProgress-10, playbackSliderY-7, 20, 22)
	label(864, 710, "<<     PLAY     >>", white)
}

func canvasCoordinates(window *sdl.Window, x, y float32) (float32, float32) {
	w, h, err := window.Size()
	if err != nil || w <= 0 || h <= 0 {
		log.Fatal("could not read window size for cursor control")
	}
	return x * float32(canvasW) / float32(w), y * float32(canvasH) / float32(h)
}

func inCursorSlider(x, y float32) bool {
	return x >= cursorSliderX-18 && x <= cursorSliderX+cursorSliderW+30 && y >= 5 && y <= 38
}

func inVolumeSlider(x, y float32) bool {
	return x >= volumeSliderX-18 && x <= volumeSliderX+volumeSliderW+18 &&
		y >= volumeSliderY-21 && y <= volumeSliderY+25
}

func inPlaybackSlider(x, y float32) bool {
	return x >= playbackSliderX-20 && x <= playbackSliderX+playbackSliderW+20 &&
		y >= playbackSliderY-23 && y <= playbackSliderY+25
}

func volumeAt(x float32) int {
	position := (x - volumeSliderX) / volumeSliderW
	position = max(0, min(position, 1))
	return int(math.Round(float64(position * 100)))
}

func progressAt(x, start, width float32) float32 {
	position := (x - start) / width
	return max(0, min(position, 1))
}

func cursorSizeAt(x float32) int {
	position := (x - cursorSliderX) / cursorSliderW
	position = max(0, min(position, 1))
	steps := int(math.Round(float64(position * float32(maxCursorSize-minCursorSize) / cursorSizeStep)))
	return minCursorSize + steps*cursorSizeStep
}

func createMouseCursor(size int) (*sdl.Cursor, error) {
	surface, err := createCursorSurface(size)
	if err != nil {
		return nil, err
	}
	defer surface.Destroy()
	return surface.CreateColorCursor(1, 1)
}

func createCursorTexture(renderer *sdl.Renderer, size int) (*sdl.Texture, error) {
	surface, err := createCursorSurface(size)
	if err != nil {
		return nil, err
	}
	defer surface.Destroy()
	return renderer.CreateTextureFromSurface(surface)
}

func createCursorSurface(size int) (*sdl.Surface, error) {
	pixels := image.NewRGBA(image.Rect(0, 0, size, size))
	points := []image.Point{
		{X: 1, Y: 1},
		{X: 1, Y: size - 1},
		{X: int(float64(size) * 0.29), Y: int(float64(size) * 0.67)},
		{X: int(float64(size) * 0.43), Y: int(float64(size) * 0.97)},
		{X: int(float64(size) * 0.60), Y: int(float64(size) * 0.90)},
		{X: int(float64(size) * 0.46), Y: int(float64(size) * 0.63)},
		{X: int(float64(size) * 0.82), Y: int(float64(size) * 0.61)},
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			inside, edgeDistance := cursorShapeAt(float64(x)+0.5, float64(y)+0.5, points)
			if !inside {
				continue
			}
			if edgeDistance <= 1.35 {
				pixels.SetRGBA(x, y, stdcolor.RGBA{R: 17, G: 22, B: 30, A: 255})
			} else {
				pixels.SetRGBA(x, y, stdcolor.RGBA{R: 245, G: 248, B: 252, A: 255})
			}
		}
	}
	return sdl.CreateSurfaceFrom(size, size, sdl.PIXELFORMAT_RGBA32, pixels.Pix, pixels.Stride)
}

func cursorShapeAt(x, y float64, points []image.Point) (bool, float64) {
	inside := false
	minDistance := math.MaxFloat64
	for i, point := range points {
		next := points[(i+1)%len(points)]
		x1, y1 := float64(point.X), float64(point.Y)
		x2, y2 := float64(next.X), float64(next.Y)
		if (y1 > y) != (y2 > y) && x < (x2-x1)*(y-y1)/(y2-y1)+x1 {
			inside = !inside
		}
		minDistance = min(minDistance, distanceToSegment(x, y, x1, y1, x2, y2))
	}
	return inside, minDistance
}

func distanceToSegment(x, y, x1, y1, x2, y2 float64) float64 {
	dx, dy := x2-x1, y2-y1
	if dx == 0 && dy == 0 {
		return math.Hypot(x-x1, y-y1)
	}
	t := max(0, min(1, ((x-x1)*dx+(y-y1)*dy)/(dx*dx+dy*dy)))
	return math.Hypot(x-(x1+t*dx), y-(y1+t*dy))
}

func set(r *sdl.Renderer, c color) {
	if err := r.SetDrawColor(c.r, c.g, c.b, c.a); err != nil {
		log.Fatal(err)
	}
}

type textKey struct {
	value string
	color color
	size  int
	bold  bool
}

type textPainter struct {
	renderer *sdl.Renderer
	font     *opentype.Font
	cache    map[textKey]*sdl.Texture
}

func (p *textPainter) Draw(x, y float32, value string, c color) {
	if value == "" {
		return
	}
	size, bold := textStyle(value)
	key := textKey{value: value, color: c, size: size, bold: bold}
	texture := p.cache[key]
	if texture == nil {
		face, err := opentype.NewFace(p.font, &opentype.FaceOptions{Size: float64(size), DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			log.Fatal(err)
		}
		bounds := font.MeasureString(face, value)
		width := max(1, bounds.Ceil()+2)
		height := max(1, face.Metrics().Height.Ceil()+2)
		pixels := image.NewRGBA(image.Rect(0, 0, width, height))
		drawer := font.Drawer{
			Dst: pixels, Src: image.NewUniform(stdcolor.RGBA{R: c.r, G: c.g, B: c.b, A: c.a}),
			Face: face, Dot: fixed.P(1, face.Metrics().Ascent.Ceil()+1),
		}
		drawer.DrawString(value)
		if bold {
			drawer.Dot.X++
			drawer.DrawString(value)
		}
		texture, err = p.renderer.CreateTexture(sdl.PIXELFORMAT_RGBA32, sdl.TEXTUREACCESS_STATIC, width, height)
		if err != nil {
			log.Fatal(err)
		}
		if err := texture.Update(nil, pixels.Pix, int32(pixels.Stride)); err != nil {
			log.Fatal(err)
		}
		if err := texture.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
			log.Fatal(err)
		}
		p.cache[key] = texture
	}
	w, h, err := texture.Size()
	if err != nil {
		log.Fatal(err)
	}
	if err := p.renderer.RenderTexture(texture, nil, &sdl.FRect{X: x, Y: y, W: w, H: h}); err != nil {
		log.Fatal(err)
	}
}

func (p *textPainter) Close() {
	for _, texture := range p.cache {
		texture.Destroy()
	}
}

func textStyle(value string) (size int, bold bool) {
	switch value {
	case "audio-utility":
		return 36, false
	case "~", "O", "|||":
		return 27, false
	case "261026  22:57", "LIVE  |  48 kHz":
		return 14, false
	case "32", "60", "88":
		return 10, false
	case "Output", "Input", "Audio Tools", "Spectrum", "Now Playing":
		return 18, true
	default:
		if value == strings.ToUpper(value) && len(value) > 2 {
			return 14, true
		}
		return 16, false
	}
}

func rect(r *sdl.Renderer, c color, x, y, w, h float32) {
	set(r, c)
	if err := r.RenderFillRect(&sdl.FRect{X: x, Y: y, W: w, H: h}); err != nil {
		log.Fatal(err)
	}
}
