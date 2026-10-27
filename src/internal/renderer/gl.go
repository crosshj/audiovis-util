//go:build linux

package renderer

import (
	"fmt"
	"github.com/ebitengine/purego"
	"strings"
	"unsafe"
)

type glAPI struct {
	GetString               func(uint32) string
	ActiveTexture           func(uint32)
	PixelStorei             func(uint32, int32)
	GenFramebuffers         func(int32, *uint32)
	BindFramebuffer         func(uint32, uint32)
	FramebufferTexture2D    func(uint32, uint32, uint32, uint32, int32)
	CheckFramebufferStatus  func(uint32) uint32
	DeleteFramebuffers      func(int32, *uint32)
	GenVertexArrays         func(int32, *uint32)
	BindVertexArray         func(uint32)
	DeleteVertexArrays      func(int32, *uint32)
	Uniform1ui              func(int32, uint32)
	Uniform4fv              func(int32, int32, *float32)
	UniformMatrix4fv        func(int32, int32, bool, *float32)
	Scissor                 func(int32, int32, int32, int32)
	ClearColor              func(float32, float32, float32, float32)
	Clear                   func(uint32)
	DrawElements            func(uint32, int32, uint32, uintptr)
	GetError                func() uint32
	Viewport                func(int32, int32, int32, int32)
	CreateShader            func(uint32) uint32
	ShaderSource            func(uint32, int32, *uintptr, *int32)
	CompileShader           func(uint32)
	GetShaderiv             func(uint32, uint32, *int32)
	GetShaderInfoLog        func(uint32, int32, *int32, *byte)
	DeleteShader            func(uint32)
	CreateProgram           func() uint32
	AttachShader            func(uint32, uint32)
	BindAttribLocation      func(uint32, uint32, string)
	LinkProgram             func(uint32)
	GetProgramiv            func(uint32, uint32, *int32)
	GetProgramInfoLog       func(uint32, int32, *int32, *byte)
	DeleteProgram           func(uint32)
	UseProgram              func(uint32)
	GetUniformLocation      func(uint32, string) int32
	Uniform1f               func(int32, float32)
	Uniform1i               func(int32, int32)
	Uniform2f               func(int32, float32, float32)
	GenBuffers              func(int32, *uint32)
	BindBuffer              func(uint32, uint32)
	BufferData              func(uint32, uintptr, unsafe.Pointer, uint32)
	DeleteBuffers           func(int32, *uint32)
	EnableVertexAttribArray func(uint32)
	VertexAttribPointer     func(uint32, int32, uint32, bool, int32, uintptr)
	DrawArrays              func(uint32, int32, int32)
	GenTextures             func(int32, *uint32)
	BindTexture             func(uint32, uint32)
	TexParameteri           func(uint32, uint32, int32)
	TexImage2D              func(uint32, int32, int32, int32, int32, int32, uint32, uint32, unsafe.Pointer)
	TexSubImage2D           func(uint32, int32, int32, int32, int32, int32, uint32, uint32, unsafe.Pointer)
	DeleteTextures          func(int32, *uint32)
	Enable                  func(uint32)
	Disable                 func(uint32)
	BlendFunc               func(uint32, uint32)
	ReadPixels              func(int32, int32, int32, int32, uint32, uint32, unsafe.Pointer)
}

func loadGL() *glAPI {
	h, e := purego.Dlopen("libGLESv2.so.2", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	must(e)
	g := &glAPI{}
	bindAPI(h, "gl", g)
	return g
}
func (g *glAPI) program(fragment string) uint32 {
	return g.programSources(`attribute vec2 position; varying vec2 uv; void main(){uv=vec2((position.x+1.)*.5,(1.-position.y)*.5);gl_Position=vec4(position,0.,1.);}`, fragment, []string{"position"})
}
func (g *glAPI) programSources(vertex, fragment string, attributes []string) uint32 {
	compile := func(kind uint32, source string) uint32 {
		s := g.CreateShader(kind)
		bytes := append([]byte(source), 0)
		p := uintptr(unsafe.Pointer(&bytes[0]))
		n := int32(len(source))
		g.ShaderSource(s, 1, &p, &n)
		g.CompileShader(s)
		var ok int32
		g.GetShaderiv(s, 0x8B81, &ok)
		if ok == 0 {
			b := make([]byte, 4096)
			g.GetShaderInfoLog(s, int32(len(b)), nil, &b[0])
			panic(strings.TrimRight(string(b), "\x00"))
		}
		return s
	}
	vs := compile(0x8B31, vertex)
	fs := compile(0x8B30, fragment)
	defer g.DeleteShader(vs)
	defer g.DeleteShader(fs)
	p := g.CreateProgram()
	g.AttachShader(p, vs)
	g.AttachShader(p, fs)
	for i, name := range attributes {
		g.BindAttribLocation(p, uint32(i), name)
	}
	g.LinkProgram(p)
	var ok int32
	g.GetProgramiv(p, 0x8B82, &ok)
	if ok == 0 {
		b := make([]byte, 4096)
		g.GetProgramInfoLog(p, int32(len(b)), nil, &b[0])
		panic(string(b))
	}
	return p
}
func (g *glAPI) check() {
	if e := g.GetError(); e != 0 {
		panic(fmt.Sprintf("GLES error 0x%x", e))
	}
}

const sceneShader = `precision mediump float;varying vec2 uv;uniform float elapsed;uniform float scene;uniform vec2 pointer;void main(){float wave=.5+.5*sin(uv.x*8.+uv.y*5.+elapsed);vec3 c=mix(vec3(.04,.09,.18),vec3(.12,.45,.44),wave);c=mix(c,c.bgr,mod(scene,2.));float cursor=1.-smoothstep(.003,.006,distance(uv,pointer));gl_FragColor=vec4(mix(c,vec3(1.),cursor),1.);}`

// The polygon tip is the input hotspot (0,0). Distance-based coverage keeps the
// white outline and black fill smooth at both supported HTML resolutions.
const uiShader = `precision highp float;
varying vec2 uv;
uniform sampler2D tex;
uniform vec2 pointer;
uniform vec2 uvScale;
void edge(vec2 p, vec2 a, vec2 b, inout float distanceToEdge, inout bool inside) {
    vec2 ab = b-a;
    vec2 nearest = a + ab * clamp(dot(p-a,ab)/dot(ab,ab),0.,1.);
    distanceToEdge = min(distanceToEdge,length(p-nearest));
    if ((a.y>p.y)!=(b.y>p.y)) {
        if (p.x < (b.x-a.x)*(p.y-a.y)/(b.y-a.y)+a.x) inside = !inside;
    }
}
void main() {
    vec4 color = texture2D(tex,uv*uvScale);
    vec2 p = (uv-pointer)*vec2(1280.,800.);
    if (p.x < -2. || p.x > 19. || p.y < -2. || p.y > 27.) {
        gl_FragColor = color;
        return;
    }
    float d = 100.;
    bool inside = false;
    edge(p,vec2(0.,0.),vec2(2.,20.),d,inside);
    edge(p,vec2(2.,20.),vec2(6.5,15.5),d,inside);
    edge(p,vec2(6.5,15.5),vec2(10.5,24.),d,inside);
    edge(p,vec2(10.5,24.),vec2(14.,22.3),d,inside);
    edge(p,vec2(14.,22.3),vec2(10.,14.),d,inside);
    edge(p,vec2(10.,14.),vec2(17.,14.),d,inside);
    edge(p,vec2(17.,14.),vec2(0.,0.),d,inside);
    float sd = inside ? -d : d;
    float outline = 1.-smoothstep(.55,1.05,sd);
    float fill = 1.-smoothstep(-.55,-.05,sd);
    color = mix(color,vec4(.98,.98,.98,1.),outline);
    gl_FragColor = mix(color,vec4(.025,.025,.03,1.),fill);
}`
