package main

import "strings"

// matchURI returns true if the uri pattern matches the request path.
// '*' matches exactly one non-empty path segment.
func matchURI(pattern, path string) bool {
	// Strip query string from path if present
	if i := strings.IndexByte(path, '?'); i != -1 {
		path = path[:i]
	}

	patSegs := strings.Split(strings.Trim(pattern, "/"), "/")
	reqSegs := strings.Split(strings.Trim(path, "/"), "/")

	if len(patSegs) != len(reqSegs) {
		return false
	}

	for i, seg := range patSegs {
		if seg == "*" {
			if reqSegs[i] == "" {
				return false
			}
			continue
		}
		if seg != reqSegs[i] {
			return false
		}
	}
	return true
}

// findRoute returns the first route whose uri and method match the request,
// or nil if none match.
func findRoute(routes []RouteConfig, method, path string) *RouteConfig {
	for i := range routes {
		r := &routes[i]
		if r.Method != "*" && !strings.EqualFold(r.Method, method) {
			continue
		}
		if matchURI(r.URI, path) {
			return r
		}
	}
	return nil
}
