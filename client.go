package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func createClient(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		sendResponse(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var requestData struct {
		Name      string         `json:"name"`
		IP        string         `json:"ip"`
		Resources map[string]int `json:"resources"`
	}

	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		sendResponse(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	if requestData.Name == "" || requestData.IP == "" || len(requestData.Resources) == 0 {
		sendResponse(w, "Name, IP, and Resources are required", http.StatusBadRequest)
		return
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if _, ok := clients[requestData.IP]; ok {
		sendResponse(w, "Client with IP already exists", http.StatusBadRequest)
		return
	}

	rateLimiters := make(map[string]*rate.Limiter)
	for endpoint, limit := range requestData.Resources {
		rateLimiters[endpoint] = rate.NewLimiter(rate.Every(time.Second), limit)
	}

	newClient := &Client{
		Name:         requestData.Name,
		IP:           requestData.IP,
		ResourceInfo: requestData.Resources,
		RateLimiters: rateLimiters,
	}

	clients[newClient.IP] = newClient
	sendResponse(w, "Client created successfully", http.StatusCreated)
}

func getClients(w http.ResponseWriter, r *http.Request) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	res := make(map[string]interface{})
	res["status"] = http.StatusOK
	res["message"] = "Clients retrieved successfully"
	clientsList := make([]Client, 0)
	for _, v := range clients {
		clientsList = append(clientsList, *v)
	}
	res["clients"] = clientsList

	byteData, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		sendResponse(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(byteData)
}

func updateClient(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		sendResponse(w, "Client IP is required", http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		sendResponse(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var requestData struct {
		Name      string         `json:"name"`
		Resources map[string]int `json:"resources"`
	}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		sendResponse(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	if requestData.Name == "" || len(requestData.Resources) == 0 {
		sendResponse(w, "Name or Resources is required", http.StatusBadRequest)
		return
	}

	clientsMutex.Lock()
	client, ok := clients[ip]

	if !ok {
		sendResponse(w, "Client not found", http.StatusNotFound)
		return
	}

	// Update client information
	client.Name = requestData.Name
	client.ResourceInfo = requestData.Resources

	for endpoint, limit := range requestData.Resources {
		client.RateLimiters[endpoint] = rate.NewLimiter(rate.Every(time.Second), limit)
	}

	clients[ip] = client
	clientsMutex.Unlock()

	sendResponse(w, "Client updated successfully", http.StatusOK)
}

func deleteClient(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		sendResponse(w, "Client IP is required", http.StatusBadRequest)
		return
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if _, ok := clients[ip]; !ok {
		sendResponse(w, "Client not found", http.StatusNotFound)
		return
	}
	delete(clients, ip)

	sendResponse(w, "Client deleted successfully", http.StatusOK)
}

func seedClient(w http.ResponseWriter, r *http.Request) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		sendResponse(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	resourceInfo := make(map[string]int)
	resourceInfo["/api/resource1"] = 10
	resourceInfo["/api/resource2"] = 5
	resourceInfo["/api/resource3"] = 3

	rateLimiters := make(map[string]*rate.Limiter)
	for endpoint, limit := range resourceInfo {
		rateLimiters[endpoint] = rate.NewLimiter(rate.Every(time.Second), limit)
	}
	clients[ip] = &Client{
		Name: "Zocket",
		IP: ip,
		ResourceInfo: resourceInfo,
		RateLimiters: rateLimiters,
	}
	sendResponse(w, "Default Client details added successfully", http.StatusOK)
}

func sendResponse(w http.ResponseWriter, message string, status int) {
	res := &Response{Status: status, Message: message}
	jsonResponse, err := json.Marshal(res)
	if err != nil {
		sendResponse(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}
