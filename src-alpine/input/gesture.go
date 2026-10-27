// Package input translates device-level input into gestures the UI can consume.
package input

import "math"

type Point struct {
	X float32
	Y float32
}

type Button uint8

const LeftButton Button = 1

type InputKind uint8

const (
	AbsolutePointer InputKind = iota
	FingerDown
	FingerMove
	FingerUp
	ButtonDown
	ButtonUp
)

// Event is normalized by a device adapter before it reaches the recognizer.
// Position and Delta use the app's logical canvas coordinates.
type Event struct {
	Kind     InputKind
	FingerID int32
	Position Point
	Delta    Point
	Button   Button
}

type GestureKind uint8

const (
	PointerMoved GestureKind = iota
	PointerPressed
	PointerReleased
	DragStarted
	DragMoved
	DragEnded
)

// Gesture is the stable input vocabulary consumed by app controls.
type Gesture struct {
	Kind     GestureKind
	Position Point
	Delta    Point
	Button   Button
}

type contact struct {
	order uint64
}

// Recognizer keeps pointer position independent of finger position. The first
// contact moves the pointer normally. During a held left click, whichever
// contact moves can continue the drag, including after that finger lifts and
// touches down again elsewhere on the pad.
type Recognizer struct {
	position      Point
	contacts      map[int32]contact
	primary       int32
	hasPrimary    bool
	clickOwner    int32
	hasClickOwner bool
	order         uint64
	buttons       map[Button]bool
}

func NewRecognizer(initial Point) *Recognizer {
	return &Recognizer{
		position: initial,
		contacts: make(map[int32]contact),
		buttons:  make(map[Button]bool),
	}
}

func (r *Recognizer) Position() Point { return r.position }

func (r *Recognizer) SetPosition(position Point) { r.position = position }

func (r *Recognizer) Handle(event Event) []Gesture {
	switch event.Kind {
	case AbsolutePointer:
		delta := Point{X: event.Position.X - r.position.X, Y: event.Position.Y - r.position.Y}
		r.position = event.Position
		return r.pointerMotion(delta)
	case FingerDown:
		if _, exists := r.contacts[event.FingerID]; exists {
			return nil
		}
		r.order++
		r.contacts[event.FingerID] = contact{order: r.order}
		if !r.hasPrimary {
			r.primary, r.hasPrimary = event.FingerID, true
		}
	case FingerMove:
		if _, exists := r.contacts[event.FingerID]; !exists {
			return nil
		}
		if !r.buttons[LeftButton] && (!r.hasPrimary || event.FingerID != r.primary) {
			return nil
		}
		return r.pointerMotion(event.Delta)
	case FingerUp:
		var gestures []Gesture
		if r.hasClickOwner && event.FingerID == r.clickOwner && r.buttons[LeftButton] {
			// Some KMSDRM input paths do not deliver the trackpad's physical
			// button-up event. The click finger's lift is the equivalent release.
			gestures = r.releaseButton(LeftButton)
		}
		delete(r.contacts, event.FingerID)
		if r.hasPrimary && event.FingerID == r.primary {
			r.hasPrimary = false
			var oldest uint64
			for id, c := range r.contacts {
				if !r.hasPrimary || c.order < oldest {
					r.primary, oldest, r.hasPrimary = id, c.order, true
				}
			}
		}
		return gestures
	case ButtonDown:
		if r.buttons[event.Button] {
			return nil
		}
		r.buttons[event.Button] = true
		if event.Button == LeftButton && r.hasPrimary {
			r.clickOwner, r.hasClickOwner = r.primary, true
		}
		gestures := []Gesture{{Kind: PointerPressed, Position: r.position, Button: event.Button}}
		if event.Button == LeftButton {
			gestures = append(gestures, Gesture{Kind: DragStarted, Position: r.position, Button: event.Button})
		}
		return gestures
	case ButtonUp:
		return r.releaseButton(event.Button)
	}
	return nil
}

func (r *Recognizer) releaseButton(button Button) []Gesture {
	if !r.buttons[button] {
		return nil
	}
	delete(r.buttons, button)
	if button == LeftButton {
		r.hasClickOwner = false
	}
	gestures := make([]Gesture, 0, 2)
	if button == LeftButton {
		gestures = append(gestures, Gesture{Kind: DragEnded, Position: r.position, Button: button})
	}
	gestures = append(gestures, Gesture{Kind: PointerReleased, Position: r.position, Button: button})
	return gestures
}

func (r *Recognizer) pointerMotion(delta Point) []Gesture {
	if math.Hypot(float64(delta.X), float64(delta.Y)) < 0.3 {
		return nil
	}
	r.position.X += delta.X
	r.position.Y += delta.Y
	gestures := []Gesture{{Kind: PointerMoved, Position: r.position, Delta: delta}}
	if r.buttons[LeftButton] {
		gestures = append(gestures, Gesture{Kind: DragMoved, Position: r.position, Delta: delta, Button: LeftButton})
	}
	return gestures
}
