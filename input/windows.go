//go:build windows

package input

import (
	"syscall"
	"unsafe"

	"gomuxinput/protocol"
)

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
)

// WindowsSender converts protocol events to Windows input events via SendInput.
type WindowsSender struct{}

// Send injects the given event using the Windows SendInput API. This is a
// minimal placeholder implementation.
func (s *WindowsSender) Send(ev *protocol.Event) error {
	// Placeholder: map protocol.Event to INPUT structure.
	// In a real implementation you would fill in the INPUT struct according
	// to ev.Type, ev.Code and ev.Value. For simplicity we emit no real input.
	var input [1]byte // dummy
	_, _, err := procSendInput.Call(0, uintptr(unsafe.Pointer(&input[0])), 0)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
