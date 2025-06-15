package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"gomuxinput/input"
	"gomuxinput/protocol"
)

var (
	mode        = flag.String("mode", "client", "mode: server or client")
	addr        = flag.String("addr", "127.0.0.1:3333", "server address")
	devPaths    = flag.String("dev", "/dev/input/event0", "comma-separated evdev devices")
	toggleCombo = flag.String("toggle", "ctrl+alt+q", "key combo to toggle input forwarding")
)

var nameToCode = map[string]uint16{
	"esc":        1,
	"1":          2,
	"2":          3,
	"3":          4,
	"4":          5,
	"5":          6,
	"6":          7,
	"7":          8,
	"8":          9,
	"9":          10,
	"0":          11,
	"-":          12,
	"=":          13,
	"backspace":  14,
	"tab":        15,
	"q":          16,
	"w":          17,
	"e":          18,
	"r":          19,
	"t":          20,
	"y":          21,
	"u":          22,
	"i":          23,
	"o":          24,
	"p":          25,
	"[":          26,
	"]":          27,
	"enter":      28,
	"ctrl":       29,
	"a":          30,
	"s":          31,
	"d":          32,
	"f":          33,
	"g":          34,
	"h":          35,
	"j":          36,
	"k":          37,
	"l":          38,
	";":          39,
	"'":          40,
	"`":          41,
	"shift":      42,
	"\\":         43,
	"z":          44,
	"x":          45,
	"c":          46,
	"v":          47,
	"b":          48,
	"n":          49,
	"m":          50,
	",":          51,
	".":          52,
	"/":          53,
	"rshift":     54,
	"alt":        56,
	"space":      57,
	"capslock":   58,
	"f1":         59,
	"f2":         60,
	"f3":         61,
	"f4":         62,
	"f5":         63,
	"f6":         64,
	"f7":         65,
	"f8":         66,
	"f9":         67,
	"f10":        68,
	"numlock":    69,
	"scrolllock": 70,
	"f11":        87,
	"f12":        88,
	"left":       105,
	"right":      106,
	"up":         103,
	"down":       108,
}

func parseCombo(combo string) []uint16 {
	var codes []uint16
	for _, part := range strings.Split(combo, "+") {
		if code, ok := nameToCode[strings.ToLower(part)]; ok {
			codes = append(codes, code)
		}
	}
	return codes
}

func isToggleCombo(ev *protocol.Event, combo []uint16, pressed map[uint16]bool) bool {
	if ev.Type != 1 {
		return false
	}

	if ev.Value == 1 {
		pressed[ev.Code] = true
	} else if ev.Value == 0 {
		pressed[ev.Code] = false
	}

	for _, code := range combo {
		if !pressed[code] {
			return false
		}
	}
	return true
}

func runClient() {
	comboCodes := parseCombo(*toggleCombo)
	pressed := make(map[uint16]bool)
	enabled := true
	sender := &input.WindowsSender{}

	for {
		log.Printf("connecting to %s ...", *addr)
		conn, err := net.Dial("tcp", *addr)
		if err != nil {
			log.Printf("connection failed: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		log.Printf("connected to %s", *addr)
		dec := gob.NewDecoder(conn)

		for {
			var ev protocol.Event
			if err := dec.Decode(&ev); err != nil {
				log.Printf("connection lost: %v", err)
				conn.Close()
				break
			}

			if isToggleCombo(&ev, comboCodes, pressed) {
				enabled = !enabled
				log.Printf("Input forwarding: %v", enabled)
				continue
			}

			if enabled {
				if err := sender.Send(&ev); err != nil {
					log.Printf("send input: %v", err)
				}
			}
		}
	}
}

func runServer() {
	paths := strings.Split(*devPaths, ",")
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("waiting for client on %s", *addr)

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("accept: %v", err)
	}
	defer conn.Close()
	log.Printf("client connected: %s", conn.RemoteAddr())

	enc := gob.NewEncoder(conn)

	var readers []*input.LinuxReader
	for _, p := range paths {
		r, err := input.OpenLinuxReader(strings.TrimSpace(p))
		if err != nil {
			log.Fatalf("open input %s: %v", p, err)
		}
		readers = append(readers, r)
		defer r.Close()
	}

	evCh := make(chan *protocol.Event)
	var wg sync.WaitGroup
	for _, r := range readers {
		wg.Add(1)
		go func(rd *input.LinuxReader) {
			defer wg.Done()
			for {
				ev, err := rd.ReadEvent()
				if err != nil {
					log.Printf("read event: %v", err)
					return
				}
				evCh <- ev
			}
		}(r)
	}

	go func() {
		wg.Wait()
		close(evCh)
	}()

	for ev := range evCh {
		if err := enc.Encode(ev); err != nil {
			log.Printf("encode: %v", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	if *mode == "client" {
		runClient()
	} else {
		runServer()
	}
}
