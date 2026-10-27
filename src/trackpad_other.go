//go:build !linux

package main

import (
	"fmt"

	"github.com/crosshj/audio-utility/src/input"
)

func startTrackpadPointer() (<-chan input.Event, error) {
	return nil, fmt.Errorf("raw trackpad input is only available on Linux KMSDRM")
}
