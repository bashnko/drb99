package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bashnko/drb99/generator"
	gh "github.com/bashnko/drb99/github"
	"github.com/bashnko/drb99/handler"
	"github.com/bashnko/drb99/internal/config"
	"github.com/bashnko/drb99/internal/middleware"
	service "github.com/bashnko/drb99/services"
)

func main() {
	config.LoadDotEnv()

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
	corsConfig := middleware.LoadCORSConfig()

	srv := &http.Server{
		Addr:              addr,
		Handler:           middleware.CORSMiddleware(corsConfig, mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("drb99 listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
