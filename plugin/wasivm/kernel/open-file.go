package kernel

// OpenFile represents an open file description. Descriptors made by dup,
// exec and vfork share one OpenFile, and so share its offset and flags.
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

// hold keeps the file's pipe end open while a vfork child changes the
// descriptor table the parent will get back.
func (of *OpenFile) hold() {
	if of.pipe != nil {
		of.pipe.hold()
	}
}

// restoreHeld turns a hold back into the parent's table reference.
func (of *OpenFile) restoreHeld() {
	if of.pipe != nil {
		of.pipe.restoreHeld()
	}
}
