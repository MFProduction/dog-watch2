package certs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTLSConfigGeneratesNewCerts(t *testing.T) {
	tmpDir := t.TempDir()

	config, err := GetTLSConfig(tmpDir)
	if err != nil {
		t.Fatalf("GetTLSConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("expected non-nil TLS config")
	}

	if len(config.Certificates) != 1 {
		t.Errorf("expected 1 certificate, got %d", len(config.Certificates))
	}

	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")

	if !fileExists(certFile) {
		t.Error("expected cert.pem to be created")
	}
	if !fileExists(keyFile) {
		t.Error("expected key.pem to be created")
	}
}

func TestGetTLSConfigLoadsExistingCerts(t *testing.T) {
	tmpDir := t.TempDir()

	config1, err := GetTLSConfig(tmpDir)
	if err != nil {
		t.Fatalf("first GetTLSConfig failed: %v", err)
	}

	config2, err := GetTLSConfig(tmpDir)
	if err != nil {
		t.Fatalf("second GetTLSConfig failed: %v", err)
	}

	if len(config1.Certificates) != len(config2.Certificates) {
		t.Error("expected same number of certificates")
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	if !fileExists(existingFile) {
		t.Error("expected fileExists to return true for existing file")
	}

	nonExistent := filepath.Join(tmpDir, "does-not-exist.txt")
	if fileExists(nonExistent) {
		t.Error("expected fileExists to return false for non-existent file")
	}
}

func TestGetTLSConfigWithInvalidCertDir(t *testing.T) {
	invalidDir := "/nonexistent/path/that/should/not/exist"

	_, err := GetTLSConfig(invalidDir)
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}

func TestCertificateContainsLocalIPs(t *testing.T) {
	tmpDir := t.TempDir()

	config, err := GetTLSConfig(tmpDir)
	if err != nil {
		t.Fatalf("GetTLSConfig failed: %v", err)
	}

	if len(config.Certificates) == 0 {
		t.Error("expected at least one certificate")
	}

	cert := config.Certificates[0]
	if len(cert.Certificate) == 0 {
		t.Error("expected certificate data to be present")
	}
}
