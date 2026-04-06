package main

import (
	"net/http"
	"os"
)

func main() {
	addr := os.Getenv("DRB99_ADDR")
	if addr == "" {
		addr = ":8088"
	}

	mux := http.NewServeMux()
	http.ListenAndServe(addr, mux)
}
