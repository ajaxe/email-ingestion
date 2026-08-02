package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

// ChallengeRequest represents a typical webhook verification handshake payload.
type ChallengeRequest struct {
	Challenge string `json:"challenge"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("---")
		log.Printf("Received %s request on %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		for name, values := range r.Header {
			for _, value := range values {
				log.Printf("Header: %s: %s", name, value)
			}
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Printf("Body: %s", string(body))

		// Try to parse the body as JSON to extract a challenge token
		var req ChallengeRequest
		if err := json.Unmarshal(body, &req); err == nil && req.Challenge != "" {
			log.Printf("Detected challenge request. Echoing challenge: %s", req.Challenge)
			
			// Echo the challenge back as JSON
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(req)
			return
		}

		// If no challenge was found, simply return a 200 OK
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Webhook tester listening on port %s", port)
	log.Printf("Expose this server using ngrok: ngrok http %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
