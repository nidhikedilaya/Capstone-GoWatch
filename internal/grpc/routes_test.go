package grpc

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {

	// --------------------------------------------------------
	// Create a lightweight REST server instance for testing
	// --------------------------------------------------------
	// We don't need a database here because HelloWorldHandler
	// does not use the DB. So we initialize RestServer with nil.
	s := &RestServer{}

	// --------------------------------------------------------
	// Spin up an in-memory test HTTP server
	// --------------------------------------------------------
	// httptest.NewServer starts a real HTTP server, but only
	// accessible inside this test. It automatically handles
	// routes and lifecycle.
	server := httptest.NewServer(http.HandlerFunc(s.HelloWorldHandler))
	defer server.Close()

	// --------------------------------------------------------
	// Perform a GET request to the test server
	// --------------------------------------------------------
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()

	// --------------------------------------------------------
	// Assert: HTTP status code should be 200 OK
	// --------------------------------------------------------
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}

	// Expected JSON response
	expected := "{\"message\":\"Hello World\"}"

	// --------------------------------------------------------
	// Read response body into memory
	// --------------------------------------------------------
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}

	// --------------------------------------------------------
	// Assert: body must match expected JSON
	// --------------------------------------------------------
	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}
