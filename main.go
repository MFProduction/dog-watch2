package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"dog-watch/internal/api"
	"dog-watch/internal/certs"
	"dog-watch/internal/recorder"
	"dog-watch/internal/room"
	"dog-watch/internal/signaling"
)

//go:embed web/static/*
var staticFiles embed.FS

func main() {
	port := flag.Int("port", 8443, "Server port")
	certDir := flag.String("certs", ".", "Directory to store certificates")
	recordingsDir := flag.String("recordings", "./recordings", "Directory to store recordings")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", port)
	}

	tlsConfig, err := certs.GetTLSConfig(*certDir)
	if err != nil {
		log.Fatalf("Failed to setup TLS: %v", err)
	}

	r := room.New()
	hub := signaling.NewHub(r)

	store, err := recorder.NewStore(*recordingsDir)
	if err != nil {
		log.Fatalf("Failed to setup recordings: %v", err)
	}
	apiHandler := api.NewHandler(store)

	staticFS, err := fs.Sub(staticFiles, "web/static")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ws", hub.HandleConnection)
	http.HandleFunc("/station", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFS, "station.html")
	})
	http.HandleFunc("/watch", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFS, "watch.html")
	})
	http.HandleFunc("/recordings", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFS, "recordings.html")
	})
	http.HandleFunc("/api/recordings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			apiHandler.UploadRecording(w, r)
		case http.MethodGet:
			apiHandler.ListRecordings(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/recordings/chunk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			apiHandler.UploadChunk(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/recordings/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			apiHandler.GetRecording(w, r)
		case http.MethodDelete:
			apiHandler.DeleteRecording(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.Handle("/", http.FileServerFS(staticFS))

	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
	}

	log.Printf("Dog Watch server starting on https://localhost%s", addr)
	log.Printf("Station: https://localhost%s/station", addr)
	log.Printf("Watch: https://localhost%s/watch", addr)
	log.Printf("Recordings: https://localhost%s/recordings", addr)
	log.Printf("Recordings directory: %s", *recordingsDir)
	log.Println("Note: Accept the self-signed certificate warning in your browser")

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}
