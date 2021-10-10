// Package rue provides a simple router based on Mat Ryer's way package.
// It differs from _way_ by adding the path parameters to the request values
// instead of using the request context. Multiple values for the same form
// parameter are concatenate with the semi-colon separator. The host prefix
// is also added to the request values with the _host special key.
package rue

import (
	"net/http"
	"strings"
)

// Router routes HTTP requests.
type Router struct {
	routes []*route
	// NotFound is the http.Handler to call when no routes
	// match. By default uses http.NotFoundHandler().
	NotFound http.Handler
}

// NewRouter makes a new Router.
func NewRouter() *Router {
	return &Router{
		NotFound: http.NotFoundHandler(),
	}
}

func (r *Router) pathSegments(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

// Handle adds a handler with the specified method and pattern.
// Method can be any HTTP method string or "*" to match all methods.
// Pattern can contain path segments such as: /item/:id which is
// accessible via the Param function.
// If pattern ends with trailing /, it acts as a prefix.
func (r *Router) Handle(method, pattern string, handler http.Handler) {
	route := &route{
		method:  strings.ToLower(method),
		segs:    r.pathSegments(pattern),
		handler: handler,
		prefix:  strings.HasSuffix(pattern, "/") || strings.HasSuffix(pattern, "..."),
	}
	r.routes = append(r.routes, route)
}

// HandleFunc is the http.HandlerFunc alternative to http.Handle.
func (r *Router) HandleFunc(method, pattern string, fn http.HandlerFunc) {
	r.Handle(method, pattern, fn)
}

// ServeHTTP routes the incoming http.Request based on method and path
// extracting path parameters as it goes.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// prepare the req values : it's important to call it before adding values
	req.ParseForm()

	// store the host prefix
	host := strings.Split(req.Host, ".")[0]
	req.Form.Add("_host", host)

	method := strings.ToLower(req.Method)
	segs := r.pathSegments(req.URL.Path)
	for _, route := range r.routes {
		if route.method != method && route.method != "*" {
			continue
		}
		if ok := route.match(req, r, segs); ok {
			route.handler.ServeHTTP(w, req)
			return
		}
	}
	r.NotFound.ServeHTTP(w, req)
}

// Param returns the parameter's value from the request or the empty
// string if the parameter was not found.
// The parameter can be a path parameter or a form value.
// Returns a semi-colon separated value for same key multi-values.
func Param(req *http.Request, param string) string {
	values := req.Form[param]
	if len(values) == 0 {
		return ""
	}
	var value string
	for i, v := range values {
		if i > 0 {
			value = value + ";" + v
		} else {
			value = v
		}
	}
	return value
}

type route struct {
	method  string
	segs    []string
	handler http.Handler
	prefix  bool
}

func (r *route) match(req *http.Request, router *Router, segs []string) bool {
	if len(segs) > len(r.segs) && !r.prefix {
		return false
	}

	// parse the path for matching and storing path parameters if any
	for i, seg := range r.segs {
		if i > len(segs)-1 {
			return false
		}
		if strings.HasPrefix(seg, ":") {
			seg = strings.TrimPrefix(seg, ":")
			req.Form.Add(seg, segs[i])
			continue
		}
		// verbatim check
		if strings.HasSuffix(seg, "...") {
			if strings.HasPrefix(segs[i], seg[:len(seg)-3]) {
				return true
			}
		}
		if seg != segs[i] {
			return false
		}
	}
	return true
}
