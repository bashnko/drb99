package handler

import (
	"context"
	"encoding/json"
	"net/http"

	service "github.com/bashnko/drb99/services"
)

type Handler struct {
	svc *service.Service
}

const (
	apiVersionPrefix = "/api/v1"
	legacyGenerate   = "/generate"
	legactyHealth    = "/health"
	generate         = apiVersionPrefix + "/generate"
	health           = apiVersionPrefix + "/health"
	prefill          = apiVersionPrefix + "/prefill"
)

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(generate, h.handleGenerate)
	mux.HandleFunc(health, h.handleHealth)
	mux.HandleFunc(legacyGenerate, h.handleGenerate)
	mux.HandleFunc(legactyHealth, h.handleHealth)
	mux.HandleFunc(prefill, h.handlePrefill)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req service.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	resp, err := h.svc.Generate(r.Context(), req)
	if err != nil {
		code := http.StatusBadRequest
		if err == context.Canceled || err == context.DeadlineExceeded {
			code = http.StatusRequestTimeout
		}
		writeError(w, code, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)

}

func (h *Handler) handlePrefill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req service.PrefillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	resp, err := h.svc.Prefill(r.Context(), req)
	if err != nil {
		code := http.StatusBadRequest
		if err == context.Canceled || err == context.DeadlineExceeded {
			code = http.StatusRequestTimeout
		}
		writeError(w, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
