package main

import (
	"fmt"
	"log"

	"github.com/NebulousLabs/go-upnp"
)

func main() {
	fmt.Println("discovering")
	// connect to router
	d, err := upnp.Discover()
	if err != nil {
		log.Fatal(err)
	}

	// discover external IP
	ip, err := d.ExternalIP()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Your external IP is:", ip)

	// forward a port
	err = d.Forward(9001, "upnp test")
	if err != nil {
		log.Fatal(err)
	}

	// un-forward a port
	err = d.Clear(9001)
	if err != nil {
		log.Fatal(err)
	}

	// record router's location
	loc := d.Location()

	// connect to router directly
	d, err = upnp.Load(loc)
	if err != nil {
		log.Fatal(err)
	}
}
