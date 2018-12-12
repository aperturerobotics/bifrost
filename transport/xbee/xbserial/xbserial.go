package xbserial

import (
	"sync"

	"github.com/pauleyj/gobee"
	"github.com/pauleyj/gobee/api"
	"github.com/pauleyj/gobee/api/rx"
	"github.com/pauleyj/gobee/api/tx"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

// XBeeSerial communicates with a xbee over a serial line.
// Implements net.PacketConn
type XBeeSerial struct {
	le   *logrus.Entry
	port *serial.Port
	xbee *gobee.XBee

	// txMtx guards the transmission process
	txMtx      sync.Mutex
	txStatusCh chan *rx.TXStatus

	cmdMtx     sync.Mutex
	ncmdID     byte
	atHandlers map[byte]func(frame *rx.AT)

	dataHandler func(data []byte, addr XBeeAddr)
}

// NewXBeeSerial builds a xbee serial wrapper from a serial port.
func NewXBeeSerial(
	le *logrus.Entry,
	serialPort *serial.Port,
) *XBeeSerial {
	s := &XBeeSerial{le: le, port: serialPort}
	s.atHandlers = make(map[byte]func(frame *rx.AT))
	s.txStatusCh = make(chan *rx.TXStatus, 1)
	str := &serialTxRx{xbs: s}
	s.xbee = gobee.NewWithEscapeMode(str, str, api.EscapeModeActive)
	return s
}

// SetDataHandler sets the data handling function.
func (x *XBeeSerial) SetDataHandler(handler func(data []byte, addr XBeeAddr)) {
	x.dataHandler = handler
}

// ReadPump is a goroutine that reads data from the xbee.
func (x *XBeeSerial) ReadPump() error {
	buf := make([]byte, 1200)
	for {
		n, err := x.port.Read(buf)
		if err != nil {
			return err
		}

		for i := 0; i < n; i++ {
			if err := x.xbee.RX(buf[i]); err != nil {
				return err
			}
		}
	}
}

// WriteFrame attempts to write a frame to the xbee.
func (x *XBeeSerial) WriteFrame(frame tx.Frame) (int, error) {
	return x.xbee.TX(frame)
}

// frameHandler handles incoming frames.
func (x *XBeeSerial) frameHandler(frame rx.Frame) {
	switch f := frame.(type) {
	case *rx.AT:
		x.handleATFrame(f)
	case *rx.ZB:
		x.handleZBFrame(f)
	case *rx.ZBExplicit:
		x.handleZBFrameExplicit(f)
	case *rx.TXStatus:
		x.handleTxStatusFrame(f)
	default:
		x.le.Debugf("unhandled packet: %#v", frame)
	}
}
