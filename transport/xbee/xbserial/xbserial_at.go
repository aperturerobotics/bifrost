package xbserial

import (
	"context"
	"encoding/binary"

	"github.com/pauleyj/gobee/api/rx"
	"github.com/pauleyj/gobee/api/tx"
)

// handleATFrame handles an incoming AT frame.
func (x *XBeeSerial) handleATFrame(f *rx.AT) {
	id := f.ID()
	cmd := f.Command()
	data := f.Data()
	status := f.Status()

	go func() {
		x.cmdMtx.Lock()
		h, hOk := x.atHandlers[id]
		if hOk {
			delete(x.atHandlers, id)
		}
		x.cmdMtx.Unlock()
		if hOk {
			h(f)
		} else {
			x.le.Debugf(
				"unhandled AT frame: id(%v) status(%v) cmd(%v) data(%v)",
				id,
				status,
				cmd,
				data,
			)
		}
	}()
}

type ATCommand [2]byte

var (
	ATSH = ATCommand{0x53, 0x48}
	ATSL = ATCommand{0x53, 0x4C}
)

// execATCall executes an AT call and waits for a response.
func (x *XBeeSerial) execATCall(ctx context.Context, cmd ATCommand, param *byte) (*rx.AT, error) {
	frameCh := make(chan *rx.AT, 1)
	x.cmdMtx.Lock()
	x.ncmdID++
	id := x.ncmdID
	x.atHandlers[id] = func(frame *rx.AT) {
		frameCh <- frame
	}
	x.cmdMtx.Unlock()

	unregHandler := func() {
		x.cmdMtx.Lock()
		delete(x.atHandlers, id)
		x.cmdMtx.Unlock()
	}

	// x.le.Debugf("writing frame: id(%v) command(%v) param(%v)", id, cmd, param)
	_, err := x.WriteFrame(
		tx.NewATBuilder().
			ID(id).
			Command(cmd).
			Parameter(param).
			Build(),
	)
	if err != nil {
		unregHandler()
		return nil, err
	}

	select {
	case f := <-frameCh:
		return f, nil
	case <-ctx.Done():
		unregHandler()
		return nil, ctx.Err()
	}
}

// ReadLocalAddress reads the local address.
func (x *XBeeSerial) ReadLocalAddress(ctx context.Context) (XBeeAddr, error) {
	atSH, err := x.execATCall(ctx, ATSH, nil)
	if err != nil {
		return 0, err
	}

	atSL, err := x.execATCall(ctx, ATSL, nil)
	if err != nil {
		return 0, err
	}

	data := make([]byte, 8)
	copy(data[:4], atSH.Data())
	copy(data[4:], atSL.Data())
	return XBeeAddr(binary.BigEndian.Uint64(data)), nil
}
