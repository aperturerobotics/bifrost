package block

// StorePart is one labeled child of a block store built from other stores.
type StorePart struct {
	// Label names the child's role in its parent, such as "lower" or "upper".
	Label string
	// Store is the child store.
	Store StoreOps
}

// StoreParts is implemented by a block store built from other stores, so an
// inspector can open it into its children.
type StoreParts interface {
	// GetStoreParts returns the child stores.
	GetStoreParts() []StorePart
}

// StoreHolder is implemented by a controller that holds a block store it built
// itself, outside the bus.
type StoreHolder interface {
	// GetHeldStore returns the held store, or nil before it is built.
	GetHeldStore() StoreOps
}
