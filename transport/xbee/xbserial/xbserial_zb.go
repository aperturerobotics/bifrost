package xbserial

import (
	"github.com/pauleyj/gobee/api/rx"
)

// handleZBFrame handles an incoming zigbee frame
func (x *XBeeSerial) handleZBFrame(f *rx.ZB) {
	// source endpoint: E8
	// destination endpoint: E8
	// cluster id: 00 11
	// profile id: c1 05
	/*
		x.le.Debugf(
			"incoming data: from %v srce(%x) dste(%x) cluster(%x) prof(%x) data: %v",
			f.Addr64(),
			0xE8,
			0xE8,
			0x0011,
			0xC105,
			f.Data(),
		)
	*/

	remoteAddr := f.Addr64()
	data := f.Data()
	x.handleIncomingData(XBeeAddr(remoteAddr), data)
}

// handleZBFrameExplicit handles an incoming zigbee frame
func (x *XBeeSerial) handleZBFrameExplicit(f *rx.ZBExplicit) {
	remoteAddr := f.Addr64()
	data := f.Data()
	x.handleIncomingData(XBeeAddr(remoteAddr), data)
	/*
		x.le.Debugf(
			"incoming data: srce(%x) dste(%x) cluster(%x) prof(%x) data: %v",
			f.SrcEP(),
			f.DstEP(),
			f.ClusterID(),
			f.ProfileID(),
			f.Data(),
		)
	*/
}

// handleIncomingData handles incoming data by queuing it for processing.
func (x *XBeeSerial) handleIncomingData(remoteAddr XBeeAddr, data []byte) {
	if x.dataHandler != nil {
		x.dataHandler(data, remoteAddr)
	} else {
		x.le.Debugf("data handler nil: %v %v", remoteAddr, data)
	}
}
