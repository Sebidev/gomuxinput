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

// evdev event type constant for keys.
const evKey = 0x01

// Send injects the given event using the Windows SendInput API.

// minimal map from evdev KEY_* codes to Windows VK_* codes.
var keyMap = map[uint16]uint16{
	30: 0x41, // KEY_A -> VK_A
	48: 0x42, // KEY_B
	46: 0x43, // KEY_C
	32: 0x44, // KEY_D
	18: 0x45, // KEY_E
	33: 0x46, // KEY_F
	34: 0x47, // KEY_G
	35: 0x48, // KEY_H
	23: 0x49, // KEY_I
	36: 0x4A, // KEY_J
	37: 0x4B, // KEY_K
	38: 0x4C, // KEY_L
	50: 0x4D, // KEY_M
	49: 0x4E, // KEY_N
	24: 0x4F, // KEY_O
	25: 0x50, // KEY_P
	16: 0x51, // KEY_Q
	19: 0x52, // KEY_R
	31: 0x53, // KEY_S
	20: 0x54, // KEY_T
	22: 0x55, // KEY_U
	47: 0x56, // KEY_V
	17: 0x57, // KEY_W
	45: 0x58, // KEY_X
	21: 0x59, // KEY_Y
	44: 0x5A, // KEY_Z
	2:  0x31, // KEY_1
	3:  0x32, // KEY_2
	4:  0x33, // KEY_3
	5:  0x34, // KEY_4
	6:  0x35, // KEY_5
	7:  0x36, // KEY_6
	8:  0x37, // KEY_7
	9:  0x38, // KEY_8
	10: 0x39, // KEY_9
	11: 0x30, // KEY_0
}

func (s *WindowsSender) Send(ev *protocol.Event) error {
	if ev.Type != evKey {
		return nil // unsupported event type
	}

	vk, ok := keyMap[ev.Code]
	if !ok {
		// unmapped key
		return nil
	}

	const (
		inputKeyboard  = 1
		keyeventfKeyup = 0x0002
	)

	type KEYBDINPUT struct {
		wVk         uint16
		wScan       uint16
		dwFlags     uint32
		time        uint32
		dwExtraInfo uintptr
	}

	type INPUT struct {
		Type uint32
		Ki   KEYBDINPUT
	}

	var flags uint32
	if ev.Value == 0 {
		flags = keyeventfKeyup
	}

	in := INPUT{
		Type: inputKeyboard,
		Ki: KEYBDINPUT{
			wVk:     vk,
			dwFlags: flags,
		},
	}

	_, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
