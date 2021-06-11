package main

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

func main() {
	// generate 32 bytes key
	d := make([]byte, 32)
	_, err := rand.Read(d)
	if err != nil {
		panic(err)
	}
	os.Stdout.WriteString(hex.EncodeToString(d))
	os.Stdout.WriteString("\n")
}
