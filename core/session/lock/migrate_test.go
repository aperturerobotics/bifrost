package session_lock

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

	"github.com/s4wave/spacewave/db/kvtx"
	kvtest "github.com/s4wave/spacewave/db/kvtx/kvtest"
	store_kvtx_inmem "github.com/s4wave/spacewave/db/store/kvtx/inmem"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/keypem"
)

func TestCopyCredentialPreservesPINAcrossTransactionRetry(t *testing.T) {
	skipGoScriptScryptCost(t)
	ctx := t.Context()
	source, destination := store_kvtx_inmem.NewStore(), store_kvtx_inmem.NewStore()
	key, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := keypem.MarshalPrivKeyPem(key)
	if err != nil {
		t.Fatal(err)
	}
	private, symmetric, config, err := CreatePINLock(plain, []byte("123456"))
	if err != nil {
		t.Fatal(err)
	}
	if err := WritePINLock(ctx, source, "source", private, symmetric, config); err != nil {
		t.Fatal(err)
	}
	if err := WriteEnvelope(ctx, source, "source", []byte("recovery envelope")); err != nil {
		t.Fatal(err)
	}
	if err := kvtx.RunTransaction(ctx, true, func(ctx context.Context) (kvtx.Tx, error) { return source.NewTransaction(ctx, true) }, func(ctx context.Context, tx kvtx.Tx) error {
		return tx.Set(ctx, MakeKey("source", SuffixSetupDone), []byte{1})
	}); err != nil {
		t.Fatal(err)
	}
	fault := kvtest.NewFaultStore(destination, kvtest.FaultBeforeCommit)
	if err := CopyCredential(ctx, source, fault, "source", "destination", key, [32]byte{3}); err != nil {
		t.Fatal(err)
	}
	if fault.Opened() != 2 || fault.DelegatedCommits() != 1 {
		t.Fatal("credential copy did not replay as one transaction")
	}
	gotPrivate, gotSymmetric, gotConfig, err := ReadPINLockFiles(ctx, destination, "destination")
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnlockPIN(gotPrivate, gotSymmetric, gotConfig, []byte("123456"))
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("PIN did not survive: %v", err)
	}
	if !bytes.Equal(gotPrivate, private) || !bytes.Equal(gotSymmetric, symmetric) || !gotConfig.EqualVT(config) {
		t.Fatal("PIN credential was replaced")
	}
	if err := CopyCredential(ctx, source, destination, "source", "destination", key, [32]byte{3}); err != nil {
		t.Fatal(err)
	}
	read, err := destination.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()
	for suffix, expected := range map[string][]byte{string(SuffixEnvelope): []byte("recovery envelope"), string(SuffixSetupDone): {1}} {
		got, found, err := read.Get(ctx, MakeKey("destination", []byte(suffix)))
		if err != nil || !found || !bytes.Equal(got, expected) {
			t.Fatalf("missing retained field %s: %v", suffix, err)
		}
	}
}

func TestCopyCredentialRewrapsAutoUnlockAndRejectsConflictingKey(t *testing.T) {
	ctx := t.Context()
	source, destination := store_kvtx_inmem.NewStore(), store_kvtx_inmem.NewStore()
	key, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := keypem.MarshalPrivKeyPem(key)
	if err != nil {
		t.Fatal(err)
	}
	original, err := EncryptAutoUnlock([32]byte{1}, plain)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteAutoUnlock(ctx, source, "source", original); err != nil {
		t.Fatal(err)
	}
	if err := CopyCredential(ctx, source, destination, "source", "destination", key, [32]byte{2}); err != nil {
		t.Fatal(err)
	}
	encrypted, _, err := ReadAutoUnlockKey(ctx, destination, "destination")
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecryptAutoUnlock([32]byte{2}, encrypted)
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("destination key did not unlock: %v", err)
	}
	retained, _, err := ReadAutoUnlockKey(ctx, source, "source")
	if err != nil || !bytes.Equal(retained, original) {
		t.Fatal("source credential changed")
	}
	other, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := CopyCredential(ctx, source, destination, "source", "destination", other, [32]byte{2}); err == nil {
		t.Fatal("a different destination key was overwritten")
	}
	retained, _, err = ReadAutoUnlockKey(ctx, destination, "destination")
	if err != nil || !bytes.Equal(retained, encrypted) {
		t.Fatal("conflicting attempt changed the destination")
	}
}
