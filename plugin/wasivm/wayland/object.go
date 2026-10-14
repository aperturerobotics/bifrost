package wayland

// object is a protocol object a client holds.
type object interface {
	// request handles the request with opcode op, reading its arguments
	// from m.
	request(op uint16, m *message) error
}

// requestFunc is an object whose requests one function handles.
type requestFunc func(op uint16, m *message) error

// request calls f.
func (f requestFunc) request(op uint16, m *message) error {
	return f(op, m)
}

// noRequests is an object without requests, such as a callback.
var noRequests = requestFunc(func(uint16, *message) error { return nil })

// _ is a type assertion.
var _ object = requestFunc(nil)
