package recorder

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	if store == nil {
		t.Fatal("expected non-nil store")
	}

	if store.dir != tmpDir {
		t.Errorf("expected dir %s, got %s", tmpDir, store.dir)
	}
}

func TestNewStoreCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "recordings")

	store, err := NewStore(newDir)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}

	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	data := []byte("test recording data")
	filename := "recording-2026-05-29T22-30-00.webm"

	err := store.Save(data, filename)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	filePath := filepath.Join(tmpDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("expected content %s, got %s", data, content)
	}
}

func TestSaveInvalidFilename(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	tests := []struct {
		name     string
		filename string
	}{
		{"empty", ""},
		{"directory traversal", "../recording-2026-05-29T22-30-00.webm"},
		{"absolute path", "/tmp/recording-2026-05-29T22-30-00.webm"},
		{"wrong extension", "recording-2026-05-29T22-30-00.mp4"},
		{"invalid format", "video-2026-05-29T22-30-00.webm"},
		{"hidden file", ".recording-2026-05-29T22-30-00.webm"},
		{"missing timestamp", "recording.webm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Save([]byte("data"), tt.filename)
			if err == nil {
				t.Errorf("expected error for filename %s", tt.filename)
			}
		})
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	recordings, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(recordings) != 0 {
		t.Errorf("expected empty list, got %d recordings", len(recordings))
	}
}

func TestListWithRecordings(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	store.Save([]byte("data1"), "recording-2026-05-29T22-30-00.webm")
	time.Sleep(10 * time.Millisecond)
	store.Save([]byte("data2"), "recording-2026-05-29T22-31-00.webm")
	time.Sleep(10 * time.Millisecond)
	store.Save([]byte("data3"), "recording-2026-05-29T22-32-00.webm")

	recordings, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(recordings) != 3 {
		t.Fatalf("expected 3 recordings, got %d", len(recordings))
	}

	if recordings[0].Filename != "recording-2026-05-29T22-32-00.webm" {
		t.Errorf("expected newest recording first, got %s", recordings[0].Filename)
	}

	if recordings[2].Filename != "recording-2026-05-29T22-30-00.webm" {
		t.Errorf("expected oldest recording last, got %s", recordings[2].Filename)
	}

	for _, rec := range recordings {
		if rec.Size == 0 {
			t.Errorf("expected non-zero size for %s", rec.Filename)
		}
		if rec.CreatedAt.IsZero() {
			t.Errorf("expected non-zero CreatedAt for %s", rec.Filename)
		}
	}
}

func TestListIgnoresInvalidFiles(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	store.Save([]byte("data"), "recording-2026-05-29T22-30-00.webm")
	os.WriteFile(filepath.Join(tmpDir, "invalid.txt"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "recording-2026-05-29T22-31-00.mp4"), []byte("data"), 0644)

	recordings, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(recordings) != 1 {
		t.Errorf("expected 1 valid recording, got %d", len(recordings))
	}
}

func TestGet(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	data := []byte("test recording data")
	filename := "recording-2026-05-29T22-30-00.webm"
	store.Save(data, filename)

	file, err := store.Get(filename)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer file.Close()

	content := make([]byte, len(data))
	n, err := file.Read(content)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if n != len(data) {
		t.Errorf("expected to read %d bytes, read %d", len(data), n)
	}

	if string(content) != string(data) {
		t.Errorf("expected content %s, got %s", data, content)
	}
}

func TestGetNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	_, err := store.Get("recording-2026-05-29T22-30-00.webm")
	if err != ErrFileNotFound {
		t.Errorf("expected ErrFileNotFound, got %v", err)
	}
}

func TestGetInvalidFilename(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	_, err := store.Get("../recording-2026-05-29T22-30-00.webm")
	if err != ErrInvalidFilename {
		t.Errorf("expected ErrInvalidFilename, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	filename := "recording-2026-05-29T22-30-00.webm"
	store.Save([]byte("data"), filename)

	err := store.Delete(filename)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	filePath := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}
}

func TestDeleteNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	err := store.Delete("recording-2026-05-29T22-30-00.webm")
	if err != ErrFileNotFound {
		t.Errorf("expected ErrFileNotFound, got %v", err)
	}
}

func TestDeleteInvalidFilename(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	err := store.Delete("../recording-2026-05-29T22-30-00.webm")
	if err != ErrInvalidFilename {
		t.Errorf("expected ErrInvalidFilename, got %v", err)
	}
}

func TestConcurrentSave(t *testing.T) {
	tmpDir := t.TempDir()
	store, _ := NewStore(tmpDir)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			filename := "recording-2026-05-29T22-30-0" + string(rune('0'+i)) + ".webm"
			store.Save([]byte("data"), filename)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	recordings, _ := store.List()
	if len(recordings) != 10 {
		t.Errorf("expected 10 recordings, got %d", len(recordings))
	}
}
