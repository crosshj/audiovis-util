// Check all Go binding symbols and required drivers without a physical display.
package main

import (
	"fmt"
	"github.com/Zyko0/go-sdl3/sdl"
	"log"
)

func main() {
	if err := sdl.LoadLibrary(sdl.Path()); err != nil {
		log.Fatal(err)
	}
	defer sdl.CloseLibrary()
	fmt.Printf("SDL %s: Go binding symbols loaded\n", sdl.GetVersion())
	kms, gles := false, false
	for i := 0; i < sdl.GetNumVideoDrivers(); i++ {
		driver := sdl.GetVideoDriver(i)
		fmt.Println("Video driver:", driver)
		kms = kms || driver == "kmsdrm"
	}
	for i := 0; i < sdl.GetNumRenderDrivers(); i++ {
		driver := sdl.GetRenderDriver(i)
		fmt.Println("Renderer:", driver)
		gles = gles || driver == "opengles2"
	}
	if !kms || !gles {
		log.Fatal("KMSDRM or OpenGL ES renderer missing")
	}
	fmt.Println("PASS: runtime symbols and required drivers available; physical rendering still requires a hardware test.")
}
