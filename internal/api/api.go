package api

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"dog-watch/internal/recorder"
)

type Handler struct {
	store *recorder.Store
}

func NewHandler(store *recorder.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) UploadRecording(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(1024 * 1024 * 1024); err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("recording")
	if err != nil {
		http.Error(w, "Missing recording file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := header.Filename
	if filename == "" {
		http.Error(w, "Missing filename", http.StatusBadRequest)
		return
	}

	filename = filepath.Base(filename)

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	if err := h.store.Save(data, filename); err != nil {
		if err == recorder.ErrInvalidFilename {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to save recording", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"filename": filename})
}

func (h *Handler) ListRecordings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recordings, err := h.store.List()
	if err != nil {
		http.Error(w, "Failed to list recordings", http.StatusInternalServerError)
		return
	}

	if recordings == nil {
		recordings = []recorder.Recording{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recordings)
}

func (h *Handler) GetRecording(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := extractFilename(r.URL.Path)
	if filename == "" {
		http.Error(w, "Missing filename", http.StatusBadRequest)
		return
	}

	file, err := h.store.Get(filename)
	if err != nil {
		if err == recorder.ErrFileNotFound {
			http.Error(w, "Recording not found", http.StatusNotFound)
			return
		}
		if err == recorder.ErrInvalidFilename {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to get recording", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Failed to stat file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Content-Length", string(rune(stat.Size())))
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, r, filename, stat.ModTime(), file)
}

func (h *Handler) DeleteRecording(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := extractFilename(r.URL.Path)
	if filename == "" {
		http.Error(w, "Missing filename", http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(filename); err != nil {
		if err == recorder.ErrFileNotFound {
			http.Error(w, "Recording not found", http.StatusNotFound)
			return
		}
		if err == recorder.ErrInvalidFilename {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to delete recording", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractFilename(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func (h *Handler) UploadChunk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(1024 * 1024 * 100); err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	sessionId := r.FormValue("sessionId")
	if sessionId == "" {
		http.Error(w, "Missing sessionId", http.StatusBadRequest)
		return
	}

	indexStr := r.FormValue("index")
	if indexStr == "" {
		http.Error(w, "Missing index", http.StatusBadRequest)
		return
	}

	finalStr := r.FormValue("final")
	isFinal := finalStr == "true"

	file, _, err := r.FormFile("chunk")
	if err != nil {
		http.Error(w, "Missing chunk file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read chunk", http.StatusInternalServerError)
		return
	}

	if err := h.store.SaveChunk(sessionId, indexStr, data, isFinal); err != nil {
		http.Error(w, "Failed to save chunk: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
