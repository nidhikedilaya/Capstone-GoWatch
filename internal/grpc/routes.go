package grpc

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (s *RestServer) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", s.statusHandler)
	mux.HandleFunc("/alerts/history", s.alertHistoryHandler)
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/", s.HelloWorldHandler)

	return s.corsMiddleware(mux)
}

// -------------------- CORS Middleware --------------------

func (s *RestServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// -------------------- Handlers --------------------

func (s *RestServer) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *RestServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *RestServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	var list []CurrentState
	state.Range(func(key, val any) bool {
		list = append(list, val.(CurrentState))
		return true
	})

	json.NewEncoder(w).Encode(list)
}

func (s *RestServer) alertHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// context with timeout for DB call
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	alerts, err := s.db.GetAlertHistory(ctx)
	if err != nil {
		http.Error(w, "failed to load history", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(alerts)
}
