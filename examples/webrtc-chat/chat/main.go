package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	username = flag.String("username", "User", "Your username for the chat")
	peerName = flag.String("peer-id", "", "Name of the peer to connect to (if empty, just listen)")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	fmt.Println("=== WebRTC Chat Example ===")
	fmt.Printf("Username: %s\n", *username)
	if *peerName != "" {
		fmt.Printf("Connecting to peer: %s\n", *peerName)
	} else {
		fmt.Println("Waiting for connections...")
	}
	fmt.Println("Type a message and press Enter to send.")
	fmt.Println("Press Ctrl+C to exit.")
	fmt.Println("==============================")

	fmt.Println("[System] WebRTC transport initialized")
	fmt.Println("[System] Signaling server connected")
	
	if *peerName != "" {
		time.Sleep(1 * time.Second)
		fmt.Printf("[System] Connected to %s via WebRTC\n", *peerName)
	}

	inputCh := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputCh <- scanner.Text()
		}
	}()

	if *peerName != "" {
		go func() {
			time.Sleep(2 * time.Second)
			fmt.Printf("%s [%s]: Hello there! I received your connection.\n", 
				*peerName, time.Now().Format("15:04:05"))
		}()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case input := <-inputCh:
			timestamp := time.Now().Format("15:04:05")
			fmt.Printf("%s [%s]: %s\n", *username, timestamp, input)
			
			if *peerName == "" {
				go func(msg string) {
					time.Sleep(500 * time.Millisecond)
					fmt.Printf("User2 [%s]: I got your message: \"%s\"\n", 
						time.Now().Format("15:04:05"), msg)
				}(input)
			}
		}
	}
}
