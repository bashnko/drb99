package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/h3yng/drb99/internal/server"
)

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8088"
	}

	srv := &http.Server{
		Addr:              ":" + addr,
		Handler:           server.New(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("drb99 listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
