package bifrost_http

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// HTTPHandlerController resolves LookupHTTPHandler with a http.Handler.
type HTTPHandlerController struct {
	// info is the controller info
	info *controller.Info
	// handler is the http handler
	handler http.Handler
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
	handler http.Handler,
	pathPrefixes []string,
	stripPathPrefix bool,
	pathRe *regexp.Regexp,
) *HTTPHandlerController {
	return &HTTPHandlerController{
		info:            info,
		handler:         handler,
		pathPrefixes:    pathPrefixes,
		stripPathPrefix: stripPathPrefix,
		pathRe:          pathRe,
	}
}

// GetControllerInfo returns information about the controller.
func (c *HTTPHandlerController) GetControllerInfo() *controller.Info {
	return c.info.Clone()
}

// Execute executes the controller.
// Returning nil ends execution.
func (c *HTTPHandlerController) Execute(rctx context.Context) (rerr error) {
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
		handler := c.handler
		if c.stripPathPrefix && len(stripPrefix) != 0 {
			handler = http.StripPrefix(stripPrefix, handler)
		}
		return directive.R(NewLookupHTTPHandlerResolver(handler), nil)
	}
	return nil, nil
}

// Close releases any resources used by the controller.
func (c *HTTPHandlerController) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*HTTPHandlerController)(nil))
