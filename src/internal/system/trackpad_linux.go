//go:build linux

package system

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	evSyn           = 0
	evKey           = 1
	evAbs           = 3
	synReport       = 0
	btnLeft         = 0x110
	absMTSlot       = 0x2f
	absMTPositionX  = 0x35
	absMTPositionY  = 0x36
	absMTTrackingID = 0x39
)

type inputAbsInfo struct {
	value      int32
	minimum    int32
	maximum    int32
	fuzz       int32
	flat       int32
	resolution int32
}

type multitouchSlot struct {
	trackingID      int32
	endedTrackingID int32
	x, y            int32
	lastX, lastY    int32
	active          bool
	haveX, haveY    bool
	haveLast        bool
	positionChanged bool
	downPending     bool
	upPending       bool
}

func StartTrackpad(canvasW, canvasH int) (<-chan Event, error) {
	device, err := findTrackpadEvent()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(device)
	if err != nil {
		return nil, fmt.Errorf("open trackpad event device %s: %w", device, err)
	}

	xMin, xMax, err := inputAxisRange(file, absMTPositionX)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("read trackpad X range: %w", err)
	}
	yMin, yMax, err := inputAxisRange(file, absMTPositionY)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("read trackpad Y range: %w", err)
	}
	slotMin, slotMax, err := inputAxisRange(file, absMTSlot)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("read trackpad slot range: %w", err)
	}
	if xMax <= xMin || yMax <= yMin || slotMax < slotMin || slotMax-slotMin > 31 {
		file.Close()
		return nil, fmt.Errorf("invalid trackpad ranges X=%d..%d Y=%d..%d slots=%d..%d", xMin, xMax, yMin, yMax, slotMin, slotMax)
	}

	events := make(chan Event, 256)
	log.Printf("Trackpad input: %s, X=%d..%d Y=%d..%d slots=%d..%d", device, xMin, xMax, yMin, yMax, slotMin, slotMax)
	go readTrackpadEvents(file, float32(xMax-xMin), float32(yMax-yMin), slotMin, slotMax, canvasW, canvasH, events)
	return events, nil
}

func findTrackpadEvent() (string, error) {
	entries, err := os.ReadDir("/sys/class/input")
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "event") {
			continue
		}
		if _, err := strconv.Atoi(strings.TrimPrefix(entry.Name(), "event")); err != nil {
			continue
		}
		name, err := os.ReadFile(filepath.Join("/sys/class/input", entry.Name(), "device/name"))
		if err == nil && strings.TrimSpace(string(name)) == "bcm5974" {
			return filepath.Join("/dev/input", entry.Name()), nil
		}
	}
	return "", fmt.Errorf("could not find bcm5974 trackpad event device")
}

func inputAxisRange(file *os.File, axis uint32) (int32, int32, error) {
	var info inputAbsInfo
	request := uintptr(0x80184540 + axis) // EVIOCGABS(axis) on 64-bit Linux.
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, file.Fd(), request, uintptr(unsafe.Pointer(&info)))
	if errno != 0 {
		return 0, 0, errno
	}
	return info.minimum, info.maximum, nil
}

func readTrackpadEvents(file *os.File, xRange, yRange float32, slotMin, slotMax int32, canvasW, canvasH int, events chan<- Event) {
	defer file.Close()
	var event [24]byte // struct input_event on the target's x86_64 Linux kernel.
	slots := make([]multitouchSlot, int(slotMax-slotMin+1))
	for i := range slots {
		slots[i].trackingID = -1
	}
	currentSlot := int32(0)

	for {
		if _, err := io.ReadFull(file, event[:]); err != nil {
			log.Printf("Trackpad input stopped: %v", err)
			return
		}
		typ := binary.LittleEndian.Uint16(event[16:18])
		code := binary.LittleEndian.Uint16(event[18:20])
		value := int32(binary.LittleEndian.Uint32(event[20:24]))
		switch typ {
		case evKey:
			// SDL's KMSDRM backend does not consistently surface the
			// bcm5974 physical click as mouse-button events. Read BTN_LEFT
			// from the same evdev stream as the contact slots so click-up
			// always ends a drag in the gesture recognizer.
			if code == btnLeft {
				switch value {
				case 0:
					events <- Event{Kind: ButtonUp, Button: LeftButton}
				case 1:
					events <- Event{Kind: ButtonDown, Button: LeftButton}
				}
			}
		case evAbs:
			switch code {
			case absMTSlot:
				currentSlot = value
			case absMTTrackingID:
				if i := int(currentSlot - slotMin); i >= 0 && i < len(slots) {
					slot := &slots[i]
					if value < 0 {
						if slot.active {
							slot.endedTrackingID = slot.trackingID
							slot.upPending = true
						}
						slot.active = false
						slot.trackingID = -1
					} else if !slot.active || slot.trackingID != value {
						if slot.active {
							slot.endedTrackingID = slot.trackingID
							slot.upPending = true
						}
						slot.active = true
						slot.trackingID = value
						slot.haveLast = false
						slot.downPending = true
					}
				}
			case absMTPositionX:
				if i := int(currentSlot - slotMin); i >= 0 && i < len(slots) {
					slots[i].x, slots[i].haveX, slots[i].positionChanged = value, true, true
				}
			case absMTPositionY:
				if i := int(currentSlot - slotMin); i >= 0 && i < len(slots) {
					slots[i].y, slots[i].haveY, slots[i].positionChanged = value, true, true
				}
			}
		case evSyn:
			if code != synReport {
				continue
			}
			for i := range slots {
				slot := &slots[i]
				if slot.upPending {
					events <- Event{Kind: FingerUp, FingerID: slot.endedTrackingID}
					slot.upPending = false
				}
				if slot.downPending {
					events <- Event{Kind: FingerDown, FingerID: slot.trackingID}
					slot.downPending = false
					if slot.haveX && slot.haveY {
						slot.lastX, slot.lastY, slot.haveLast = slot.x, slot.y, true
					}
					slot.positionChanged = false
					continue
				}
				if slot.active && slot.haveX && slot.haveY {
					if !slot.haveLast {
						slot.lastX, slot.lastY, slot.haveLast = slot.x, slot.y, true
					} else if slot.positionChanged {
						delta := Point{
							X: float32(slot.x-slot.lastX) / xRange * float32(canvasW),
							Y: float32(slot.y-slot.lastY) / yRange * float32(canvasH),
						}
						slot.lastX, slot.lastY = slot.x, slot.y
						if delta.X != 0 || delta.Y != 0 {
							events <- Event{Kind: FingerMove, FingerID: slot.trackingID, Delta: delta}
						}
					}
				}
				slot.positionChanged = false
			}
		}
	}
}
