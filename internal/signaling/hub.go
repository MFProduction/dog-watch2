package signaling

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"dog-watch/internal/room"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type Peer struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (p *Peer) Send(msg []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn.WriteMessage(websocket.TextMessage, msg)
}

type Hub struct {
	room     *room.Room
	upgrader websocket.Upgrader
}

func NewHub(r *room.Room) *Hub {
	return &Hub{
		room: r,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	peer := &Peer{conn: conn}
	role := r.URL.Query().Get("role")

	var registerErr error
	switch role {
	case "station":
		registerErr = h.room.RegisterStation(peer)
		if registerErr == nil {
			log.Println("Station connected")
			h.notifyStationReady()
		}
	case "viewer":
		registerErr = h.room.RegisterViewer(peer)
		if registerErr == nil {
			log.Println("Viewer connected")
			h.notifyViewerReady()
		}
	default:
		errMsg, _ := json.Marshal(Message{Type: "error", Data: json.RawMessage(`"invalid role"`)})
		peer.Send(errMsg)
		conn.Close()
		return
	}

	if registerErr != nil {
		errMsg, _ := json.Marshal(Message{Type: "error", Data: json.RawMessage(`"` + registerErr.Error() + `"`)})
		peer.Send(errMsg)
		conn.Close()
		return
	}

	defer func() {
		h.room.Remove(peer)
		conn.Close()
		log.Printf("%s disconnected", role)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		h.routeMessage(peer, role, message)
	}
}

func (h *Hub) routeMessage(sender *Peer, role string, message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	var target room.Connection
	switch role {
	case "station":
		target = h.room.GetViewer()
	case "viewer":
		target = h.room.GetStation()
	}

	if target != nil {
		if err := target.Send(message); err != nil {
			log.Printf("Error sending message to %s: %v", role, err)
		}
	}
}

func (h *Hub) notifyStationReady() {
	viewer := h.room.GetViewer()
	if viewer != nil {
		msg, _ := json.Marshal(Message{Type: "station-ready"})
		viewer.Send(msg)
	}
}

func (h *Hub) notifyViewerReady() {
	station := h.room.GetStation()
	if station != nil {
		msg, _ := json.Marshal(Message{Type: "viewer-ready"})
		station.Send(msg)
	}
}
