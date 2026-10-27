//go:build linux

package renderer

import (
	"fmt"
	"github.com/ebitengine/purego"
	"log"
	"os"
	"time"
	"unsafe"
)

// Native layouts are checked against the pinned SDK by gpu_bridge.c at startup.
type renderTarget struct {
	Empty                                                         bool
	Width, Height, TextureID, TextureWidth, TextureHeight, Format uint32
	Left, Top, Right, Bottom                                      float32
	RenderBufferID                                                uint32
}
type matrix struct{ Data [16]float32 }
type gpuState struct {
	Width, Height                                uint32
	Transform                                    matrix
	Texturing, Blend                             bool
	Shader                                       uint8
	RenderBufferID, Texture1, Texture2, Texture3 uint32
	Scalar                                       [8]float32
	Vector                                       [32]float32
	ClipSize                                     uint8
	Clip                                         [128]float32
	Scissor                                      bool
	ScissorRect                                  rect
}
type gpuCommand struct {
	Type                      uint8
	State                     gpuState
	GeometryID, Count, Offset uint32
}
type renderBuffer struct {
	TextureID, Width, Height uint32
	Stencil, Depth           bool
}
type vertexBuffer struct {
	Format, Size uint32
	Data         unsafe.Pointer
}
type indexBuffer struct {
	Size uint32
	Data unsafe.Pointer
}
type commandList struct {
	Size     uint32
	Commands *gpuCommand
}
type gpuTexture struct {
	name          uint32
	width, height uint32
	format        uint32
	empty         bool
}
type gpuGeometry struct{ vao, vertices, indices, format uint32 }
type gpuProgram struct {
	id                                               uint32
	state, transform, scalar, vector, clipSize, clip int32
}
type gpuDriver struct {
	g                            *glAPI
	u                            *ulAPI
	textures                     map[uint32]gpuTexture
	buffers                      map[uint32]uint32
	geometry                     map[uint32]gpuGeometry
	programs                     [2]gpuProgram
	commands                     []gpuCommand
	ids                          [3]uint32
	projection                   func(uintptr, *matrix, float32, float32, bool, *matrix)
	projectionFn                 uintptr
	started                      time.Time
	draws, updates, textureBytes uint64
}

func newGPUDriver(g *glAPI, u *ulAPI, ulHandle uintptr) *gpuDriver {
	d := &gpuDriver{g: g, u: u, textures: map[uint32]gpuTexture{}, buffers: map[uint32]uint32{}, geometry: map[uint32]gpuGeometry{}, started: time.Now()}
	bridge, e := purego.Dlopen("./libgpu_bridge.so", purego.RTLD_NOW|purego.RTLD_LOCAL)
	must(e)
	var layout func(uint32) uint32
	purego.RegisterLibFunc(&layout, bridge, "bridge_layout")
	expected := []uintptr{unsafe.Sizeof(gpuState{}), unsafe.Sizeof(gpuCommand{}), unsafe.Sizeof(renderTarget{}), unsafe.Offsetof(gpuState{}.RenderBufferID), unsafe.Offsetof(gpuState{}.Scalar), unsafe.Offsetof(gpuState{}.Clip), unsafe.Offsetof(gpuState{}.ScissorRect), unsafe.Offsetof(gpuCommand{}.State), unsafe.Offsetof(gpuCommand{}.GeometryID), unsafe.Sizeof(renderBuffer{}), unsafe.Sizeof(vertexBuffer{}), unsafe.Sizeof(indexBuffer{})}
	for i, want := range expected {
		if got := layout(uint32(i)); uintptr(got) != want {
			panic(fmt.Sprintf("GPU ABI layout %d: C=%d Go=%d", i, got, want))
		}
	}
	log.Print("GPU bridge: all 12 SDK ABI layout checks passed")
	purego.RegisterLibFunc(&d.projection, bridge, "bridge_projection")
	d.projectionFn, e = purego.Dlsym(ulHandle, "ulApplyProjection")
	must(e)
	setter, e := purego.Dlsym(ulHandle, "ulPlatformSetGPUDriver")
	must(e)
	var install func(uintptr, *[14]uintptr, uintptr, uintptr, uintptr, uintptr)
	purego.RegisterLibFunc(&install, bridge, "bridge_install")
	next := func(kind int) func() uint32 { return func() uint32 { d.ids[kind]++; return d.ids[kind] } }
	callbacks := [14]uintptr{
		purego.NewCallback(func() {}), purego.NewCallback(func() {}), purego.NewCallback(next(0)),
		purego.NewCallback(d.createTexture), purego.NewCallback(d.updateTexture), purego.NewCallback(d.destroyTexture),
		purego.NewCallback(next(1)), 0, purego.NewCallback(d.destroyBuffer), purego.NewCallback(next(2)), 0, 0, purego.NewCallback(d.destroyGeometry), 0,
	}
	install(setter, &callbacks, purego.NewCallback(d.createBuffer), purego.NewCallback(d.createGeometry), purego.NewCallback(d.updateGeometry), purego.NewCallback(func(list *commandList) {
		d.commands = append(d.commands, unsafe.Slice(list.Commands, int(list.Size))...)
	}))
	for i, pair := range [][2]string{{"v2f_c4f_t2f_t2f_d28f_vert", "fill_frag"}, {"v2f_c4f_t2f_vert", "fill_path_frag"}} {
		vs, e := os.ReadFile("gpu-shaders/" + pair[0] + ".glsl")
		must(e)
		fs, e := os.ReadFile("gpu-shaders/" + pair[1] + ".glsl")
		must(e)
		attrs := []string{"in_Position", "in_Color", "in_TexCoord"}
		if i == 0 {
			attrs = append(attrs, "in_ObjCoord", "in_Data0", "in_Data1", "in_Data2", "in_Data3", "in_Data4", "in_Data5", "in_Data6")
		}
		id := g.programSources(string(vs), string(fs), attrs)
		g.UseProgram(id)
		for slot := 0; slot < 3; slot++ {
			g.Uniform1i(g.GetUniformLocation(id, fmt.Sprintf("Texture%d", slot+1)), int32(slot))
		}
		d.programs[i] = gpuProgram{id, g.GetUniformLocation(id, "State"), g.GetUniformLocation(id, "Transform"), g.GetUniformLocation(id, "Scalar4"), g.GetUniformLocation(id, "Vector"), g.GetUniformLocation(id, "ClipSize"), g.GetUniformLocation(id, "Clip")}
	}
	g.check()
	return d
}
func (d *gpuDriver) createTexture(id uint32, b uintptr) {
	g, u := d.g, d.u
	t := gpuTexture{width: u.BitmapGetWidth(b), height: u.BitmapGetHeight(b), format: u.BitmapGetFormat(b), empty: u.BitmapIsEmpty(b)}
	g.GenTextures(1, &t.name)
	g.ActiveTexture(0x84C0)
	g.BindTexture(0x0DE1, t.name)
	g.TexParameteri(0x0DE1, 0x2801, 0x2601)
	g.TexParameteri(0x0DE1, 0x2800, 0x2601)
	g.TexParameteri(0x0DE1, 0x2802, 0x812F)
	g.TexParameteri(0x0DE1, 0x2803, 0x812F)
	internal, format := int32(0x8058), uint32(0x1908)
	var pixels unsafe.Pointer
	if !t.empty {
		pixels = u.BitmapLockPixels(b)
		defer u.BitmapUnlockPixels(b)
		if t.format == 0 {
			internal, format = 0x8229, 0x1903
		} else if t.format == 1 {
			g.TexParameteri(0x0DE1, 0x8E42, 0x1905)
			g.TexParameteri(0x0DE1, 0x8E44, 0x1903)
		} else {
			panic("unsupported SDK bitmap format")
		}
		g.PixelStorei(0x0CF2, int32(u.BitmapGetRowBytes(b)/u.BitmapGetBpp(b)))
		d.textureBytes += uint64(u.BitmapGetRowBytes(b)) * uint64(t.height)
	}
	g.PixelStorei(0x0CF5, 1)
	g.TexImage2D(0x0DE1, 0, internal, int32(t.width), int32(t.height), 0, format, 0x1401, pixels)
	g.PixelStorei(0x0CF2, 0)
	g.PixelStorei(0x0CF5, 4)
	d.textures[id] = t
	g.check()
}
func (d *gpuDriver) updateTexture(id uint32, b uintptr) {
	g, u := d.g, d.u
	t, ok := d.textures[id]
	if !ok {
		panic("update of missing GPU texture")
	}
	g.ActiveTexture(0x84C0)
	g.BindTexture(0x0DE1, t.name)
	g.PixelStorei(0x0CF5, 1)
	g.PixelStorei(0x0CF2, int32(u.BitmapGetRowBytes(b)/u.BitmapGetBpp(b)))
	format := uint32(0x1908)
	if t.format == 0 {
		format = 0x1903
	}
	pixels := u.BitmapLockPixels(b)
	g.TexSubImage2D(0x0DE1, 0, 0, 0, int32(t.width), int32(t.height), format, 0x1401, pixels)
	u.BitmapUnlockPixels(b)
	d.textureBytes += uint64(u.BitmapGetRowBytes(b)) * uint64(t.height)
	g.PixelStorei(0x0CF2, 0)
	g.PixelStorei(0x0CF5, 4)
	g.check()
}
func (d *gpuDriver) destroyTexture(id uint32) {
	if t, ok := d.textures[id]; ok {
		d.g.DeleteTextures(1, &t.name)
		delete(d.textures, id)
	}
}
func (d *gpuDriver) createBuffer(id uint32, b *renderBuffer) {
	t, ok := d.textures[b.TextureID]
	if !ok {
		panic("render buffer has no texture")
	}
	var fbo uint32
	g := d.g
	g.GenFramebuffers(1, &fbo)
	g.BindFramebuffer(0x8D40, fbo)
	g.FramebufferTexture2D(0x8D40, 0x8CE0, 0x0DE1, t.name, 0)
	if g.CheckFramebufferStatus(0x8D40) != 0x8CD5 {
		panic("GPU framebuffer incomplete")
	}
	d.buffers[id] = fbo
	g.check()
}
func (d *gpuDriver) destroyBuffer(id uint32) {
	if fbo, ok := d.buffers[id]; ok {
		d.g.DeleteFramebuffers(1, &fbo)
		delete(d.buffers, id)
	}
}
func (d *gpuDriver) createGeometry(id uint32, v *vertexBuffer, i *indexBuffer) {
	g := d.g
	geom := gpuGeometry{format: v.Format}
	g.GenVertexArrays(1, &geom.vao)
	g.BindVertexArray(geom.vao)
	g.GenBuffers(1, &geom.vertices)
	g.GenBuffers(1, &geom.indices)
	d.geometry[id] = geom
	d.updateGeometry(id, v, i)
	stride := int32(20)
	count := 3
	if v.Format == 1 {
		stride = 140
		count = 11
	} else if v.Format != 0 {
		panic("unsupported vertex format")
	}
	for attr := 0; attr < count; attr++ {
		components, kind, normalized, offset := int32(2), uint32(0x1406), false, uintptr(0)
		switch attr {
		case 0:
			offset = 0
		case 1:
			components, kind, normalized, offset = 4, 0x1401, true, 8
		case 2:
			offset = 12
		case 3:
			offset = 20
		default:
			components, offset = 4, uintptr(28+(attr-4)*16)
		}
		g.EnableVertexAttribArray(uint32(attr))
		g.VertexAttribPointer(uint32(attr), components, kind, normalized, stride, offset)
	}
	g.BindVertexArray(0)
	g.check()
}
func (d *gpuDriver) updateGeometry(id uint32, v *vertexBuffer, i *indexBuffer) {
	geom, ok := d.geometry[id]
	if !ok {
		panic("update of missing geometry")
	}
	g := d.g
	g.BindVertexArray(geom.vao)
	g.BindBuffer(0x8892, geom.vertices)
	g.BufferData(0x8892, uintptr(v.Size), v.Data, 0x88E8)
	g.BindBuffer(0x8893, geom.indices)
	g.BufferData(0x8893, uintptr(i.Size), i.Data, 0x88E8)
	g.check()
}
func (d *gpuDriver) destroyGeometry(id uint32) {
	if geom, ok := d.geometry[id]; ok {
		g := d.g
		g.DeleteVertexArrays(1, &geom.vao)
		g.DeleteBuffers(1, &geom.vertices)
		g.DeleteBuffers(1, &geom.indices)
		delete(d.geometry, id)
	}
}
func (d *gpuDriver) draw() {
	g := d.g
	g.Disable(0x0B71)
	g.Disable(0x0B44)
	g.BlendFunc(1, 0x0303)
	if len(d.commands) > 0 {
		d.updates++
	}
	for _, c := range d.commands {
		s := &c.State
		fbo, ok := d.buffers[s.RenderBufferID]
		if !ok {
			panic(fmt.Sprintf("missing render buffer %d", s.RenderBufferID))
		}
		g.BindFramebuffer(0x8D40, fbo)
		if c.Type == 0 {
			g.Disable(0x0C11)
			g.ClearColor(0, 0, 0, 0)
			g.Clear(0x4000)
			continue
		}
		if c.Type != 1 {
			panic("unsupported GPU command")
		}
		if s.Shader > 1 {
			panic("unsupported GPU shader")
		}
		p := d.programs[s.Shader]
		g.UseProgram(p.id)
		g.Viewport(0, 0, int32(s.Width), int32(s.Height))
		var transform matrix
		d.projection(d.projectionFn, &s.Transform, float32(s.Width), float32(s.Height), true, &transform)
		state := [4]float32{float32(time.Since(d.started).Seconds()), float32(s.Width), float32(s.Height), 1}
		g.Uniform4fv(p.state, 1, &state[0])
		g.UniformMatrix4fv(p.transform, 1, false, &transform.Data[0])
		g.Uniform4fv(p.scalar, 2, &s.Scalar[0])
		g.Uniform4fv(p.vector, 8, &s.Vector[0])
		g.Uniform1ui(p.clipSize, uint32(s.ClipSize))
		g.UniformMatrix4fv(p.clip, 8, false, &s.Clip[0])
		for slot, id := range []uint32{s.Texture1, s.Texture2, s.Texture3} {
			g.ActiveTexture(0x84C0 + uint32(slot))
			g.BindTexture(0x0DE1, d.textures[id].name)
		}
		if s.Scissor {
			g.Enable(0x0C11)
			r := s.ScissorRect
			g.Scissor(r.Left, r.Top, r.Right-r.Left, r.Bottom-r.Top)
		} else {
			g.Disable(0x0C11)
		}
		if s.Blend {
			g.Enable(0x0BE2)
		} else {
			g.Disable(0x0BE2)
		}
		geom, ok := d.geometry[c.GeometryID]
		if !ok {
			panic("draw of missing geometry")
		}
		g.BindVertexArray(geom.vao)
		g.DrawElements(0x0004, int32(c.Count), 0x1405, uintptr(c.Offset)*4)
		d.draws++
	}
	d.commands = d.commands[:0]
	g.BindVertexArray(0)
	g.BindFramebuffer(0x8D40, 0)
	g.Disable(0x0C11)
	g.ActiveTexture(0x84C0)
	g.check()
}
func (d *gpuDriver) close() {
	for id := range d.geometry {
		d.destroyGeometry(id)
	}
	for id := range d.buffers {
		d.destroyBuffer(id)
	}
	for id := range d.textures {
		d.destroyTexture(id)
	}
	for _, p := range d.programs {
		d.g.DeleteProgram(p.id)
	}
}
