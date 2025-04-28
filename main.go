package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
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

		// Create and configure the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Set the host header to the target host
		r.Host = target.Host

		// Log the proxied request
		log.Printf("Proxying request: %s %s -> %s%s\n", r.Method, path, targetURL, r.URL.Path)

		// Serve the request
		proxy.ServeHTTP(w, r)
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

	http.HandleFunc("/", proxyHandler(config))

	port := ":8448"
	fmt.Printf("Starting proxy server on%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
