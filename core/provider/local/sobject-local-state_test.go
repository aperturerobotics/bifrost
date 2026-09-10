package provider_local

import (
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/util/ulid"
	"github.com/s4wave/spacewave/core/sobject"
)

// TestSharedObjectLocalStateIsolation reopens independent stores that use the
// same key, including two stores belonging to one SharedObject.
func TestSharedObjectLocalStateIsolation(t *testing.T) {
	// Mount two independent SharedObjects in one account's backing store.
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	_, _, account, _, release := setupProviderAndSessionInternal(ctx, t)
	defer release()
	var objects []sobject.SharedObject
	for range 2 {
		ref, err := account.CreateSharedObject(ctx, ulid.NewULID(), &sobject.SharedObjectMeta{BodyType: "space"}, "", "")
		if err != nil {
			t.Fatal(err)
		}
		object, releaseObject, err := account.MountSharedObject(ctx, ref, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer releaseObject()
		objects = append(objects, object)
	}

	// Persist distinct values at an identical key in all four local stores.
	for objectIndex, object := range objects {
		for storeIndex, storeID := range []string{"copy", "rejected-candidates"} {
			store, releaseStore, err := object.AccessLocalStateStore(ctx, storeID, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer releaseStore()
			tx, err := store.NewTransaction(ctx, true)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Discard()
			value := []byte{byte(objectIndex), byte(storeIndex)}
			if err := tx.Set(ctx, []byte("progress"), value); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			tx.Discard()
			releaseStore()
		}
	}

	// Reopening a store must recover only its own SharedObject and consumer state.
	for objectIndex, object := range objects {
		for storeIndex, storeID := range []string{"copy", "rejected-candidates"} {
			store, releaseStore, err := object.AccessLocalStateStore(ctx, storeID, nil)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := store.NewTransaction(ctx, false)
			if err != nil {
				t.Fatal(err)
			}
			value, found, err := tx.Get(ctx, []byte("progress"))
			tx.Discard()
			releaseStore()
			if err != nil || !found || len(value) != 2 || value[0] != byte(objectIndex) || value[1] != byte(storeIndex) {
				t.Fatalf("object %d store %s lost its own state: value=%v found=%v err=%v", objectIndex, storeID, value, found, err)
			}
		}
	}
}
