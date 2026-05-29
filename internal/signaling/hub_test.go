package signaling

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dog-watch/internal/protocol"
	"dog-watch/internal/room"

	"github.com/gorilla/websocket"
)

func setupTestServer() (*httptest.Server, *room.Room) {
	r := room.New()
	hub := NewHub(r)

	server := httptest.NewServer(http.HandlerFunc(hub.HandleConnection))
	return server, r
}

func connectWebSocket(t *testing.T, server *httptest.Server, role string) *websocket.Conn {
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?role=" + role
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect WebSocket: %v", err)
	}
	return conn
}

func TestStationConnection(t *testing.T) {
	server, r := setupTestServer()
	defer server.Close()

	conn := connectWebSocket(t, server, "station")
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	hasStation, _ := r.GetState()
	if !hasStation {
		t.Error("expected station to be registered")
	}
}

func TestViewerConnection(t *testing.T) {
	server, r := setupTestServer()
	defer server.Close()

	conn := connectWebSocket(t, server, "viewer")
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	_, hasViewer := r.GetState()
	if !hasViewer {
		t.Error("expected viewer to be registered")
	}
}

func TestInvalidRole(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	conn := connectWebSocket(t, server, "invalid")
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var message protocol.Message
	if err := json.Unmarshal(msg, &message); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if message.Type != "error" {
		t.Errorf("expected error message, got %s", message.Type)
	}
}

func TestRejectSecondStation(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	conn1 := connectWebSocket(t, server, "station")
	defer conn1.Close()

	time.Sleep(50 * time.Millisecond)

	conn2 := connectWebSocket(t, server, "station")
	defer conn2.Close()

	conn2.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var message protocol.Message
	if err := json.Unmarshal(msg, &message); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if message.Type != "error" {
		t.Errorf("expected error message, got %s", message.Type)
	}
}

func TestRejectSecondViewer(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	conn1 := connectWebSocket(t, server, "viewer")
	defer conn1.Close()

	time.Sleep(50 * time.Millisecond)

	conn2 := connectWebSocket(t, server, "viewer")
	defer conn2.Close()

	conn2.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn2.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var message protocol.Message
	if err := json.Unmarshal(msg, &message); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if message.Type != "error" {
		t.Errorf("expected error message, got %s", message.Type)
	}
}

func TestMessageRouting(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	station := connectWebSocket(t, server, "station")
	defer station.Close()

	viewer := connectWebSocket(t, server, "viewer")
	defer viewer.Close()

	time.Sleep(50 * time.Millisecond)

	testMsg := protocol.Message{
		Type: "offer",
		Data: json.RawMessage(`{"sdp":"test"}`),
	}
	msgBytes, _ := json.Marshal(testMsg)

	if err := station.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	viewer.SetReadDeadline(time.Now().Add(time.Second))
	for {
		_, received, err := viewer.ReadMessage()
		if err != nil {
			t.Fatalf("failed to read message: %v", err)
		}

		var receivedMsg protocol.Message
		if err := json.Unmarshal(received, &receivedMsg); err != nil {
			t.Fatalf("failed to unmarshal message: %v", err)
		}

		if receivedMsg.Type == "offer" {
			break
		}
	}
}

func TestViewerReadyNotification(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	station := connectWebSocket(t, server, "station")
	defer station.Close()

	time.Sleep(50 * time.Millisecond)

	viewer := connectWebSocket(t, server, "viewer")
	defer viewer.Close()

	station.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := station.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var message protocol.Message
	if err := json.Unmarshal(msg, &message); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if message.Type != "viewer-ready" {
		t.Errorf("expected viewer-ready message, got %s", message.Type)
	}
}

func TestStationReadyNotification(t *testing.T) {
	server, _ := setupTestServer()
	defer server.Close()

	viewer := connectWebSocket(t, server, "viewer")
	defer viewer.Close()

	time.Sleep(50 * time.Millisecond)

	station := connectWebSocket(t, server, "station")
	defer station.Close()

	viewer.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := viewer.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var message protocol.Message
	if err := json.Unmarshal(msg, &message); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if message.Type != "station-ready" {
		t.Errorf("expected station-ready message, got %s", message.Type)
	}
}

func TestDisconnectCleanup(t *testing.T) {
	server, r := setupTestServer()
	defer server.Close()

	conn := connectWebSocket(t, server, "station")
	time.Sleep(50 * time.Millisecond)

	conn.Close()
	time.Sleep(50 * time.Millisecond)

	hasStation, _ := r.GetState()
	if hasStation {
		t.Error("expected station to be removed after disconnect")
	}
}
