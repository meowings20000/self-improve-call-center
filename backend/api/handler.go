package api

import (
	"backend/learning"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	engine *learning.Engine
	mux    *http.ServeMux
}

func NewHandler() *Handler {
	h := &Handler{engine: learning.NewEngine(), mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /api/health", h.health)
	h.mux.HandleFunc("POST /api/sessions", h.startSession)
	h.mux.HandleFunc("GET /api/sessions/{id}/exercise", h.nextExercise)
	h.mux.HandleFunc("GET /api/sessions/{id}/progress", h.progress)
	h.mux.HandleFunc("POST /api/sessions/{id}/answers", h.checkAnswer)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	allowedOrigin := origin == "http://localhost:3000" || origin == "http://localhost:3001"
	if origin != "" && !allowedOrigin {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}
	if allowedOrigin {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	h.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "LingoLift API"})
}

func (h *Handler) startSession(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusCreated, h.engine.NewSession())
}

func (h *Handler) nextExercise(w http.ResponseWriter, r *http.Request) {
	exercise, err := h.engine.Next(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, exercise)
}

func (h *Handler) progress(w http.ResponseWriter, r *http.Request) {
	progress, err := h.engine.Progress(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

func (h *Handler) checkAnswer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ExerciseID string `json:"exerciseId"`
		Answer     string `json:"answer"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(input.ExerciseID) == "" {
		writeError(w, http.StatusBadRequest, "exerciseId is required")
		return
	}
	grade, err := h.engine.Check(r.PathValue("id"), input.ExerciseID, input.Answer)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "session not found" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, grade)
}
