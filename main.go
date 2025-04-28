package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Mappings map[string]string `json:"mappings"`
}

func loadConfig(filename string) (Config, error) {
	var config Config

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

// Custom response writer that captures the response
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseCapture) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseCapture) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func saveResponse(service, path string, statusCode int, headers http.Header, body []byte) error {
	// Create a timestamp and sanitize the path for filename
	timestamp := time.Now().Format("20060102-150405.000")
	sanitizedPath := strings.ReplaceAll(path, "/", "_")
	if sanitizedPath == "" {
		sanitizedPath = "root"
	}

	// Create the responses directory if it doesn't exist
	responsesDir := "responses"
	if err := os.MkdirAll(responsesDir, 0755); err != nil {
		return fmt.Errorf("failed to create responses directory: %w", err)
	}

	// Create a filename with service, path, and timestamp
	filename := filepath.Join(responsesDir, fmt.Sprintf("%s-%s-%s.json", service, sanitizedPath, timestamp))

	// Create a response object to save
	response := struct {
		Service    string      `json:"service"`
		Path       string      `json:"path"`
		Timestamp  string      `json:"timestamp"`
		StatusCode int         `json:"statusCode"`
		Headers    http.Header `json:"headers"`
		Body       string      `json:"body"`
	}{
		Service:    service,
		Path:       path,
		Timestamp:  timestamp,
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(body),
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write response file: %w", err)
	}

	log.Printf("Response saved to %s", filename)
	return nil
}

func proxyHandler(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the first part of the path as the service name
		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.SplitN(path, "/", 2)

		if len(parts) == 0 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		service := parts[0]
		targetURL, exists := config.Mappings[service]

		if !exists {
			http.Error(w, fmt.Sprintf("No mapping found for service: %s", service), http.StatusNotFound)
			return
		}

		// Parse the target URL
		target, err := url.Parse(targetURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid target URL: %s", err), http.StatusInternalServerError)
			return
		}

		// Create a new path without the service prefix
		var newPath string
		if len(parts) > 1 {
			newPath = parts[1]
		} else {
			newPath = ""
		}

		// Update the request URL path
		r.URL.Path = "/" + newPath

		// Create response capture wrapper
		capture := &responseCapture{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code
			body:           &bytes.Buffer{},
		}

		// Create and configure the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Store the default Director function
		defaultDirector := proxy.Director

		// Override the Director function to log the complete target URL
		proxy.Director = func(req *http.Request) {
			defaultDirector(req) // Call the default Director
			// Set the host header to the target host
			req.Host = target.Host
			log.Printf("Proxying request: %s %s -> %s%s\n", req.Method, path, targetURL, req.URL.Path)
		}

		// Override the ModifyResponse function to save the response
		proxy.ModifyResponse = func(resp *http.Response) error {
			// Read the response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			// Save the response to a file
			if err := saveResponse(service, newPath, resp.StatusCode, resp.Header, body); err != nil {
				log.Printf("Warning: Failed to save response: %v", err)
			}

			// Put the body back
			resp.Body = io.NopCloser(bytes.NewReader(body))
			return nil
		}

		// Serve the request
		proxy.ServeHTTP(capture, r)
	}
}

func main() {
	config, err := loadConfig("reserve-config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Print loaded mappings
	fmt.Println("Loaded proxy mappings:")
	for service, target := range config.Mappings {
		fmt.Printf("  /%s -> %s\n", service, target)
	}

	fmt.Println("Responses will be saved to the 'responses' directory")

	http.HandleFunc("/", proxyHandler(config))

	port := ":8448"
	fmt.Printf("Starting proxy server on%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
