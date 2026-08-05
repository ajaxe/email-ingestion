package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Hello, world!")
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("error: %w", err)
	}
	apiKeyStr := hex.EncodeToString(b)

	hashBytes := sha256.Sum256([]byte(apiKeyStr))
	hashStr := hex.EncodeToString(hashBytes[:])

	fmt.Printf("API Key: %s\nHashed key: %s\n\n", apiKeyStr, hashStr)

	hashBytes = sha256.Sum256([]byte(apiKeyStr))
	rehashStr := hex.EncodeToString(hashBytes[:])

	fmt.Printf("API Key: %s\nRe-Hashed key: %s\n", apiKeyStr, rehashStr)

	if subtle.ConstantTimeCompare([]byte(rehashStr), []byte(hashStr)) == 1 {
		fmt.Println("Success!")
	} else {
		fmt.Println("Failure!")
	}
}
