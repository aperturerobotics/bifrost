package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()
	fmt.Printf("listening on %s\n", pc.LocalAddr().String())

	go func() {
		_, err := pc.WriteTo([]byte(string("testing")), pc.LocalAddr())
		if err != nil {
			panic(err)
		}
	}()

	b := make([]byte, 1024)
	_, addr, err := pc.ReadFrom(b)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("received packet from %v: %v\n", addr.String(), string(b))
}
