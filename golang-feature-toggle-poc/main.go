package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.flipt.io/flipt/rpc/go/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to Flipt
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Flipt: %v", err)
	}
	defer conn.Close()

	fliptClient := client.New(conn)

	// Simple HTTP server using flags
	http.HandleFunc("/feature", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		// Check boolean flag 'new-ui-enabled'
		enabled, err := checkFlag(ctx, fliptClient, "new-ui-enabled", map[string]string{
			"userId": "user-123", // Context for segmentation
		})
		
		if err != nil {
			log.Printf("Flag check error: %v", err)
			// Fallback behavior
			fmt.Fprintln(w, "Old UI (Fallback)")
			return
		}

		if enabled {
			fmt.Fprintln(w, "✨ New UI Enabled! ✨")
		} else {
			fmt.Fprintln(w, "Old UI")
		}
	})

	log.Println("Server running on :8083")
	http.ListenAndServe(":8083", nil)
}

func checkFlag(ctx context.Context, c client.Client, flagKey string, context map[string]string) (bool, error) {
	resp, err := c.Evaluation().Boolean(ctx, &client.EvaluationRequest{
		NamespaceKey: "default",
		FlagKey:      flagKey,
		EntityId:     context["userId"], // Often used as EntityId
		Context:      context,
	})
	if err != nil {
		return false, err
	}
	return resp.Enabled, nil
}
