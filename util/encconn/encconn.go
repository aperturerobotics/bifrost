package encconn

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// EncConn wraps net.Conn with box encryption
//
// NOTE: currently not working (as of August 2021)
type EncConn struct {
	net.Conn
	sharedSecret [32]byte

	rxBuf            bytes.Buffer
	rxN, txN         int
	rxNonce, txNonce [24]byte
}

// NewEncConn wraps a net.Conn with a shared secret.
func NewEncConn(c net.Conn, sharedSecret [32]byte) *EncConn {
	ec := &EncConn{
		Conn:         c,
		sharedSecret: sharedSecret,
	}

	xorBy := sharedSecret[int(sharedSecret[0])%(len(sharedSecret)-1)]
	ec.initNonce(ec.rxNonce[:], xorBy)
	ec.initNonce(ec.txNonce[:], xorBy)
	return ec
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *EncConn) Read(b []byte) (int, error) {
	if c.rxBuf.Len() != 0 {
		return c.rxBuf.Read(b)
	}

	var buf [1500]byte
	n, err := c.Conn.Read(buf[:])
	if err != nil {
		return 0, err
	}

	c.incNonce(&c.rxN, c.rxNonce[:])
	d, ok := box.OpenAfterPrecomputation(nil, buf[:n], &c.rxNonce, &c.sharedSecret)
	if !ok {
		return 0, errors.New("secretbox decryption failed")
	}

	_, _ = c.rxBuf.Write(d)
	return c.rxBuf.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *EncConn) Write(b []byte) (n int, err error) {
	c.incNonce(&c.txN, c.txNonce[:])
	encBuf := box.SealAfterPrecomputation(nil, b, &c.txNonce, &c.sharedSecret)
	if wn, err := c.Conn.Write(encBuf); err != nil {
		return wn, err
	}

	return len(b), nil
}

// initNonce generates the initial nonce
func (c *EncConn) initNonce(buf []byte, xorBy byte) {
	for i := 0; i < len(buf)-4; i++ {
		buf[i] = c.sharedSecret[i] ^ xorBy
	}
}

// incNonce increments a nonce
func (c *EncConn) incNonce(counter *int, buf []byte) {
	v := *counter
	v++
	*counter = v
	binary.LittleEndian.PutUint32(buf[len(buf)-5:], uint32(v))
}
