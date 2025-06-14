# gomuxinput

This project forwards keyboard and mouse events from a Linux host to a Windows client via TCP.
It contains a simple server that reads evdev input events and sends them serialized over the network.
The client receives events and injects them using Windows `SendInput`.

The code is a minimal skeleton meant for further development and does not implement full event translation.

## Usage

The server can forward events from multiple evdev devices. Provide a comma-separated
list of event device paths via the `-dev` flag:

```
go run ./cmd/server -dev /dev/input/event3,/dev/input/event4
```

The client connects to the server and injects the received events:

```
go run ./cmd/client
```
