package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bashnko/drb99/generator"
	gh "github.com/bashnko/drb99/github"
	"github.com/bashnko/drb99/handler"
	service "github.com/bashnko/drb99/services"
)

func main() {
	addr := os.Getenv("DRB99")
	if addr == "" {
		addr = ":8088"
	}

	ghClient := gh.NewClient()
	gen := generator.New()
	svc := service.New(ghClient, gen)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("drb99 listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
