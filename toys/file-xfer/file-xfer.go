package main

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/aperturerobotics/bifrost/core"
	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	nctr "github.com/aperturerobotics/bifrost/peer/controller"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/transport/common/pconn"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"runtime/pprof"
)

var testProtocolID = protocol.ID("/x/test")

func main() {
	if err := doIt(true); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}

func doIt(doProf bool) error {
	if doProf {
		cpuProf, err := os.OpenFile("cpu-profile.prof", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer cpuProf.Close()

		pprof.StartCPUProfile(cpuProf)
		defer pprof.StopCPUProfile()
	}

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)
	ctx := context.Background()

	// Construct node 1.
	n1Priv, n1Pub, err := keypem.GeneratePrivKey()
	_ = n1Pub
	if err != nil {
		return err
	}
	n1Pem, err := keypem.MarshalPrivKeyPem(n1Priv)
	if err != nil {
		return err
	}
	p1, err := peer.IDFromPrivateKey(n1Priv)
	if err != nil {
		return err
	}
	b1, _, err := core.NewCoreBus(ctx, le.WithField("node", "1"))
	if err != nil {
		return err
	}
	le.Debugf("node 1: %s", p1.Pretty())

	// Construct node 2.
	n2Priv, n2Pub, err := keypem.GeneratePrivKey()
	_ = n2Pub
	if err != nil {
		return err
	}
	n2Pem, err := keypem.MarshalPrivKeyPem(n2Priv)
	if err != nil {
		return err
	}
	p2, err := peer.IDFromPrivateKey(n2Priv)
	if err != nil {
		return err
	}
	b2, _, err := core.NewCoreBus(ctx, le.WithField("node", "2"))
	if err != nil {
		return err
	}
	le.Debugf("node 2: %s", p2.Pretty())

	_, n1Ref, err := bus.ExecOneOff(
		ctx,
		b1,
		resolver.NewLoadControllerWithConfigSingleton(
			&nctr.Config{
				PrivKey: string(n1Pem),
			},
		),
		nil,
	)
	if err != nil {
		return err
	}
	defer n1Ref.Release()

	_, n2Ref, err := bus.ExecOneOff(
		ctx,
		b2,
		resolver.NewLoadControllerWithConfigSingleton(
			&nctr.Config{
				PrivKey: string(n2Pem),
			},
		),
		nil,
	)
	if err != nil {
		return err
	}
	defer n2Ref.Release()

	pconnOpts := &pconn.Opts{
		KcpMode: pconn.KCPMode_KCPMode_FAST3,
		// BlockCrypt:   pconn.BlockCrypt_BlockCrypt_NONE,
		BlockCrypt:   pconn.BlockCrypt_BlockCrypt_AES256,
		DataShards:   10,
		ParityShards: 3,
	}

	// Construct transports.
	_, n1UdpRef, err := b1.AddDirective(
		resolver.NewLoadControllerWithConfigSingleton(&udptpt.Config{
			ListenAddr: "127.0.0.1:9823",
			PacketOpts: pconnOpts,
		}),
		bus.NewCallbackHandler(func(val directive.Value) {
			le.Infof("UDP listening on: %s", "127.0.0.1:9823")
		}, nil, nil),
	)
	if err != nil {
		return errors.Wrap(err, "listen on udp n1")
	}
	defer n1UdpRef.Release()

	_, n2UdpRef, err := b2.AddDirective(
		resolver.NewLoadControllerWithConfigSingleton(&udptpt.Config{
			ListenAddr: "127.0.0.1:9824",
			DialAddrs:  []string{"127.0.0.1:9823"},
			PacketOpts: pconnOpts,
		}),
		bus.NewCallbackHandler(func(val directive.Value) {
			le.Infof("UDP listening on: %s", "127.0.0.1:9824")
		}, nil, nil),
	)
	if err != nil {
		return errors.Wrap(err, "listen on udp n2")
	}
	defer n2UdpRef.Release()

	// Open stream
	doneCh := make(chan struct{})
	so := stream.OpenOpts{Encrypted: true, Reliable: true}
	// open stream p1 -> p2 with protocol pid
	_, s1_2Ref, err := b1.AddDirective(
		link.NewOpenStreamWithPeer(testProtocolID, p1, p2, 0, so),
		bus.NewCallbackHandler(func(v directive.Value) {
			mstrm := v.(link.MountedStream)
			le.Debug("opened stream p1 -> p2")
			defer mstrm.GetStream().Close()
			f, err := os.OpenFile("file-p1.out", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				le.WithError(err).Warn("unable to open output file")
				return
			}

			t1 := time.Now()
			buf := make([]byte, 1500)
			nw, err := io.CopyBuffer(f, mstrm.GetStream(), buf)
			t2 := time.Now()
			f.Close()
			le.
				WithError(err).
				WithField("bytes-written", nw).
				WithField("time", t2.Sub(t1).String()).
				Info("copybuffer finished")
			close(doneCh)
		}, nil, nil),
	)
	if err != nil {
		return err
	}
	defer s1_2Ref.Release()

	// Handle stream
	s1h := &StreamHandler{le: le}
	if err := b2.AddHandler(s1h); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
	case _, _ = <-doneCh:
	}
	return nil
}

// StreamHandler handles HandleMountedStream directives with a file.
type StreamHandler struct {
	le *logrus.Entry
}

// HandleDirective asks if the handler can resolve the directive.
func (h *StreamHandler) HandleDirective(
	ctx context.Context,
	inst directive.Instance,
) (directive.Resolver, error) {
	dir := inst.GetDirective()
	switch d := dir.(type) {
	case link.HandleMountedStream:
		return h.resolveHandleMountedStream(ctx, inst, d)
	}

	return nil, nil
}

func (h *StreamHandler) resolveHandleMountedStream(
	ctx context.Context,
	inst directive.Instance,
	dir link.HandleMountedStream,
) (directive.Resolver, error) {
	h.le.Debugf("handling protocol: %s", string(dir.HandleMountedStreamProtocolID()))
	if dir.HandleMountedStreamProtocolID() != testProtocolID {
		return nil, nil
	}

	return &HandleMountedStreamResolver{h: h}, nil
}

// HandleMountedStream handles an incoming mounted stream.
// Any returned error indicates the stream should be closed.
// This function should return as soon as possible, and start
// additional goroutines to manage the lifecycle of the stream.
func (h *StreamHandler) HandleMountedStream(
	ctx context.Context,
	strm link.MountedStream,
) error {
	f, err := os.Open("file.in")
	if err != nil {
		h.le.WithError(err).Warn("cannot open file.in")
		return err
	}
	go func() {
		defer f.Close()
		h.le.Debug("starting transfer of file.in")
		buf := make([]byte, 100000)
		nw, err := io.CopyBuffer(strm.GetStream(), f, buf)
		if err != nil {
			h.le.WithError(err).Warn("unable to transfer file.in")
		} else {
			h.le.Debugf("transferred file.in bytes %d", nw)
		}
		strm.GetStream().Close()
	}()
	return nil
}

// _ is a type assertion
var _ directive.Handler = ((*StreamHandler)(nil))

// _ is a type assertion
var _ link.MountedStreamHandler = ((*StreamHandler)(nil))

type HandleMountedStreamResolver struct {
	h *StreamHandler
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (r *HandleMountedStreamResolver) Resolve(
	ctx context.Context,
	handler directive.ResolverHandler,
) error {
	_, _ = handler.AddValue(r.h)
	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*HandleMountedStreamResolver)(nil))
