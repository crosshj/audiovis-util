//go:build linux

package main

import (
	"log"

	"github.com/Zyko0/go-sdl3/sdl"
)

func loadSDL() func() {
	if err := sdl.LoadLibrary(sdl.Path()); err != nil {
		log.Fatal(err)
	}
	return func() {
		if err := sdl.CloseLibrary(); err != nil {
			log.Print(err)
		}
	}
}
