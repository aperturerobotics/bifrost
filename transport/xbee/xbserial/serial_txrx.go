package xbserial

import (
	// "encoding/hex"
	"github.com/pauleyj/gobee"
	"github.com/pauleyj/gobee/api/rx"
)

// serialTxRx implements the gobee interfaces with a serial port.
type serialTxRx struct {
	xbs *XBeeSerial
}

// Transmit writes data to the xbee.
func (x *serialTxRx) Transmit(data []byte) (int, error) {
	// x.xbs.le.Debugf("transmitting: %s", hex.Dump(data))
	return x.xbs.port.Write(data)
}

// Receive indicates a received frame.
func (x *serialTxRx) Receive(rxf rx.Frame) error {
	// x.xbs.le.Debugf("handling: %v", rxf)
	x.xbs.frameHandler(rxf)
	return nil
}

// _ is a type assertion
var _ gobee.XBeeTransmitter = ((*serialTxRx)(nil))

// _ is a type assertion
var _ gobee.XBeeReceiver = ((*serialTxRx)(nil))
