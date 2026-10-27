//go:build linux

// Package renderer owns the Ultralight/GLES implementation and display resources.
package renderer

import (
	"fmt"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/ebitengine/purego"
	"image"
	"image/png"
	"log"
	"os"
	"time"
	"unsafe"
)

const canvasW, canvasH = 1280, 800

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type Renderer struct {
	Window                                                  *sdl.Window
	u                                                       *ulAPI
	j                                                       *jsAPI
	view, renderer                                          uintptr
	g                                                       *glAPI
	driver                                                  *gpuDriver
	width, height                                           int32
	scene, ui, vao                                          uint32
	elapsedLoc, sceneLoc, pointerLoc, uiPointer, uvScaleLoc int32
	cleanup                                                 []func()
}

func New(width, height uint32) (result *Renderer, err error) {
	owner := &Renderer{}
	defer func() {
		if result == nil {
			owner.Close()
		}
	}()
	h, err := purego.Dlopen("libAppCore.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, err
	}
	u := &ulAPI{}
	bindAPI(h, "ul", u)
	j := &jsAPI{}
	bindAPI(h, "JS", j)
	u.EnablePlatformFontLoader()
	u.with(".", u.EnablePlatformFileSystem)
	u.with("ultralight.log", u.EnableDefaultLogger)

	var window *sdl.Window
	var g *glAPI
	var w, hh int32
	must(sdl.LoadLibrary(sdl.Path()))
	owner.cleanup = append(owner.cleanup, func() { sdl.CloseLibrary() })
	must(sdl.Init(sdl.INIT_VIDEO))
	owner.cleanup = append(owner.cleanup, func() { sdl.Quit() })
	must(sdl.GL_SetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, int32(sdl.GL_CONTEXT_PROFILE_ES)))
	must(sdl.GL_SetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3))
	must(sdl.GL_SetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 0))
	window, err = sdl.CreateWindow("Ultralight stack evaluation", 0, 0, sdl.WINDOW_FULLSCREEN|sdl.WINDOW_OPENGL)
	if err != nil {
		return nil, err
	}
	owner.cleanup = append(owner.cleanup, func() { window.Destroy() })
	context, err := sdl.GL_CreateContext(window)
	if err != nil {
		return nil, err
	}
	owner.cleanup = append(owner.cleanup, func() { sdl.GL_DestroyContext(context) })
	must(sdl.GL_MakeCurrent(window, context))
	// The bcm5974 adapter owns the pointer. SDL's evdev cursor otherwise moves
	// independently, producing two visible cursors with different hit positions.
	must(sdl.HideCursor())
	owner.cleanup = append(owner.cleanup, func() { sdl.ShowCursor() })
	must(sdl.GL_SetSwapInterval(1))
	must(window.StartTextInput())
	owner.cleanup = append(owner.cleanup, func() { window.StopTextInput() })
	w, hh, err = window.SizeInPixels()
	if err != nil {
		return nil, err
	}
	g = loadGL()
	log.Printf("GLES %s; GPU %s; display %dx%d; HTML %dx%d", g.GetString(0x1F02), g.GetString(0x1F01), w, hh, width, height)
	g.Viewport(0, 0, int32(w), int32(hh))
	driver := newGPUDriver(g, u, h)
	owner.cleanup = append(owner.cleanup, func() { driver.close() })
	c := u.CreateConfig()
	owner.cleanup = append(owner.cleanup, func() { u.DestroyConfig(c) })
	r := u.CreateRenderer(c)
	if r == 0 {
		return nil, fmt.Errorf("renderer creation failed")
	}
	owner.cleanup = append(owner.cleanup, func() { u.DestroyRenderer(r) })
	vc := u.CreateViewConfig()
	u.ViewConfigSetIsAccelerated(vc, true)
	u.ViewConfigSetIsTransparent(vc, true)
	v := u.CreateView(r, width, height, vc, 0)
	u.DestroyViewConfig(vc)
	if v == 0 {
		return nil, fmt.Errorf("view creation failed")
	}
	owner.cleanup = append(owner.cleanup, func() { u.DestroyView(v) })
	log.Printf("HTML renderer=gpu accelerated=%t", u.ViewIsAccelerated(v))
	if !u.ViewIsAccelerated(v) {
		return nil, fmt.Errorf("GPU acceleration is required")
	}
	u.with("file:///ui/index.html", func(p uintptr) { u.ViewLoadURL(v, p) })
	deadline := time.Now().Add(15 * time.Second)
	for {
		u.Update(r)
		ready, e := u.eval(v, "document.readyState")
		if e == nil && ready == "complete" {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("page load timed out")
		}
		time.Sleep(10 * time.Millisecond)
	}
	scene := g.program(sceneShader)
	owner.cleanup = append(owner.cleanup, func() { g.DeleteProgram(scene) })
	ui := g.program(uiShader)
	owner.cleanup = append(owner.cleanup, func() { g.DeleteProgram(ui) })
	var mainVAO uint32
	g.GenVertexArrays(1, &mainVAO)
	owner.cleanup = append(owner.cleanup, func() { g.DeleteVertexArrays(1, &mainVAO) })
	g.BindVertexArray(mainVAO)
	var buffer uint32
	vertices := []float32{-1, -1, 1, -1, -1, 1, 1, 1}
	g.GenBuffers(1, &buffer)
	owner.cleanup = append(owner.cleanup, func() { g.DeleteBuffers(1, &buffer) })
	g.BindBuffer(0x8892, buffer)
	g.BufferData(0x8892, uintptr(len(vertices)*4), unsafe.Pointer(&vertices[0]), 0x88E4)
	g.EnableVertexAttribArray(0)
	g.VertexAttribPointer(0, 2, 0x1406, false, 0, 0)
	elapsedLoc := g.GetUniformLocation(scene, "elapsed")
	sceneLoc := g.GetUniformLocation(scene, "scene")
	pointerLoc := g.GetUniformLocation(scene, "pointer")
	uiPointer := g.GetUniformLocation(ui, "pointer")
	g.UseProgram(ui)
	g.Uniform1i(g.GetUniformLocation(ui, "tex"), 0)
	uvScaleLoc := g.GetUniformLocation(ui, "uvScale")
	g.Uniform2f(uvScaleLoc, 1, 1)
	u.ViewFocus(v)

	owner.Window = window
	owner.u = u
	owner.j = j
	owner.view = v
	owner.renderer = r
	owner.g = g
	owner.driver = driver
	owner.width = w
	owner.height = hh
	owner.scene = scene
	owner.ui = ui
	owner.vao = mainVAO
	owner.elapsedLoc = elapsedLoc
	owner.sceneLoc = sceneLoc
	owner.pointerLoc = pointerLoc
	owner.uiPointer = uiPointer
	owner.uvScaleLoc = uvScaleLoc
	return owner, nil
}
func (r *Renderer) Close() {
	for i := len(r.cleanup) - 1; i >= 0; i-- {
		r.cleanup[i]()
	}
	r.cleanup = nil
}
func (r *Renderer) Eval(script string) string {
	value, err := r.u.eval(r.view, script)
	must(err)
	return value
}
func (r *Renderer) Mouse(kind uint32, x, y int32, button uint32) {
	r.u.mouse(r.view, kind, x, y, button)
}
func (r *Renderer) Key(kind, mods uint32, vk, native int32, text string, repeat bool) {
	r.u.key(r.view, kind, mods, vk, native, text, repeat)
}
func (r *Renderer) Scroll(x, y int32) {
	e := r.u.CreateScrollEvent(0, x, y)
	defer r.u.DestroyScrollEvent(e)
	r.u.ViewFireScrollEvent(r.view, e)
}
func (r *Renderer) Update() { r.u.Update(r.renderer) }
func (r *Renderer) Draw(age time.Duration, static bool, calls int, x, y float32) (time.Duration, time.Duration, error) {
	u, g, driver := r.u, r.g, r.driver
	u.Update(r.renderer)
	u.RefreshDisplay(r.renderer, 0)
	t := time.Now()
	u.Render(r.renderer)
	renderTime := time.Since(t)
	t = time.Now()
	driver.draw()
	submitTime := time.Since(t)
	target := u.ViewGetRenderTarget(r.view)
	tex, ok := driver.textures[target.TextureID]
	if !ok {
		return 0, 0, fmt.Errorf("GPU view has no output texture")
	}
	g.ActiveTexture(0x84C0)
	g.BindTexture(0x0DE1, tex.name)
	g.UseProgram(r.ui)
	g.Uniform2f(r.uvScaleLoc, target.Right, target.Bottom)
	g.BindFramebuffer(0x8D40, 0)
	g.Viewport(0, 0, r.width, r.height)
	g.BindVertexArray(r.vao)
	g.Disable(0x0BE2)
	g.UseProgram(r.scene)
	seconds := float32(age.Seconds())
	if static {
		seconds = 0
	}
	g.Uniform1f(r.elapsedLoc, seconds)
	g.Uniform1f(r.sceneLoc, float32(calls))
	g.Uniform2f(r.pointerLoc, x/canvasW, y/canvasH)
	g.DrawArrays(0x0005, 0, 4)
	g.Enable(0x0BE2)
	g.BlendFunc(1, 0x0303)
	g.UseProgram(r.ui)
	g.Uniform2f(r.uiPointer, x/canvasW, y/canvasH)
	g.DrawArrays(0x0005, 0, 4)
	g.check()
	return renderTime, submitTime, nil
}
func (r *Renderer) Swap() error { return sdl.GL_SwapWindow(r.Window) }
func (r *Renderer) LogStats(frames int, submission time.Duration) {
	d := r.driver
	log.Printf("GPU commands_drawn=%d command_updates=%d asset_upload_bytes=%d draw_submit_ms_per_frame=%.3f", d.draws, d.updates, d.textureBytes, submission.Seconds()*1000/float64(max(1, frames)))
}
func (r *Renderer) Capture() { capture(r.g, int(r.width), int(r.height)) }
func capture(g *glAPI, w, h int) {
	pixels := make([]byte, w*h*4)
	g.ReadPixels(0, 0, int32(w), int32(h), 0x1908, 0x1401, unsafe.Pointer(&pixels[0]))
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		copy(out.Pix[y*out.Stride:(y+1)*out.Stride], pixels[(h-1-y)*w*4:(h-y)*w*4])
	}
	f, e := os.Create("screen.png")
	must(e)
	defer f.Close()
	must(png.Encode(f, out))
}
