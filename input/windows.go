//go:build windows

package input

import (
	"syscall"
	"unsafe"

	"gomuxinput/protocol"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	procSendInput       = user32.NewProc("SendInput")
)

const (
	INPUT_MOUSE    = 0
	INPUT_KEYBOARD = 1

	MOUSEEVENTF_MOVE      = 0x0001
	MOUSEEVENTF_LEFTDOWN  = 0x0002
	MOUSEEVENTF_LEFTUP    = 0x0004
	MOUSEEVENTF_RIGHTDOWN = 0x0008
	MOUSEEVENTF_RIGHTUP   = 0x0010
	MOUSEEVENTF_MIDDLEDOWN = 0x0020
	MOUSEEVENTF_MIDDLEUP   = 0x0040
	MOUSEEVENTF_XDOWN      = 0x0080
	MOUSEEVENTF_XUP        = 0x0100
	MOUSEEVENTF_WHEEL      = 0x0800

	KEYEVENTF_KEYUP = 0x0002
)

// EvdevToVK maps evdev key codes to Windows virtual key codes.
var evdevToVK = map[uint16]uint16{
	// Alphanumerisch
	30:  0x41, // A
	48:  0x42, // B
	46:  0x43, // C
	32:  0x44, // D
	18:  0x45, // E
	33:  0x46, // F
	34:  0x47, // G
	35:  0x48, // H
	23:  0x49, // I
	36:  0x4A, // J
	37:  0x4B, // K
	38:  0x4C, // L
	50:  0x4D, // M
	49:  0x4E, // N
	24:  0x4F, // O
	25:  0x50, // P
	16:  0x51, // Q
	19:  0x52, // R
	31:  0x53, // S
	20:  0x54, // T
	22:  0x55, // U
	47:  0x56, // V
	17:  0x57, // W
	45:  0x58, // X
	21:  0x59, // Y
	44:  0x5A, // Z

	// Zahlen (oben)
	2:  0x31, // 1
	3:  0x32, // 2
	4:  0x33, // 3
	5:  0x34, // 4
	6:  0x35, // 5
	7:  0x36, // 6
	8:  0x37, // 7
	9:  0x38, // 8
	10: 0x39, // 9
	11: 0x30, // 0

	// Sonderzeichen
	1:  0x1B, // ESC
	14: 0x08, // BACKSPACE
	15: 0x09, // TAB
	28: 0x0D, // ENTER
	57: 0x20, // SPACE
	12: 0xBD, // Minus
	13: 0xBB, // Gleichheitszeichen
	26: 0xDB, // [
	27: 0xDD, // ]
	39: 0xBA, // ;
	40: 0xDE, // '
	41: 0xC0, // `
	43: 0xDC, // Backslash
	51: 0xBC, // ,
	52: 0xBE, // .
	53: 0xBF, // /

	// Funktionstasten
	59: 0x70, // F1
	60: 0x71, // F2
	61: 0x72, // F3
	62: 0x73, // F4
	63: 0x74, // F5
	64: 0x75, // F6
	65: 0x76, // F7
	66: 0x77, // F8
	67: 0x78, // F9
	68: 0x79, // F10
	87: 0x7A, // F11
	88: 0x7B, // F12

	// Steuerung
	29: 0x11, // LEFT CTRL
	42: 0x10, // LEFT SHIFT
	56: 0x12, // LEFT ALT
	361: 0x5B, // LEFT META / WIN

	97: 0x11, // RIGHT CTRL
	54: 0x10, // RIGHT SHIFT
	100: 0x12, // RIGHT ALT
	367: 0x5C, // RIGHT META / WIN

	// Navigation
	75: 0x25, // LEFT
	77: 0x27, // RIGHT
	72: 0x26, // UP
	80: 0x28, // DOWN
	71: 0x24, // HOME
	79: 0x23, // END
	73: 0x21, // PAGE UP
	81: 0x22, // PAGE DOWN
	82: 0x2D, // INSERT
	83: 0x2E, // DELETE

	// NumPad
	69: 0x90, // NUM LOCK
	55: 0x6A, // KP *
	74: 0x6D, // KP -
	78: 0x6B, // KP +
	28: 0x0D, // KP ENTER (gleich mit normalem ENTER)
	83: 0x6E, // KP .
	71: 0x60, // KP 0
	72: 0x61, // KP 1
	73: 0x62, // KP 2
	74: 0x63, // KP 3
	75: 0x64, // KP 4
	76: 0x65, // KP 5
	77: 0x66, // KP 6
	78: 0x67, // KP 7
	79: 0x68, // KP 8
	80: 0x69, // KP 9
}

type keyboardInput struct {
	Type  uint32
	Vk    uint16
	Scan  uint16
	Flags uint32
	Time  uint32
	Extra uintptr
}

type mouseInput struct {
	Dx      int32
	Dy      int32
	MouseData uint32
	Flags   uint32
	Time    uint32
	Extra   uintptr
}

type inputUnion struct {
	Ki keyboardInput
	Mi mouseInput
}

type input struct {
	Type uint32
	U    inputUnion
}

// WindowsSender converts protocol events to Windows input events via SendInput.
type WindowsSender struct {
	relX int32
	relY int32
}

// Send injects the given event using the Windows SendInput API.
func (s *WindowsSender) Send(ev *protocol.Event) error {
	switch ev.Type {
	case 1: // EV_KEY
		switch ev.Code {
		case 272: // BTN_LEFT
			return s.sendMouseButton(MOUSEEVENTF_LEFTDOWN, MOUSEEVENTF_LEFTUP, ev.Value)
		case 273: // BTN_RIGHT
			return s.sendMouseButton(MOUSEEVENTF_RIGHTDOWN, MOUSEEVENTF_RIGHTUP, ev.Value)
		case 274: // BTN_MIDDLE
			return s.sendMouseButton(MOUSEEVENTF_MIDDLEDOWN, MOUSEEVENTF_MIDDLEUP, ev.Value)
		case 275:
			return s.sendMouseButton(MOUSEEVENTF_XDOWN, MOUSEEVENTF_XUP, ev.Value, 1)
		case 276:
			return s.sendMouseButton(MOUSEEVENTF_XDOWN, MOUSEEVENTF_XUP, ev.Value, 2)
		default:
			vk, ok := evdevToVK[ev.Code]
			if !ok {
				return nil // unmapped key
			}
			flags := uint32(0)
			if ev.Value == 0 {
				flags |= KEYEVENTF_KEYUP
			}
			in := input{
				Type: INPUT_KEYBOARD,
				U: inputUnion{
					Ki: keyboardInput{
						Vk:    vk,
						Scan:  0,
						Flags: flags,
						Time:  0,
						Extra: 0,
					},
				},
			}
			_, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
			if err != syscall.Errno(0) {
				return err
			}
			return nil
	case 2: // EV_REL
		switch ev.Code {
		case 0:
			s.relX += int32(ev.Value)
		case 1:
			s.relY += int32(ev.Value)
		case 8:
			// REL_WHEEL
			return s.sendWheel(int32(ev.Value))
		}
	}
	if s.relX != 0 || s.relY != 0 {
		in := input{
			Type: INPUT_MOUSE,
			U: inputUnion{
				Mi: mouseInput{
					Dx:    s.relX,
					Dy:    s.relY,
					Flags: MOUSEEVENTF_MOVE,
				},
			},
		}
		s.relX = 0
		s.relY = 0
		_, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
		if err != syscall.Errno(0) {
			return err
		}
	}
	return nil
}

func (s *WindowsSender) sendMouseButton(downFlag, upFlag uint32, value int32, xButton ...uint32) error {
	flag := downFlag
	if value == 0 {
		flag = upFlag
	}
	mouseData := uint32(0)
	if len(xButton) > 0 {
		mouseData = xButton[0]
	}
	in := input{
		Type: INPUT_MOUSE,
		U: inputUnion{
			Mi: mouseInput{
				Flags:     flag,
				MouseData: mouseData,
			},
		},
	}
	_, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func (s *WindowsSender) sendWheel(amount int32) error {
	in := input{
		Type: INPUT_MOUSE,
		U: inputUnion{
			Mi: mouseInput{
				Flags:     MOUSEEVENTF_WHEEL,
				MouseData: uint32(amount * 120),
			},
		},
	}
	_, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}
