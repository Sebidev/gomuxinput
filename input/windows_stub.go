//go:build !windows

package input

import "gomuxinput/protocol"

// WindowsSender is a stub for non-Windows builds.
type WindowsSender struct{}

// Send does nothing on non-Windows platforms.
func (s *WindowsSender) Send(ev *protocol.Event) error {
	// Placeholder stub
	_ = ev
	return nil
}
