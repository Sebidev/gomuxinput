//go:build linux

package input

import (
	"encoding/binary"
	"io"
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
	const eventSize = 24
	buf := make([]byte, eventSize)
	_, err := io.ReadFull(r.f, buf)
	if err != nil {
		return nil, err
	}
	ev := &protocol.Event{
		Type:  binary.LittleEndian.Uint16(buf[16:18]),
		Code:  binary.LittleEndian.Uint16(buf[18:20]),
		Value: int32(binary.LittleEndian.Uint32(buf[20:24])),
	}
	return ev, nil
}