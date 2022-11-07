package bifrost_http

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/controllerbus/util/ccontainer"
	"github.com/aperturerobotics/controllerbus/util/refcount"
)

// HTTPHandlerController resolves LookupHTTPHandler with a http.Handler.
type HTTPHandlerController struct {
	// info is the controller info
	info *controller.Info
	// handleCtr is the refcount handle to the UnixFS
	handleCtr *ccontainer.CContainer[*http.Handler]
	// errCtr contains any error building FSHandle
	errCtr *ccontainer.CContainer[*error]
	// rc is the refcount container
	rc *refcount.RefCount[*http.Handler]
	// pathPrefixes is the list of URL path prefixes to match.
	// ignores if empty
	pathPrefixes []string
	// stripPathPrefix removes the first matched pathPrefix from the URL.
	stripPathPrefix bool
	// pathRe is a regex to match URL paths.
	// ignores if empty
	pathRe *regexp.Regexp
}

// NewHTTPHandlerController constructs a new controller.
//
// Responds if a URL matches either pathPrefixes OR pathRe.
// pathPrefixes and pathRe can be empty.
// if stripPathPrefix is set, removes the pathPrefix from the URL.
func NewHTTPHandlerController(
	info *controller.Info,
	resolver HTTPHandlerBuilder,
	pathPrefixes []string,
	stripPathPrefix bool,
	pathRe *regexp.Regexp,
) *HTTPHandlerController {
	h := &HTTPHandlerController{
		info:            info,
		handleCtr:       ccontainer.NewCContainer[*http.Handler](nil),
		errCtr:          ccontainer.NewCContainer[*error](nil),
		pathPrefixes:    pathPrefixes,
		stripPathPrefix: stripPathPrefix,
		pathRe:          pathRe,
	}
	h.rc = refcount.NewRefCount(nil, h.handleCtr, h.errCtr, resolver)
	return h
}

// GetControllerInfo returns information about the controller.
func (c *HTTPHandlerController) GetControllerInfo() *controller.Info {
	return c.info.Clone()
}

// Execute executes the controller.
func (c *HTTPHandlerController) Execute(ctx context.Context) error {
	c.rc.SetContext(ctx)
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
func (c *HTTPHandlerController) HandleDirective(
	ctx context.Context,
	inst directive.Instance,
) ([]directive.Resolver, error) {
	switch d := inst.GetDirective().(type) {
	case LookupHTTPHandler:
		rurl, err := url.Parse(d.LookupHTTPHandlerURL())
		if err != nil {
			return nil, err
		}
		rpath := rurl.Path
		// if we have no filters, match all.
		matched := len(c.pathPrefixes) == 0 && c.pathRe == nil
		var stripPrefix string
		if !matched && len(c.pathPrefixes) != 0 {
			for _, prefix := range c.pathPrefixes {
				if strings.HasPrefix(rpath, prefix) {
					matched = true
					stripPrefix = prefix
					break
				}
			}
		}
		if !matched && c.pathRe != nil {
			matched = c.pathRe.MatchString(rpath)
		}
		if !matched {
			return nil, nil
		}
		builder := c.HTTPHandleBuilder
		if c.stripPathPrefix && len(stripPrefix) != 0 {
			builder = func(ctx context.Context) (*http.Handler, func(), error) {
				handlerPtr, handlerRel, err := c.HTTPHandleBuilder(ctx)
				if err != nil || handlerPtr == nil || *handlerPtr == nil {
					return handlerPtr, handlerRel, err
				}
				handler := *handlerPtr
				handler = http.StripPrefix(stripPrefix, handler)
				return &handler, handlerRel, nil
			}
		}
		return directive.R(NewLookupHTTPHandlerBuilderResolver(inst, builder), nil)
	}
	return nil, nil
}

// HTTPHandleBuilder builds a handle by adding a reference to the refcounter.
func (c *HTTPHandlerController) HTTPHandleBuilder(ctx context.Context) (*http.Handler, func(), error) {
	valCh := make(chan *http.Handler, 1)
	errCh := make(chan error, 1)
	ref := c.rc.AddRef(func(val *http.Handler, err error) {
		if err != nil {
			select {
			case errCh <- err:
			default:
			}
		} else if val != nil {
			select {
			case valCh <- val:
			default:
			}
		}
	})
	select {
	case err := <-errCh:
		ref.Release()
		return nil, nil, err
	case val := <-valCh:
		return val, ref.Release, nil
	}
}

// Close releases any resources used by the controller.
func (c *HTTPHandlerController) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*HTTPHandlerController)(nil))
