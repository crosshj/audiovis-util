//go:build linux

package renderer

import (
	"fmt"
	"github.com/ebitengine/purego"
	"reflect"
	"unsafe"
)

type rect struct{ Left, Top, Right, Bottom int32 }

// Signatures follow the pinned 1.4 SDK C headers. Handles and callbacks are native pointers.
type ulAPI struct {
	ViewGetRenderTarget        func(uintptr) renderTarget
	ViewIsAccelerated          func(uintptr) bool
	BitmapGetWidth             func(uintptr) uint32
	BitmapGetHeight            func(uintptr) uint32
	BitmapGetFormat            func(uintptr) uint32
	BitmapGetBpp               func(uintptr) uint32
	BitmapGetRowBytes          func(uintptr) uint32
	BitmapIsEmpty              func(uintptr) bool
	BitmapLockPixels           func(uintptr) unsafe.Pointer
	BitmapUnlockPixels         func(uintptr)
	CreateString               func(string) uintptr
	DestroyString              func(uintptr)
	StringGetData              func(uintptr) unsafe.Pointer
	StringGetLength            func(uintptr) uintptr
	EnablePlatformFontLoader   func()
	EnablePlatformFileSystem   func(uintptr)
	EnableDefaultLogger        func(uintptr)
	CreateConfig               func() uintptr
	DestroyConfig              func(uintptr)
	CreateRenderer             func(uintptr) uintptr
	DestroyRenderer            func(uintptr)
	CreateViewConfig           func() uintptr
	DestroyViewConfig          func(uintptr)
	ViewConfigSetIsAccelerated func(uintptr, bool)
	ViewConfigSetIsTransparent func(uintptr, bool)
	CreateView                 func(uintptr, uint32, uint32, uintptr, uintptr) uintptr
	DestroyView                func(uintptr)
	ViewLoadURL                func(uintptr, uintptr)
	ViewEvaluateScript         func(uintptr, uintptr, *uintptr) uintptr
	ViewFocus                  func(uintptr)
	ViewLockJSContext          func(uintptr) uintptr
	ViewUnlockJSContext        func(uintptr)
	Update                     func(uintptr)
	RefreshDisplay             func(uintptr, uint32)
	Render                     func(uintptr)
	CreateMouseEvent           func(uint32, int32, int32, uint32) uintptr
	DestroyMouseEvent          func(uintptr)
	ViewFireMouseEvent         func(uintptr, uintptr)
	CreateKeyEvent             func(uint32, uint32, int32, int32, uintptr, uintptr, bool, bool, bool) uintptr
	DestroyKeyEvent            func(uintptr)
	ViewFireKeyEvent           func(uintptr, uintptr)
	CreateScrollEvent          func(uint32, int32, int32) uintptr
	DestroyScrollEvent         func(uintptr)
	ViewFireScrollEvent        func(uintptr, uintptr)
}
type jsAPI struct {
	ValueToStringCopy               func(uintptr, uintptr, *uintptr) uintptr
	StringGetMaximumUTF8CStringSize func(uintptr) uintptr
	StringGetUTF8CString            func(uintptr, unsafe.Pointer, uintptr) uintptr

	ValueToNumber                  func(uintptr, uintptr, *uintptr) float64
	StringCreateWithUTF8CString    func(string) uintptr
	StringRelease                  func(uintptr)
	ContextGetGlobalObject         func(uintptr) uintptr
	ObjectMakeFunctionWithCallback func(uintptr, uintptr, uintptr) uintptr
	ObjectSetProperty              func(uintptr, uintptr, uintptr, uintptr, uint32, *uintptr)
	ValueMakeNumber                func(uintptr, float64) uintptr
}

func bindAPI(h uintptr, prefix string, obj any) {
	v := reflect.ValueOf(obj).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		purego.RegisterLibFunc(v.Field(i).Addr().Interface(), h, prefix+t.Field(i).Name)
	}
}
func (u *ulAPI) text(p uintptr) string {
	if p == 0 {
		return ""
	}
	return string(unsafe.Slice((*byte)(u.StringGetData(p)), int(u.StringGetLength(p))))
}
func (u *ulAPI) with(s string, f func(uintptr)) {
	p := u.CreateString(s)
	defer u.DestroyString(p)
	f(p)
}
func (u *ulAPI) eval(v uintptr, s string) (string, error) {
	var result uintptr
	var exception uintptr
	u.with(s, func(p uintptr) { result = u.ViewEvaluateScript(v, p, &exception) })
	// Result and exception are borrowed from View, not caller-owned strings.
	if e := u.text(exception); e != "" {
		return "", fmt.Errorf("JavaScript: %s", e)
	}
	return u.text(result), nil
}
func (u *ulAPI) mouse(v uintptr, kind uint32, x, y int32, button uint32) {
	e := u.CreateMouseEvent(kind, x, y, button)
	defer u.DestroyMouseEvent(e)
	u.ViewFireMouseEvent(v, e)
}
func (u *ulAPI) key(v uintptr, kind, mods uint32, vk, native int32, text string, repeat bool) {
	u.with(text, func(s uintptr) {
		e := u.CreateKeyEvent(kind, mods, vk, native, s, s, false, repeat, false)
		defer u.DestroyKeyEvent(e)
		u.ViewFireKeyEvent(v, e)
	})
}

// BindCommand must be called on the renderer's owning OS thread.
func (r *Renderer) BindCommand(handler func(string, float64) float64) error {
	j, u, v := r.j, r.u, r.view
	callback := purego.NewCallback(func(ctx, fn, this, count uintptr, args unsafe.Pointer, exception uintptr) uintptr {
		if count != 2 {
			return j.ValueMakeNumber(ctx, -1)
		}
		values := unsafe.Slice((*uintptr)(args), 2)
		var ex uintptr
		name := j.ValueToStringCopy(ctx, values[0], &ex)
		if ex != 0 || name == 0 {
			return j.ValueMakeNumber(ctx, -1)
		}
		defer j.StringRelease(name)
		size := j.StringGetMaximumUTF8CStringSize(name)
		if size > 256 {
			return j.ValueMakeNumber(ctx, -1)
		}
		bytes := make([]byte, size)
		length := j.StringGetUTF8CString(name, unsafe.Pointer(&bytes[0]), size)
		value := j.ValueToNumber(ctx, values[1], &ex)
		if ex != 0 || length == 0 {
			return j.ValueMakeNumber(ctx, -1)
		}
		return j.ValueMakeNumber(ctx, handler(string(bytes[:length-1]), value))
	})
	ctx := u.ViewLockJSContext(v)
	defer u.ViewUnlockJSContext(v)
	name := j.StringCreateWithUTF8CString("nativeInvoke")
	defer j.StringRelease(name)
	fn := j.ObjectMakeFunctionWithCallback(ctx, name, callback)
	var exception uintptr
	j.ObjectSetProperty(ctx, j.ContextGetGlobalObject(ctx), name, fn, 0, &exception)
	if exception != 0 {
		return fmt.Errorf("native command registration failed")
	}
	return nil
}
