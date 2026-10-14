package kernel

// File is host code's reference to an open file passed over a Unix socket.
// It keeps the file open until Close, like a guest descriptor.
type File struct {
	// k is the kernel whose lock guards the file.
	k *Kernel
	// of is the open file, or nil once closed.
	of *OpenFile
}

// Close drops the reference.
func (f *File) Close() error {
	f.k.mu.Lock()
	defer f.k.mu.Unlock()
	if f.of != nil {
		f.of.unref()
		f.of = nil
	}
	return nil
}
