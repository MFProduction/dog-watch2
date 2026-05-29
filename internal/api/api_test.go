package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"dog-watch/internal/recorder"
)

func setupTestHandler(t *testing.T) (*Handler, string) {
	tmpDir := t.TempDir()
	store, err := recorder.NewStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return NewHandler(store), tmpDir
}

func TestUploadRecording(t *testing.T) {
	handler, _ := setupTestHandler(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("recording", "recording-2026-05-29T22-30-00.webm")
	part.Write([]byte("test video data"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/recordings", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.UploadRecording(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["filename"] != "recording-2026-05-29T22-30-00.webm" {
		t.Errorf("expected filename recording-2026-05-29T22-30-00.webm, got %s", response["filename"])
	}
}

func TestUploadRecordingInvalidFilename(t *testing.T) {
	handler, _ := setupTestHandler(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("recording", "../invalid.webm")
	part.Write([]byte("test video data"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/recordings", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.UploadRecording(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUploadRecordingMissingFile(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/api/recordings", nil)
	w := httptest.NewRecorder()

	handler.UploadRecording(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUploadRecordingWrongMethod(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/api/recordings", nil)
	w := httptest.NewRecorder()

	handler.UploadRecording(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestListRecordings(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/api/recordings", nil)
	w := httptest.NewRecorder()

	handler.ListRecordings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var recordings []recorder.Recording
	json.NewDecoder(w.Body).Decode(&recordings)

	if len(recordings) != 0 {
		t.Errorf("expected 0 recordings, got %d", len(recordings))
	}
}

func TestListRecordingsWithData(t *testing.T) {
	handler, _ := setupTestHandler(t)

	uploadRecording(t, handler, "recording-2026-05-29T22-30-00.webm", "data1")
	uploadRecording(t, handler, "recording-2026-05-29T22-31-00.webm", "data2")

	req := httptest.NewRequest("GET", "/api/recordings", nil)
	w := httptest.NewRecorder()

	handler.ListRecordings(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var recordings []recorder.Recording
	json.NewDecoder(w.Body).Decode(&recordings)

	if len(recordings) != 2 {
		t.Errorf("expected 2 recordings, got %d", len(recordings))
	}
}

func TestListRecordingsWrongMethod(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/api/recordings", nil)
	w := httptest.NewRecorder()

	handler.ListRecordings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestGetRecording(t *testing.T) {
	handler, _ := setupTestHandler(t)

	uploadRecording(t, handler, "recording-2026-05-29T22-30-00.webm", "test video data")

	req := httptest.NewRequest("GET", "/api/recordings/recording-2026-05-29T22-30-00.webm", nil)
	w := httptest.NewRecorder()

	handler.GetRecording(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "video/webm" {
		t.Errorf("expected Content-Type video/webm, got %s", w.Header().Get("Content-Type"))
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != "test video data" {
		t.Errorf("expected body 'test video data', got '%s'", body)
	}
}

func TestGetRecordingNotFound(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/api/recordings/recording-2026-05-29T22-30-00.webm", nil)
	w := httptest.NewRecorder()

	handler.GetRecording(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestGetRecordingInvalidFilename(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/api/recordings/../invalid.webm", nil)
	w := httptest.NewRecorder()

	handler.GetRecording(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetRecordingWrongMethod(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/api/recordings/recording-2026-05-29T22-30-00.webm", nil)
	w := httptest.NewRecorder()

	handler.GetRecording(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestDeleteRecording(t *testing.T) {
	handler, _ := setupTestHandler(t)

	uploadRecording(t, handler, "recording-2026-05-29T22-30-00.webm", "test video data")

	req := httptest.NewRequest("DELETE", "/api/recordings/recording-2026-05-29T22-30-00.webm", nil)
	w := httptest.NewRecorder()

	handler.DeleteRecording(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	listReq := httptest.NewRequest("GET", "/api/recordings", nil)
	listW := httptest.NewRecorder()
	handler.ListRecordings(listW, listReq)

	var recordings []recorder.Recording
	json.NewDecoder(listW.Body).Decode(&recordings)

	if len(recordings) != 0 {
		t.Errorf("expected 0 recordings after delete, got %d", len(recordings))
	}
}

func TestDeleteRecordingNotFound(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("DELETE", "/api/recordings/recording-2026-05-29T22-30-00.webm", nil)
	w := httptest.NewRecorder()

	handler.DeleteRecording(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDeleteRecordingInvalidFilename(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("DELETE", "/api/recordings/../invalid.webm", nil)
	w := httptest.NewRecorder()

	handler.DeleteRecording(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDeleteRecordingWrongMethod(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/api/recordings/recording-2026-05-29T22-30-00.webm", nil)
	w := httptest.NewRecorder()

	handler.DeleteRecording(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func uploadRecording(t *testing.T, handler *Handler, filename, data string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("recording", filename)
	part.Write([]byte(data))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/recordings", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.UploadRecording(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to upload recording: status %d", w.Code)
	}
}
