package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Router methods

// ServeHTTP implements http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find matching route
	for _, route := range r.routes {
		if route.Match(req.Method, req.URL.Path) {
			// Apply route-specific middleware
			handler := route.Handler
			for i := len(route.Middlewares) - 1; i >= 0; i-- {
				handler = route.Middlewares[i](handler)
			}

			// Apply global middleware
			for i := len(r.middleware) - 1; i >= 0; i-- {
				handler = r.middleware[i](handler)
			}

			handler(w, req)
			return
		}
	}

	// No route found, check if there's a NotFound handler
	for _, route := range r.routes {
		if route.Path == "*" && route.Method == "*" {
			handler := route.Handler
			for i := len(route.Middlewares) - 1; i >= 0; i-- {
				handler = route.Middlewares[i](handler)
			}
			handler(w, req)
			return
		}
	}

	// Default 404
	http.NotFound(w, req)
}

// Match checks if route matches request
func (r *Route) Match(method, path string) bool {
	// Check method
	if r.Method != method && r.Method != "*" {
		return false
	}

	// Check path
	return r.MatchPath(path)
}

// MatchPath checks if path matches route pattern
func (r *Route) MatchPath(path string) bool {
	if r.Path == path {
		return true
	}

	// Handle wildcards
	if strings.Contains(r.Path, "{") && strings.Contains(r.Path, "}") {
		return r.matchPattern(path)
	}

	// Handle trailing slashes
	if strings.HasSuffix(r.Path, "/") && path == strings.TrimSuffix(r.Path, "/") {
		return true
	}

	if strings.HasSuffix(path, "/") && r.Path == strings.TrimSuffix(path, "/") {
		return true
	}

	return false
}

// matchPattern matches path with pattern (e.g., /users/{id})
func (r *Route) matchPattern(path string) bool {
	patternParts := strings.Split(r.Path, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, patternPart := range patternParts {
		pathPart := pathParts[i]

		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			// This is a parameter, accept any value
			continue
		}

		if patternPart != pathPart {
			return false
		}
	}

	return true
}

// GetParams extracts parameters from path
func (r *Route) GetParams(path string) map[string]string {
	params := make(map[string]string)

	if !strings.Contains(r.Path, "{") {
		return params
	}

	patternParts := strings.Split(r.Path, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return params
	}

	for i, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			paramName := strings.TrimSuffix(strings.TrimPrefix(patternPart, "{"), "}")
			params[paramName] = pathParts[i]
		}
	}

	return params
}

// AddRoute adds a route to router
func (r *Router) AddRoute(method, path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	route := Route{
		Method:      strings.ToUpper(method),
		Path:        path,
		Handler:     handler,
		Middlewares: middleware,
	}

	r.routes = append(r.routes, route)
	log.Printf("Route registered: %-6s %s", method, path)
}

// Get adds GET route
func (r *Router) Get(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.AddRoute("GET", path, handler, middleware...)
}

// Post adds POST route
func (r *Router) Post(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.AddRoute("POST", path, handler, middleware...)
}

// Put adds PUT route
func (r *Router) Put(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.AddRoute("PUT", path, handler, middleware...)
}

// Delete adds DELETE route
func (r *Router) Delete(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.AddRoute("DELETE", path, handler, middleware...)
}

// Patch adds PATCH route
func (r *Router) Patch(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.AddRoute("PATCH", path, handler, middleware...)
}

// Options adds OPTIONS route
func (r *Router) Options(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.AddRoute("OPTIONS", path, handler, middleware...)
}

// Head adds HEAD route
func (r *Router) Head(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.AddRoute("HEAD", path, handler, middleware...)
}

// Any adds route for any method
func (r *Router) Any(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.AddRoute("*", path, handler, middleware...)
}

// NotFound sets 404 handler
func (r *Router) NotFound(handler http.HandlerFunc) {
	r.AddRoute("*", "*", handler)
}

// Use adds global middleware
func (r *Router) Use(middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.middleware = append(r.middleware, middleware...)
}

// AddMiddlewareToRoute adds middleware to specific route
func (r *Router) AddMiddlewareToRoute(method, path string, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, route := range r.routes {
		if route.Method == method && route.Path == path {
			r.routes[i].Middlewares = append(r.routes[i].Middlewares, middleware...)
			return
		}
	}
}

// GetRoutes returns all registered routes
func (r *Router) GetRoutes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]Route, len(r.routes))
	copy(routes, r.routes)
	return routes
}

// PrintRoutes prints all registered routes
func (r *Router) PrintRoutes() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, route := range r.routes {
		fmt.Printf("  %-6s %s\n", route.Method, route.Path)
	}
}

// Group creates a route group with prefix
func (r *Router) Group(prefix string, middleware ...func(http.HandlerFunc) http.HandlerFunc) *RouteGroup {
	return &RouteGroup{
		router:     r,
		prefix:     prefix,
		middleware: middleware,
	}
}

// RouteGroup represents a group of routes with common prefix/middleware
type RouteGroup struct {
	router     *Router
	prefix     string
	middleware []func(http.HandlerFunc) http.HandlerFunc
}

// Get adds GET route to group
func (g *RouteGroup) Get(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	fullPath := g.prefix + path
	allMiddleware := append(g.middleware, middleware...)
	g.router.Get(fullPath, handler, allMiddleware...)
}

// Post adds POST route to group
func (g *RouteGroup) Post(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	fullPath := g.prefix + path
	allMiddleware := append(g.middleware, middleware...)
	g.router.Post(fullPath, handler, allMiddleware...)
}

// Put adds PUT route to group
func (g *RouteGroup) Put(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	fullPath := g.prefix + path
	allMiddleware := append(g.middleware, middleware...)
	g.router.Put(fullPath, handler, allMiddleware...)
}

// Delete adds DELETE route to group
func (g *RouteGroup) Delete(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	fullPath := g.prefix + path
	allMiddleware := append(g.middleware, middleware...)
	g.router.Delete(fullPath, handler, allMiddleware...)
}

// Patch adds PATCH route to group
func (g *RouteGroup) Patch(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	fullPath := g.prefix + path
	allMiddleware := append(g.middleware, middleware...)
	g.router.Patch(fullPath, handler, allMiddleware...)
}

// Any adds route for any method to group
func (g *RouteGroup) Any(path string, handler http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	fullPath := g.prefix + path
	allMiddleware := append(g.middleware, middleware...)
	g.router.Any(fullPath, handler, allMiddleware...)
}

// Use adds middleware to group
func (g *RouteGroup) Use(middleware ...func(http.HandlerFunc) http.HandlerFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// SubGroup creates a subgroup
func (g *RouteGroup) SubGroup(prefix string, middleware ...func(http.HandlerFunc) http.HandlerFunc) *RouteGroup {
	return &RouteGroup{
		router:     g.router,
		prefix:     g.prefix + prefix,
		middleware: append(g.middleware, middleware...),
	}
}

// Helper function to get path parameters from request context
func GetParam(r *http.Request, name string) string {
	// In a real implementation, you'd extract params from context
	// For this simple router, we parse from URL path
	parts := strings.Split(r.URL.Path, "/")

	// Simple extraction for patterns like /users/{id}
	for i, part := range parts {
		if part == "{"+name+"}" {
			if i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}

	return ""
}
