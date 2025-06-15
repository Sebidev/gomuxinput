# gomuxinput

**gomuxinput** is a lightweight tool that forwards keyboard and mouse events from a Linux host to a Windows client over TCP.

The server captures input events from evdev devices on Linux and sends them over the network. The client receives these events and injects them into the Windows system using the `SendInput` API.

> ⚠️ This is a minimal working prototype. Full input translation (e.g., modifier state, layout conversion) is not implemented.

---

## Usage

### Server (Linux)
The server reads input events from one or more evdev input devices.
Provide a comma-separated list of device paths using the `-dev` flag:

```bash
go run ./cmd/main.go -mode=server -dev=/dev/input/event0,/dev/input/event1
```

### Client (Windows)
The client connects to the server and injects the received input events:

```bash
go run ./cmd/main.go -mode=client -addr=192.168.1.10:3333 -toggle=ctrl+alt+q
```

- The `-toggle` flag sets a key combination to enable/disable input forwarding.
- If not set, it defaults to `ctrl+alt+q`.

---

## Building

### Statically Linked Binary for Linux (server)
```bash
cd cmd
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gomuxinput
```

### Statically Linked Binary for Windows (client)
```bash
cd cmd
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o gomuxinput.exe
```

---

## Features
- Keyboard and mouse input forwarding
- Multiple input devices supported
- TCP-based communication (UDP planned)
- Auto-reconnect if the connection is lost
- Toggle forwarding via configurable hotkey (default: `ctrl+alt+q`)
- Single binary for both client and server (via `-mode` flag)

---

## Project Structure
- `cmd/main.go`: combined client/server entry point
- `input/`: platform-specific input handlers
- `protocol/`: event data structures

---

## Notes
- Running the server requires root access on Linux to read `/dev/input/event*`.
- You may configure udev rules to allow non-root access if needed.
- No encryption/authentication is implemented — use only in trusted networks.

---

## Planned Features
- UDP support
- Drag-and-drop and double-click simulation
- More key mappings and layout support
- Optional GUI/tray icon for Windows
- TLS encryption for secure networks

---

Feel free to fork and extend this tool to fit your workflow!
