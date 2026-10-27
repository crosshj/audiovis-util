//go:build linux

package main

import "github.com/Zyko0/go-sdl3/sdl"

func keyMods(m sdl.Keymod) uint32 {
	var v uint32
	if m&sdl.KMOD_ALT != 0 {
		v |= 1
	}
	if m&sdl.KMOD_CTRL != 0 {
		v |= 2
	}
	if m&sdl.KMOD_GUI != 0 {
		v |= 4
	}
	if m&sdl.KMOD_SHIFT != 0 {
		v |= 8
	}
	return v
}
func virtualKey(k sdl.Keycode) int32 {
	if k >= sdl.K_A && k <= sdl.K_Z {
		return int32(k) - 32
	}
	if k >= sdl.K_0 && k <= sdl.K_9 {
		return int32(k)
	}
	switch k {
	case sdl.K_BACKSPACE:
		return 8
	case sdl.K_TAB:
		return 9
	case sdl.K_RETURN:
		return 13
	case sdl.K_ESCAPE:
		return 27
	case sdl.K_SPACE:
		return 32
	case sdl.K_PAGEUP:
		return 33
	case sdl.K_PAGEDOWN:
		return 34
	case sdl.K_END:
		return 35
	case sdl.K_HOME:
		return 36
	case sdl.K_LEFT:
		return 37
	case sdl.K_UP:
		return 38
	case sdl.K_RIGHT:
		return 39
	case sdl.K_DOWN:
		return 40
	case sdl.K_DELETE:
		return 46
	}
	return 0
}
