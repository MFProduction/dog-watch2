package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"embed"
	"encoding/pem"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dog-watch/internal/room"
	"dog-watch/internal/signaling"
)

//go:embed web/static/*
var staticFiles embed.FS

func main() {
	port := flag.Int("port", 8443, "Server port")
	certDir := flag.String("certs", ".", "Directory to store certificates")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", port)
	}

	tlsConfig, err := getTLSConfig(*certDir)
	if err != nil {
		log.Fatalf("Failed to setup TLS: %v", err)
	}

	r := room.New()
	hub := signaling.NewHub(r)

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
	http.Handle("/", http.FileServerFS(staticFS))

	addr := fmt.Sprintf(":%d", *port)
	server := &http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
	}

	log.Printf("Dog Watch server starting on https://localhost%s", addr)
	log.Printf("Station: https://localhost%s/station", addr)
	log.Printf("Watch: https://localhost%s/watch", addr)
	log.Println("Note: Accept the self-signed certificate warning in your browser")

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}

func getTLSConfig(certDir string) (*tls.Config, error) {
	certFile := filepath.Join(certDir, "cert.pem")
	keyFile := filepath.Join(certDir, "key.pem")

	var cert tls.Certificate
	var err error

	if fileExists(certFile) && fileExists(keyFile) {
		cert, err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing certificates: %w", err)
		}
		log.Println("Loaded existing TLS certificates")
	} else {
		cert, err = generateSelfSignedCert(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}
		log.Println("Generated new self-signed TLS certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

func generateSelfSignedCert(certFile, keyFile string) (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Dog Watch"},
			CommonName:   "Dog Watch Local",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:              []string{"localhost"},
	}

	// Add all local IP addresses to the certificate
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil || ipnet.IP.To16() != nil {
					template.IPAddresses = append(template.IPAddresses, ipnet.IP)
				}
			}
		}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		return tls.Certificate{}, err
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return tls.Certificate{}, err
	}

	return tls.X509KeyPair(certPEM, keyPEM)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
