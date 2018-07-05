// Package directive declares the Directive type. The Directive is an
// instruction to a node controller of desired connectivity. Directives are
// de-duplicated in the controller. Each directive is added with a listener,
// which receives events updating the status of the directive, and yielded
// connectivity (Links, etc.).
package directive

// Handler handles new directives.
type Handler interface {
	// AddDirective handles a new directive.
	AddDirective(Directive)
}

// Directive implements a request for connectivity.
type Directive interface{}
