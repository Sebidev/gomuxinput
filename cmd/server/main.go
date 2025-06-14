package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net"
	"strings"
	"sync"

	"gomuxinput/input"
	"gomuxinput/protocol"
)

func main() {
	var (
		addr     = flag.String("addr", "127.0.0.1:3333", "address to listen on")
		devPaths = flag.String("dev", "/dev/input/event0", "comma-separated evdev devices")
	)
	flag.Parse()

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
			log.Fatalf("encode: %v", err)
		}
	}
}
