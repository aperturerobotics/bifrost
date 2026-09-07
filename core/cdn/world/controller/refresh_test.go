package cdn_world_controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aperturerobotics/starpc/srpc"
	"github.com/s4wave/spacewave/bldr/util/packedmsg"
	"github.com/s4wave/spacewave/core/cdn"
	"github.com/s4wave/spacewave/core/sobject"
	sobject_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/sirupsen/logrus"
)

// TestRefreshRPCRefetchesMountedWorld proves invalidation reaches the mounted
// CDN owner without replacing its engine or fetching unrelated Spaces.
func TestRefreshRPCRefetchesMountedWorld(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	state, err := (&sobject_engine.InnerState{HeadRef: &bucket.ObjectRef{}}).MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	inner, err := (&sobject.SORootInner{Seqno: 1, StateData: state}).MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	pointer, err := (&cdn.CdnRootPointer{SpaceId: "release-space", Root: &sobject.SORoot{Inner: inner, InnerSeqno: 1}}).MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	refetched := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) > 1 {
			select {
			case refetched <- struct{}{}:
			default:
			}
		}
		_, _ = w.Write([]byte(packedmsg.EncodePackedMessage(pointer)))
	}))
	defer server.Close()
	ctrl := NewController(logrus.NewEntry(logrus.New()), nil, NewConfig("release", "release-space", server.URL))
	done := make(chan error, 1)
	go func() { done <- ctrl.Execute(ctx) }()
	defer func() { cancel(); <-done }()
	engine, err := ctrl.GetWorldEngine(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client := NewSRPCWorldRefreshClientWithServiceID(srpc.NewClient(srpc.NewServerPipe(srpc.NewServer(ctrl))), WorldRefreshServiceID("release"))
	response, err := client.Refresh(ctx, &RefreshRequest{SpaceId: "unrelated"})
	if err != nil || response.GetAccepted() || requests.Load() != 1 {
		t.Fatalf("unrelated refresh: %v %v requests=%d", response, err, requests.Load())
	}
	response, err = client.Refresh(ctx, &RefreshRequest{SpaceId: "release-space"})
	if err != nil || !response.GetAccepted() {
		t.Fatalf("refresh: %v %v", response, err)
	}
	select {
	case <-refetched:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	current, err := ctrl.GetWorldEngine(ctx)
	if err != nil || current != engine {
		t.Fatal("refresh replaced the mounted engine")
	}
}
