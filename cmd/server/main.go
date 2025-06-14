package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net"
	"os"
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
	defer ln.Close()

	log.Printf("Server gestartet, wartet auf Verbindung unter %s", *addr)

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("accept: %v", err)
	}
	log.Printf("Client verbunden: %s", conn.RemoteAddr())

	enc := gob.NewEncoder(conn)
	var readers []*input.LinuxReader

	for _, p := range paths {
		r, err := input.OpenLinuxReader(strings.TrimSpace(p))
		if err != nil {
			log.Fatalf("open input %s: %v", p, err)
		}
		readers = append(readers, r)
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
					log.Printf("Reader beendet (%v)", err)
					return
				}
				evCh <- ev
			}
		}(r)
	}

	// Encoder-Loop (solange Client da)
	go func() {
		for ev := range evCh {
			if err := enc.Encode(ev); err != nil {
				log.Printf("Verbindung verloren: %v", err)
				break
			}
		}

		// Verbindung verloren → evdev Reader schließen → beendet alle Goroutinen
		for _, r := range readers {
			_ = r.Close()
		}
	}()

	// Warte bis alle Reader-Goroutinen durch sind
	wg.Wait()
	log.Println("Alle Reader beendet. Server fährt runter.")
	conn.Close()
	os.Exit(0)
}
