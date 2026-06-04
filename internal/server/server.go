package server

import (
	"net/http"

	"github.com/h3yng/drb99/generator"
	gh "github.com/h3yng/drb99/github"
	"github.com/h3yng/drb99/handler"
	"github.com/h3yng/drb99/internal/config"
	"github.com/h3yng/drb99/internal/middleware"
	service "github.com/h3yng/drb99/services"
)

func New() http.Handler {
	config.LoadDotEnv()

	ghClient := gh.NewClient()
	gen := generator.New()
	svc := service.New(ghClient, gen)
	h := handler.New(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	return middleware.CORSMiddleware(middleware.LoadCORSConfig(), mux)
}