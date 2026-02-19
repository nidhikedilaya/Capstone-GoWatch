package grpc

import (
	"encoding/json"
	"log"
	"net/http"
)

// /////////////////////////////////////////////////////////////////////////////
// RegisterRoutes
// ---------------------------------------------------------------------------
// This sets up all REST endpoints and attaches them to a ServeMux.
// Finally, it wraps everything with CORS middleware so frontend clients can
// call your API without cross-origin issues.
// /////////////////////////////////////////////////////////////////////////////
func (s *RestServer) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// -----------------------------
	// Register REST API endpoints
	// -----------------------------
	mux.HandleFunc("/status", s.statusHandler)               // Returns current metric states in memory
	mux.HandleFunc("/alerts/history", s.alertHistoryHandler) // Returns alert history from MySQL
	mux.HandleFunc("/health", s.healthHandler)               // DB health check
	mux.HandleFunc("/", s.HelloWorldHandler)                 // Default endpoint

	// Add CORS middleware → allows safe browser access
	return s.corsMiddleware(mux)
}

// /////////////////////////////////////////////////////////////////////////////
// CORS Middleware
// ---------------------------------------------------------------------------
// Adds headers that allow browsers to call your backend without blocking
// cross-origin requests. This is important if your frontend runs on a
// different port (e.g., React on :3000, API on :8080).
// /////////////////////////////////////////////////////////////////////////////
func (s *RestServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Allow all origins (for development)
		// Replace "*" with your frontend domain in production.
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Allowed HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// Allowed request headers
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Authorization, Content-Type, X-CSRF-Token")

		// Whether cookies and credentials are allowed
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		// Preflight request handler (browser sends OPTIONS request first)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Forward the request to the actual handler
		next.ServeHTTP(w, r)
	})
}

// /////////////////////////////////////////////////////////////////////////////
// HelloWorldHandler
// ---------------------------------------------------------------------------
// A simple functional endpoint used for testing.
// Returns: { "message": "Hello World" }
// /////////////////////////////////////////////////////////////////////////////
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

// /////////////////////////////////////////////////////////////////////////////
// healthHandler
// ---------------------------------------------------------------------------
// Calls your MySQL service Health() method.
// Example output: { "database": "UP" }
// /////////////////////////////////////////////////////////////////////////////
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

// /////////////////////////////////////////////////////////////////////////////
// statusHandler
// ---------------------------------------------------------------------------
// Returns the current *in-memory* state of all active services/agents.
// This reads from your global sync.Map named `state` (from worker logic).
// /////////////////////////////////////////////////////////////////////////////
func (s *RestServer) statusHandler(w http.ResponseWriter, r *http.Request) {

	var list []CurrentState

	// Iterate through sync.Map
	state.Range(func(key, val any) bool {
		list = append(list, val.(CurrentState))
		return true
	})

	// Encode array to JSON
	json.NewEncoder(w).Encode(list)
}

// /////////////////////////////////////////////////////////////////////////////
// alertHistoryHandler
// ---------------------------------------------------------------------------
// Fetches all saved alerts from MySQL.
// Returns the complete stored alert history.
// /////////////////////////////////////////////////////////////////////////////
func (s *RestServer) alertHistoryHandler(w http.ResponseWriter, r *http.Request) {

	alerts, err := s.db.GetAlertHistory()
	if err != nil {
		http.Error(w, "failed to load history", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(alerts)
}
