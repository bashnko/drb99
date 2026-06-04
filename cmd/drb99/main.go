package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/h3yng/drb99/internal/server"
)

func main() {
	addr := listenAddrFromEnv("PORT")
	if addr == "" {
		addr = listenAddrFromEnv("DRB99")
	}
	if addr == "" {
		addr = ":8088"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           server.New(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("drb99 listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func listenAddrFromEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		return ""
	}
	if value[0] == ':' {
		return value
	}
	return ":" + value
}
