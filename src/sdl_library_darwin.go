//go:build darwin

package main

import "github.com/Zyko0/go-sdl3/bin/binsdl"

func loadSDL() func() {
	library := binsdl.Load()
	return library.Unload
}
