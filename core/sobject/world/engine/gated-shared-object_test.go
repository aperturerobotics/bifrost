//go:build !js

package sobject_world_engine_test

import (
	"context"
	"sync"
	"testing"

	provider_local "github.com/s4wave/spacewave/core/provider/local"
	"github.com/s4wave/spacewave/core/sobject"
	sobject_world_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	"github.com/s4wave/spacewave/db/block"
	transform_all "github.com/s4wave/spacewave/db/block/transform/all"
	world_mock "github.com/s4wave/spacewave/db/world/mock"
)

// gatedSharedObject is a local SharedObject that can hold the physical
// commit of the next publication a World write admits, then apply or fail it.
type gatedSharedObject struct {
	*provider_local.SharedObject
	// ctrl is the World engine controller serving the SharedObject.
	ctrl *sobject_world_engine.Controller
	// stopServe stops the engine serve started.
	stopServe func()

	// mtx guards next.
	mtx sync.Mutex
	// next wraps the state head of the next admitted publication, or is nil.
	next func(head *block.AtomicHeadUpdate) *block.AtomicHeadUpdate
}

// startGatedWorld starts a local provider on s4db storage under dir and a
// World engine serving the SharedObject soRef through a gatedSharedObject,
// creating the SharedObject when soRef is nil. Test cleanup stops them.
func startGatedWorld(t *testing.T, dir string, soRef *sobject.SharedObjectRef) (*localWorld, *gatedSharedObject) {
	// Mount the SharedObject and wrap it.
	t.Helper()
	tb, acc, soRef := openLocalAccount(t, dir, soRef)
	so, soMountRef, err := sobject.ExMountSharedObject(t.Context(), tb.Bus, soRef, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(soMountRef.Release)
	local, ok := so.(*provider_local.SharedObject)
	if !ok {
		t.Fatalf("SharedObject is %T, not a local SharedObject", so)
	}
	ctrl, err := sobject_world_engine.NewController(tb.Logger, tb.Bus, sobject_world_engine.NewConfig("gated-world-engine", soRef), transform_all.BuildFactorySet())
	if err != nil {
		t.Fatal(err)
	}
	ctrl.SetStaticLookupOp(world_mock.LookupMockOp)
	gated := &gatedSharedObject{SharedObject: local, ctrl: ctrl}

	// Serve its World with the mock operations until cleanup.
	gated.serve(t)
	eng, err := ctrl.GetWorldEngine(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return newLocalWorld(t, tb, acc, eng, soRef), gated
}

// serve starts serving the World of s with its controller until stop or test
// cleanup. A controller that served before replaces its engine.
func (s *gatedSharedObject) serve(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	exited := make(chan error, 1)
	go func() {
		exited <- sobject_world_engine.ExecuteWorld(ctx, s.ctrl, s)
	}()
	var once sync.Once
	s.stopServe = func() {
		once.Do(func() {
			cancel()
			<-exited
		})
	}
	t.Cleanup(s.stopServe)
}

// stop ends the engine serve started, as losing read access does.
func (s *gatedSharedObject) stop() {
	s.stopServe()
}

// hold holds the physical commit of the next publication a World write
// admits. applying receives once the commit reaches the state head. release
// lets it go on: it fails with failure when set, and applies otherwise.
func (s *gatedSharedObject) hold(failure error) (applying <-chan struct{}, release func()) {
	// Gate the state head's replacement on release.
	reached := make(chan struct{}, 1)
	released := make(chan struct{})
	s.mtx.Lock()
	s.next = func(head *block.AtomicHeadUpdate) *block.AtomicHeadUpdate {
		return &block.AtomicHeadUpdate{
			ObjectStoreID: head.ObjectStoreID,
			Key:           head.Key,
			Replace: func(ctx context.Context, current []byte, found bool) ([]byte, error) {
				select {
				case reached <- struct{}{}:
				default:
				}
				select {
				case <-released:
				case <-ctx.Done():
					return nil, context.Cause(ctx)
				}
				if failure != nil {
					return nil, failure
				}
				return head.Replace(ctx, current, found)
			},
		}
	}
	s.mtx.Unlock()
	return reached, func() { close(released) }
}

// QueueOperation queues op, holding its publication when hold armed it.
func (s *gatedSharedObject) QueueOperation(ctx context.Context, op []byte) (string, error) {
	// Take the gate for an admitted write.
	publish := sobject.GetPublishState(ctx)
	if publish == nil {
		return s.SharedObject.QueueOperation(ctx, op)
	}
	s.mtx.Lock()
	wrap := s.next
	s.next = nil
	s.mtx.Unlock()

	// Publish through it.
	if wrap != nil {
		ctx = sobject.WithPublishState(ctx, func(ctx context.Context, head *block.AtomicHeadUpdate) (*block.PublicationReceipt, error) {
			return publish(ctx, wrap(head))
		})
	}
	return s.SharedObject.QueueOperation(ctx, op)
}

// _ is a type assertion
var _ sobject.SharedObject = (*gatedSharedObject)(nil)
