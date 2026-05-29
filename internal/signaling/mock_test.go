package signaling

import (
	"encoding/json"
	"sync"
	"testing"

	"dog-watch/internal/protocol"
	"dog-watch/internal/room"
)

type mockRouter struct {
	mu        sync.Mutex
	peers     map[string]room.Connection
	registered map[string]bool
}

func newMockRouter() *mockRouter {
	return &mockRouter{
		peers:     make(map[string]room.Connection),
		registered: make(map[string]bool),
	}
}

func (m *mockRouter) Register(role string, conn room.Connection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.registered[role] {
		return room.ErrStationAlreadyConnected
	}

	m.peers[role] = conn
	m.registered[role] = true
	return nil
}

func (m *mockRouter) Remove(conn room.Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for role, c := range m.peers {
		if c == conn {
			delete(m.peers, role)
			delete(m.registered, role)
		}
	}
}

func (m *mockRouter) GetPeer(role string) room.Connection {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.peers[role]
}

type mockConnection struct {
	messages [][]byte
	mu       sync.Mutex
}

func (m *mockConnection) Send(msg []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockConnection) getMessages() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

func TestHubWithMockRouter(t *testing.T) {
	router := newMockRouter()
	hub := NewHub(router)

	if hub == nil {
		t.Fatal("expected non-nil hub")
	}

	if hub.router != router {
		t.Error("expected hub to use provided router")
	}
}

func TestMockRouterRegister(t *testing.T) {
	router := newMockRouter()
	conn := &mockConnection{}

	err := router.Register("station", conn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	peer := router.GetPeer("station")
	if peer != conn {
		t.Error("expected to get registered connection")
	}
}

func TestMockRouterRejectsDuplicate(t *testing.T) {
	router := newMockRouter()
	conn1 := &mockConnection{}
	conn2 := &mockConnection{}

	router.Register("station", conn1)
	err := router.Register("station", conn2)

	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestMockRouterRemove(t *testing.T) {
	router := newMockRouter()
	conn := &mockConnection{}

	router.Register("station", conn)
	router.Remove(conn)

	peer := router.GetPeer("station")
	if peer != nil {
		t.Error("expected peer to be removed")
	}
}

func TestHubMessageRoutingWithMock(t *testing.T) {
	router := newMockRouter()
	stationConn := &mockConnection{}
	viewerConn := &mockConnection{}

	router.Register("station", stationConn)
	router.Register("viewer", viewerConn)

	hub := NewHub(router)

	testMsg := protocol.Message{
		Type: "offer",
		Data: json.RawMessage(`{"sdp":"test"}`),
	}
	msgBytes, _ := json.Marshal(testMsg)

	hub.routeMessage(&Peer{}, "station", msgBytes)

	messages := viewerConn.getMessages()
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
}

func TestHubNotifyReadyWithMock(t *testing.T) {
	router := newMockRouter()
	viewerConn := &mockConnection{}

	router.Register("viewer", viewerConn)

	hub := NewHub(router)
	hub.notifyReady("station")

	messages := viewerConn.getMessages()
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	var msg protocol.Message
	json.Unmarshal(messages[0], &msg)
	if msg.Type != "station-ready" {
		t.Errorf("expected station-ready, got %s", msg.Type)
	}
}
