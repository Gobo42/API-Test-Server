package main

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync/atomic"
	"unsafe"
)

// configPtr holds the active *Config via atomic swap so SIGHUP reloads are race-free.
var configPtr unsafe.Pointer

func getConfig() *Config {
	return (*Config)(atomic.LoadPointer(&configPtr))
}

func setConfig(cfg *Config) {
	atomic.StorePointer(&configPtr, unsafe.Pointer(cfg))
}

func handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()

	dump := buildDump(r, body)
	fmt.Print(dump)

	cfg := getConfig()
	route := findRoute(cfg.Routes, r.Method, r.URL.Path)

	if route != nil {
		writeResponse(w, route.Status, route.Headers, route.Body)
		return
	}

	// Use default config entry
	def := cfg.Default
	writeResponse(w, def.Status, def.Headers, def.Body)
}

func buildDump(r *http.Request, body []byte) string {
	var sb strings.Builder
	sb.WriteString("--- Request ---\n")
	fmt.Fprintf(&sb, "Method: %s\n", r.Method)
	fmt.Fprintf(&sb, "Path:   %s\n", r.URL.Path)
	if r.URL.RawQuery != "" {
		fmt.Fprintf(&sb, "Query:  %s\n", r.URL.RawQuery)
	}
	sb.WriteString("\nHeaders:\n")

	// Sort headers for deterministic output
	names := make([]string, 0, len(r.Header))
	for name := range r.Header {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		for _, val := range r.Header[name] {
			fmt.Fprintf(&sb, "  %s: %s\n", name, val)
		}
	}

	if len(body) > 0 {
		sb.WriteString("\nBody:\n")
		sb.Write(body)
		sb.WriteByte('\n')
	}
	sb.WriteString("\n")
	return sb.String()
}

func writeResponse(w http.ResponseWriter, status int, headers map[string]string, body string) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}
