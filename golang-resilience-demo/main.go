package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rinkachi/golang-demos/golang-resilience-demo/pkg/httpclient"
)

func main() {
	// Start a flaky server to test against
	go startFlakyServer()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	client := httpclient.NewResilientClient()

	for i := 0; i < 10; i++ {
		fmt.Printf("\n--- Call %d ---\n", i+1)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		
		body, err := client.Get(ctx, "http://localhost:8082/resource")
		cancel()

		if err != nil {
			log.Printf("Call failed: %v", err)
		} else {
			log.Printf("Call succeeded: %s", string(body))
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func startFlakyServer() {
	var requestCount int
	mux := http.NewServeMux()
	mux.HandleFunc("/resource", func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// Fail every 2 out of 3 requests
		if requestCount%3 != 0 {
			http.Error(w, "Temporary Failure", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success Data"))
	})

	log.Println("Flaky server starting on :8082")
	http.ListenAndServe(":8082", mux)
}
