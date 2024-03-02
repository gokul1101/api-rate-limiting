package main

import (
	"net"
	"net/http"
	"time"
)

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {

	go cleanupExpiredRateLimiters()

	return func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.URL.Path
		ip, port, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			sendResponse(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		connectionString := ip + ":" + port + endpoint
		queueMutex.Lock()
		if len(requestQueue) >= maxQueueSize {
			queueMutex.Unlock()
			sendResponse(w, "Server busy, try again later", http.StatusServiceUnavailable)
			return
		}
		requestQueue = append(requestQueue, Request{Endpoint: connectionString, Time: time.Now()})
		queueMutex.Unlock()

		clientsMutex.Lock()
		client, ok := clients[ip]
		clientsMutex.Unlock()

		if !ok {
			sendResponse(w, "Client not found", http.StatusBadRequest)
			return
		}

		clientsMutex.Lock()
		limiter, ok := client.RateLimiters[endpoint]
		clientsMutex.Unlock()
		if !ok {
			sendResponse(w, "Endpoint not found for the client", http.StatusBadRequest)
			return
		}

		if !limiter.Allow() {
			sendResponse(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			sendResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		} else if token == "Zocket" {
			sendResponse(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
