package kernel

// OpenFile represents an open file description. Descriptors made by dup,
// exec and fork share one OpenFile, and so share its offset and flags.
type OpenFile struct {
	node   *VNode
	path   string // absolute path at time of open
	offset int64
	flags  uint32

	// stderr routes console writes to the kernel's stderr stream.
	stderr bool

	// pipe is the pipe end the file reads or writes, if any. Each
	// descriptor table entry holds one reference to it.
	pipe *pipeEnd

	// data is a synthesized file's content, generated at open.
	data []byte

	// Directory reading state for getdents64.
	dirNames []string
}

// ref takes a descriptor table entry's reference to the file's pipe end.
func (of *OpenFile) ref() {
	if of.pipe != nil {
		of.pipe.ref()
	}
}

// unref drops a descriptor table entry's reference to the file's pipe end.
func (of *OpenFile) unref() {
	if of.pipe != nil {
		of.pipe.unref()
	}
}
