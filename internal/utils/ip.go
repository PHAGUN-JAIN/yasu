package utils

import (
	"net"
	"net/http"
	"strings"
)

func RealIP(h http.Header) string {
	// X-Forwarded-For can have multiple IPs: "client, proxy1, proxy2"
	// first one is the real client
	if xff := h.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(h.Get("RemoteAddr"))
	if err != nil {
		return h.Get("RemoteAddr")
	}
	return ip
}
