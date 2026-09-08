package inproc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/s4wave/spacewave/net/link"
)

// TestNetworkIsolationAndAttachmentLifetime exercises real links across scoped packet networks.
func TestNetworkIsolationAndAttachmentLifetime(t *testing.T) {
	// Separate network values do not make their peers reachable to each other.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	firstBed, _ := buildTestbed(t, ctx)
	secondBed, _ := buildTestbed(t, ctx)
	_, first, firstRef := execPeer(ctx, t, firstBed, nil)
	t.Cleanup(firstRef.Release)
	_, second, secondRef := execPeer(ctx, t, secondBed, nil)
	t.Cleanup(secondRef.Release)
	network := NewNetwork()
	detachFirst, err := network.Attach(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(detachFirst)
	other := NewNetwork()
	detachOther, err := other.Attach(ctx, second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(detachOther)
	isolatedCtx, cancelIsolated := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancelIsolated()
	if _, err := second.GetPeerDialer(isolatedCtx, first.GetPeerID()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("separate networks resolved a peer: %v", err)
	}

	// Moving into the same network supplies automatic dialing without static peer maps.
	detachOther()
	detachSecond, err := network.Attach(ctx, second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(detachSecond)
	if _, err := network.Attach(ctx, second); err == nil {
		t.Fatal("duplicate peer attachment succeeded")
	}
	detachSecond()
	currentDetach, err := network.Attach(ctx, second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(currentDetach)
	detachSecond()
	linked, release, err := link.EstablishLinkWithPeerEx(ctx, secondBed.Bus, second.GetPeerID(), first.GetPeerID(), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(release)
	if linked.GetRemotePeer() != first.GetPeerID() {
		t.Fatal("link did not authenticate the requested peer")
	}
}
