// Package rue provides a simple router based on Mat Ryer's way package.
// It extends way by adding all form parameters to the context.
// Multiple values for the same form parameter are concatenate with
// the semi-colon separator.
// The request host prefix is also added to the context with the _host key.
package rue

import (
	"context"
	"net/http"
	"strings"
)

// rueContextKey is the context key type for storing
// parameters (from path or form) in context.Context.
// Also used to store host name with "_host" special key.
type rueContextKey string

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
	method := strings.ToLower(req.Method)
	segs := r.pathSegments(req.URL.Path)
	for _, route := range r.routes {
		if route.method != method && route.method != "*" {
			continue
		}
		if ctx, ok := route.match(req.Context(), r, segs); ok {
			ctx = route.form(ctx, req) // add form parameters to context
			ctx = route.host(ctx, req) // add host prefix to context
			route.handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
	}
	r.NotFound.ServeHTTP(w, req)
}

// Param gets the path or form parameter from the specified context.
// Returns an empty string if the parameter was not found.
func Param(ctx context.Context, param string) string {
	vStr, ok := ctx.Value(rueContextKey(param)).(string)
	if !ok {
		return ""
	}
	return vStr
}

type route struct {
	method  string
	segs    []string
	handler http.Handler
	prefix  bool
}

func (r *route) match(ctx context.Context, router *Router, segs []string) (context.Context, bool) {
	if len(segs) > len(r.segs) && !r.prefix {
		return nil, false
	}
	for i, seg := range r.segs {
		if i > len(segs)-1 {
			return nil, false
		}
		isParam := false
		if strings.HasPrefix(seg, ":") {
			isParam = true
			seg = strings.TrimPrefix(seg, ":")
		}
		if !isParam { // verbatim check
			if strings.HasSuffix(seg, "...") {
				if strings.HasPrefix(segs[i], seg[:len(seg)-3]) {
					return ctx, true
				}
			}
			if seg != segs[i] {
				return nil, false
			}
		}
		if isParam {
			ctx = context.WithValue(ctx, rueContextKey(seg), segs[i])
		}
	}
	return ctx, true
}

// Add all form parameters to the context.
// Multiple values are concatenate with semi-colon separator.
func (r *route) form(ctx context.Context, req *http.Request) context.Context {
	err := req.ParseForm()
	if err != nil {
		return ctx
	}
	var value string
	for k, values := range req.Form {
		if len(values) == 0 {
			continue
		}
		for i, v := range values {
			if i > 0 {
				value = value + ";" + v
			} else {
				value = v
			}
		}
		ctx = context.WithValue(ctx, rueContextKey(k), value)
	}
	return ctx
}

func (r *route) host(ctx context.Context, req *http.Request) context.Context {
	host := strings.Split(req.Host, ".")[0]
	ctx = context.WithValue(ctx, rueContextKey("_host"), host)
	return ctx
}
