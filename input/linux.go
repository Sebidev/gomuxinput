//go:build linux

package input

import (
	"encoding/binary"
	"os"

	"gomuxinput/protocol"
)

// LinuxReader reads input events from an evdev device.
type LinuxReader struct {
	f *os.File
}

// OpenLinuxReader opens the given evdev device (e.g. /dev/input/event0).
func OpenLinuxReader(devPath string) (*LinuxReader, error) {
	f, err := os.Open(devPath)
	if err != nil {
		return nil, err
	}
	return &LinuxReader{f: f}, nil
}

// Close closes the underlying device.
func (r *LinuxReader) Close() error {
	if r.f != nil {
		return r.f.Close()
	}
	return nil
}

// ReadEvent reads a single input event from the device.
func (r *LinuxReader) ReadEvent() (*protocol.Event, error) {
	// struct input_event { struct timeval time; unsigned short type, code; int value; };
	var (
		sec  int64
		usec int64
		typ  uint16
		code uint16
		val  int32
	)
	// timeval
	if err := binary.Read(r.f, binary.LittleEndian, &sec); err != nil {
		return nil, err
	}
	if err := binary.Read(r.f, binary.LittleEndian, &usec); err != nil {
		return nil, err
	}
	if err := binary.Read(r.f, binary.LittleEndian, &typ); err != nil {
		return nil, err
	}
	if err := binary.Read(r.f, binary.LittleEndian, &code); err != nil {
		return nil, err
	}
	if err := binary.Read(r.f, binary.LittleEndian, &val); err != nil {
		return nil, err
	}
	// No alignment adjustment is required in this simplified reader. In a
	// full implementation you may need to account for struct padding.

	ev := &protocol.Event{Type: typ, Code: code, Value: val}
	_ = sec
	_ = usec
	return ev, nil
}
