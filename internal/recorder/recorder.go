package recorder

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	ErrInvalidFilename = errors.New("invalid filename")
	ErrFileNotFound    = errors.New("file not found")
)

var filenamePattern = regexp.MustCompile(`^recording-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}\.webm$`)

type Recording struct {
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
	dir string
	mu  sync.RWMutex
}

func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create recordings directory: %w", err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &Store{dir: absDir}, nil
}

func (s *Store) Save(data []byte, filename string) error {
	if err := validateFilename(filename); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.dir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *Store) List() ([]Recording, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var recordings []Recording
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if err := validateFilename(filename); err != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		recordings = append(recordings, Recording{
			Filename:  filename,
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(recordings, func(i, j int) bool {
		return recordings[i].CreatedAt.After(recordings[j].CreatedAt)
	})

	return recordings, nil
}

func (s *Store) Get(filename string) (*os.File, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := filepath.Join(s.dir, filename)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (s *Store) Delete(filename string) error {
	if err := validateFilename(filename); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.dir, filename)
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *Store) SaveChunk(sessionId string, indexStr string, data []byte, isFinal bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create chunks directory for this session
	chunksDir := filepath.Join(s.dir, ".chunks", sessionId)
	if err := os.MkdirAll(chunksDir, 0755); err != nil {
		return fmt.Errorf("failed to create chunks directory: %w", err)
	}

	// Save chunk
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid index: %w", err)
	}

	chunkPath := filepath.Join(chunksDir, fmt.Sprintf("%06d.chunk", index))
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	// If this is the final chunk, assemble the recording
	if isFinal {
		return s.assembleRecording(sessionId, chunksDir)
	}

	return nil
}

func (s *Store) assembleRecording(sessionId string, chunksDir string) error {
	// Get all chunk files
	entries, err := os.ReadDir(chunksDir)
	if err != nil {
		return fmt.Errorf("failed to read chunks directory: %w", err)
	}

	// Sort chunks by name (which includes zero-padded index)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Create final recording file
	filename := fmt.Sprintf("recording-%s.webm", sessionId)
	if err := validateFilename(filename); err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}

	finalPath := filepath.Join(s.dir, filename)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("failed to create final file: %w", err)
	}
	defer finalFile.Close()

	// Assemble chunks
	for _, entry := range entries {
		chunkPath := filepath.Join(chunksDir, entry.Name())
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to open chunk: %w", err)
		}

		if _, err := io.Copy(finalFile, chunkFile); err != nil {
			chunkFile.Close()
			return fmt.Errorf("failed to copy chunk: %w", err)
		}
		chunkFile.Close()
	}

	// Clean up chunks directory
	if err := os.RemoveAll(chunksDir); err != nil {
		return fmt.Errorf("failed to remove chunks directory: %w", err)
	}

	return nil
}

func validateFilename(filename string) error {
	if filename == "" {
		return ErrInvalidFilename
	}

	if filepath.Base(filename) != filename {
		return ErrInvalidFilename
	}

	if !filenamePattern.MatchString(filename) {
		return ErrInvalidFilename
	}

	return nil
}
