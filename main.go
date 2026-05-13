package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tbxark/vercel-proxy/internal/proxy"
)

func main() {
	http.HandleFunc("/", proxy.Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
