package main

import (
	"net/http"
	"os"
	"strings"
)

type corsConfig struct {
	allowAll       bool
	allowedOrigins map[string]struct{}
}

func loadCORSConfig() corsConfig {
	value := strings.TrimSpace(os.Getenv("DRB99_CORS_ORIGINS"))
	if value == "" || value == "*" {
		return corsConfig{allowAll: true}
	}

	allowed := make(map[string]struct{})
	for _, origin := range strings.Split(value, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return corsConfig{allowedOrigins: allowed}
}

func corsMiddleware(config corsConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" && (config.allowAll || config.isAllowed(origin)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c corsConfig) isAllowed(origin string) bool {
	_, ok := c.allowedOrigins[origin]
	return ok
}
