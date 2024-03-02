package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestRateLimiterAPI(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/client/create", createClient).Methods("POST")
	r.HandleFunc("/api/resource1", rateLimitMiddleware(handler1)).Methods("GET")
	r.HandleFunc("/api/resource2", rateLimitMiddleware(handler2)).Methods("GET")
	r.HandleFunc("/api/resource3", rateLimitMiddleware(handler3)).Methods("GET")
	ts := httptest.NewServer(r)
	defer ts.Close()
	clients = make(map[string]*Client)

	// Create a client with rate limits for multiple resources
	clientData := map[string]interface{}{
		"name": "TestClient",
		"ip":   "127.0.0.1",
		"resources": map[string]int{
			"/api/resource1": 15,
			"/api/resource2": 10,
			"/api/resource3": 13,
		},
	}

	jsonData, _ := json.Marshal(clientData)
	resp, err := http.Post(ts.URL+"/client/create", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Errorf("Error creating client: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Error creating client, status code: %d", resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	// Make requests to the rate-limited APIs and track the response codes
	responseCodes := make([]int, 50) // Increased number of requests for better testing
	for i := 0; i < 50; i++ {
		var resourcePath string
		switch i % 3 {
		case 0:
			resourcePath = "/api/resource1"
		case 1:
			resourcePath = "/api/resource2"
		case 2:
			resourcePath = "/api/resource3"
		}
		resp, err := http.Get(ts.URL + resourcePath)
		if err != nil {
			t.Errorf("Error making request: %v", err)
			continue
		}
		defer resp.Body.Close()
		responseCodes[i] = resp.StatusCode
	}

	// Check rate limiting for each resource
	expectedLimits := map[string]int{
		"/api/resource1": 5,
		"/api/resource2": 10,
		"/api/resource3": 3,
	}
	// Check rate limiting for each resource
	for resource, limit := range expectedLimits {
		tooManyRequestsCount := 0
		successCount := 0

		// Count only the responses related to the current resource
		for i, code := range responseCodes {
			var resourcePath string
			switch i % 3 {
			case 0:
				resourcePath = "/api/resource1"
			case 1:
				resourcePath = "/api/resource2"
			case 2:
				resourcePath = "/api/resource3"
			}
			if resourcePath == resource {
				if code == http.StatusTooManyRequests {
					tooManyRequestsCount++
				} else if code == http.StatusOK {
					successCount++
				}
			}
		}

		// Perform assertions based on counts for the current resource
		if tooManyRequestsCount > limit {
			t.Errorf("Rate limiting for %s failed. Expected at most %d Too Many Requests responses, got %d", resource, limit, tooManyRequestsCount)
		}
		if successCount < limit {
			t.Errorf("Expected at least %d successful responses for %s, got %d", limit, resource, successCount)
		}
	}

}
