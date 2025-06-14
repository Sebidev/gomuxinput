package protocol

// Event represents a keyboard or mouse event transmitted over the wire.
type Event struct {
	// Type corresponds to evdev event type or Windows message.
	Type uint16
	// Code corresponds to key code or button code.
	Code uint16
	// Value carries the event value (press/release, relative movement, etc.).
	Value int32
}
