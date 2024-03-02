package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

type Request struct {
	Endpoint string
	Time     time.Time
}

type Client struct {
	Name         string
	IP           string
	ResourceInfo map[string]int
	RateLimiters map[string]*rate.Limiter
}

const maxQueueSize = 50

var (
	requestQueue []Request
	queueMutex   sync.Mutex
	clients      map[string]*Client
	clientsMutex sync.Mutex
)

func handler1(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Resource1 accessed"))
}
func handler2(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Resource2 accessed"))
}
func handler3(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Resource3 accessed"))
}

func main() {
	clients = make(map[string]*Client)

	r := mux.NewRouter()

	// Client CRUD APIs - Private
	r.HandleFunc("/clients", authMiddleware(getClients)).Methods("GET")
	r.HandleFunc("/client", authMiddleware(createClient)).Methods("POST")
	r.HandleFunc("/client", authMiddleware(updateClient)).Methods("PATCH")
	r.HandleFunc("/client", authMiddleware(deleteClient)).Methods("DELETE")
	r.HandleFunc("/seed-client", authMiddleware(seedClient)).Methods("POST")
	
	// Rate limited APIs
	r.HandleFunc("/api/resource1", rateLimitMiddleware(handler1))
	r.HandleFunc("/api/resource2", rateLimitMiddleware(handler2))
	r.HandleFunc("/api/resource3", rateLimitMiddleware(handler3))

	http.Handle("/", r)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error in listening on port 8080")
	}
}

func cleanupExpiredRateLimiters() {
	cleanupTicker := time.NewTicker(3 * time.Second)
	defer cleanupTicker.Stop()

	for range cleanupTicker.C {
		processExpiredRequests()
	}
}

func processExpiredRequests() {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	currentTime := time.Now()
	i := 0
	for i < len(requestQueue) {
		req := requestQueue[i]
		if currentTime.Sub(req.Time) > 10*time.Second {
			requestQueue[i] = requestQueue[len(requestQueue)-1]
			requestQueue = requestQueue[:len(requestQueue)-1]
		} else {
			i++
		}
	}
}
