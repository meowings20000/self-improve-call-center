package main

import (
	"backend/openllm"
	"backend/sales"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OpenLLMConfig stores the OpenLLM configuration
type OpenLLMConfig struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

var openllmClient *openllm.Client
var openllmMutex sync.Mutex

func sendError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Code:    code,
		Message: message,
	})
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message":   "Backend is running!",
			"timestamp": time.Now().Format(time.RFC3339),
			"status":    "healthy",
		})
	})

	// GET/POST /api/openllm/config
	mux.HandleFunc("/api/openllm/config", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodGet {
			openllmMutex.Lock()
			defer openllmMutex.Unlock()
			config := OpenLLMConfig{}
			if openllmClient != nil {
				config.BaseURL = openllmClient.BaseURL
				config.APIKey = openllmClient.APIKey
				config.Model = openllmClient.Model
			}
			json.NewEncoder(w).Encode(config)
			return
		}

		if r.Method == http.MethodPost {
			var config OpenLLMConfig
			if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
				sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body: "+err.Error())
				return
			}

			if config.BaseURL == "" {
				sendError(w, http.StatusBadRequest, "MISSING_PARAM", "baseUrl is required")
				return
			}

			openllmMutex.Lock()
			defer openllmMutex.Unlock()
			openllmClient = openllm.NewClient(config.BaseURL, config.APIKey, config.Model)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": "OpenLLM configuration updated",
			})
			return
		}

		sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	})

	// POST /api/openllm/test - test OpenLLM connection
	mux.HandleFunc("/api/openllm/test", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
			return
		}

		openllmMutex.Lock()
		client := openllmClient
		openllmMutex.Unlock()

		if client == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"connected": false,
				"error":     "OpenLLM not configured",
				"code":      "NOT_CONFIGURED",
			})
			return
		}

		testMessages := []openllm.Message{
			{Role: openllm.RoleUser, Content: "Hi"},
		}

		response, err := client.Chat("", testMessages)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"connected": false,
				"error":     "Connection failed: " + err.Error(),
				"code":      "CONNECTION_ERROR",
			})
			return
		}

		if response.Content != "" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"connected": true,
				"message":   "Connection successful!",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"connected": false,
				"error":     "LLM response format error (empty content)",
				"code":      "EMPTY_RESPONSE",
			})
		}
	})

	// POST /api/sales/generate - generate sales training dialogue
	mux.HandleFunc("/api/sales/generate", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
			return
		}

		openllmMutex.Lock()
		client := openllmClient
		openllmMutex.Unlock()

		if client == nil {
			sendError(w, http.StatusBadRequest, "NOT_CONFIGURED", "OpenLLM not configured. Please configure OpenLLM first.")
			return
		}

		// Parse request with language support
		var req struct {
			Product      string `json:"product"`
			CustomerType string `json:"customerType"`
			Goal         string `json:"goal"`
			Language     string `json:"language"` // "en" or "zh"
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body: "+err.Error())
			return
		}

		// Create scenario with language
		scenario := sales.DefaultScenario()
		if req.Product != "" {
			scenario.Product = req.Product
		}
		if req.CustomerType != "" {
			scenario.CustomerType = req.CustomerType
		}
		if req.Goal != "" {
			scenario.Goal = req.Goal
		}
		if req.Language == "zh" {
			scenario.Language = sales.LanguageChinese
		} else {
			scenario.Language = sales.LanguageEnglish
		}

		// Generate system prompt (role definition + format) and user prompt (scenario)
		systemPrompt := sales.GenerateSystemPrompt(scenario.Language)
		userPrompt := sales.GenerateUserPrompt(scenario)

		messages := []openllm.Message{
			{Role: openllm.RoleSystem, Content: systemPrompt},
			{Role: openllm.RoleUser, Content: userPrompt},
		}

		response, err := client.Chat("", messages)
		if err != nil {
			log.Printf("Error generating sales training: %v", err)
			sendError(w, http.StatusInternalServerError, "LLM_ERROR", "Failed to generate sales training: "+err.Error())
			return
		}

		if response.Content == "" {
			sendError(w, http.StatusInternalServerError, "EMPTY_RESPONSE", "LLM returned empty response")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"scenario": scenario,
			"result":   response.Content,
		})
	})

	// POST /api/sales/regenerate - regenerate based on improvements
	mux.HandleFunc("/api/sales/regenerate", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
			return
		}

		openllmMutex.Lock()
		client := openllmClient
		openllmMutex.Unlock()

		if client == nil {
			sendError(w, http.StatusBadRequest, "NOT_CONFIGURED", "OpenLLM not configured. Please configure OpenLLM first.")
			return
		}

		var req struct {
			Product      string   `json:"product"`
			CustomerType string   `json:"customerType"`
			Goal         string   `json:"goal"`
			Language     string   `json:"language"`
			Improvements []string `json:"improvements"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body: "+err.Error())
			return
		}

		if len(req.Improvements) == 0 {
			sendError(w, http.StatusBadRequest, "MISSING_PARAM", "improvements list is required")
			return
		}

		scenario := sales.DefaultScenario()
		if req.Product != "" {
			scenario.Product = req.Product
		}
		if req.CustomerType != "" {
			scenario.CustomerType = req.CustomerType
		}
		if req.Goal != "" {
			scenario.Goal = req.Goal
		}
		if req.Language == "zh" {
			scenario.Language = sales.LanguageChinese
		} else {
			scenario.Language = sales.LanguageEnglish
		}

		systemPrompt := sales.GenerateSystemPrompt(scenario.Language)
		userPrompt := sales.GenerateRegeneratePrompt(scenario, req.Improvements)

		messages := []openllm.Message{
			{Role: openllm.RoleSystem, Content: systemPrompt},
			{Role: openllm.RoleUser, Content: userPrompt},
		}

		response, err := client.Chat("", messages)
		if err != nil {
			log.Printf("Error regenerating sales training: %v", err)
			sendError(w, http.StatusInternalServerError, "LLM_ERROR", "Failed to regenerate: "+err.Error())
			return
		}

		if response.Content == "" {
			sendError(w, http.StatusInternalServerError, "EMPTY_RESPONSE", "LLM returned empty response")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"scenario": scenario,
			"result":   response.Content,
		})
	})

	// Serve frontend static files (catch-all for non-API routes)
	frontendFS := http.FileServer(http.Dir("../frontend"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's an API route
		if strings.HasPrefix(r.URL.Path, "/api/") {
			sendError(w, http.StatusNotFound, "NOT_FOUND", "API endpoint not found")
			return
		}
		frontendFS.ServeHTTP(w, r)
	})

	log.Println("Backend server starting on port 8080...")
	log.Println("Frontend available at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
