package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net"

	"gomuxinput/input"
	"gomuxinput/protocol"
)

func main() {
	var addr = flag.String("addr", "127.0.0.1:3333", "server address")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	log.Printf("connected to %s", *addr)

	dec := gob.NewDecoder(conn)
	sender := &input.WindowsSender{}

	for {
		var ev protocol.Event
		if err := dec.Decode(&ev); err != nil {
			log.Fatalf("decode: %v", err)
		}
		if err := sender.Send(&ev); err != nil {
			log.Printf("send input: %v", err)
		}
	}
}
