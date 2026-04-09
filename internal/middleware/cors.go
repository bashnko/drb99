package middleware

import (
	"net/http"
	"os"
	"strings"
)

type CORSConfig struct {
	AllowAll       bool
	AllowedOrigins map[string]struct{}
}

func LoadCORSConfig() CORSConfig {
	value := strings.TrimSpace(os.Getenv("DRB99_CORS_ORIGINS"))
	if value == "" || value == "*" {
		return CORSConfig{AllowAll: true}
	}

	allowed := make(map[string]struct{})
	for _, origin := range strings.Split(value, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return CORSConfig{AllowedOrigins: allowed}
}

func CORSMiddleware(config CORSConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" && (config.AllowAll || config.isAllowed(origin)) {
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

func (c CORSConfig) isAllowed(origin string) bool {
	_, ok := c.AllowedOrigins[origin]
	return ok
}
