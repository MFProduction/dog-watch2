package signaling

import (
	"log"
	"net/http"
	"sync"

	"dog-watch/internal/protocol"
	"dog-watch/internal/room"

	"github.com/gorilla/websocket"
)

type SessionRouter interface {
	Register(role string, conn room.Connection) error
	Remove(conn room.Connection)
	GetPeer(role string) room.Connection
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
	router   SessionRouter
	upgrader websocket.Upgrader
}

func NewHub(router SessionRouter) *Hub {
	return &Hub{
		router: router,
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

	registerErr := h.router.Register(role, peer)
	if registerErr == nil {
		log.Printf("%s connected", role)
		h.notifyReady(role)
	} else {
		errMsg := protocol.NewError(registerErr.Error())
		msgBytes, _ := errMsg.Marshal()
		peer.Send(msgBytes)
		conn.Close()
		return
	}

	defer func() {
		h.router.Remove(peer)
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
	_, err := protocol.Unmarshal(message)
	if err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	var peerRole string
	switch role {
	case "station":
		peerRole = "viewer"
	case "viewer":
		peerRole = "station"
	}

	target := h.router.GetPeer(peerRole)
	if target != nil {
		if err := target.Send(message); err != nil {
			log.Printf("Error sending message to %s: %v", peerRole, err)
		}
	}
}

func (h *Hub) notifyReady(role string) {
	var peerRole string

	switch role {
	case "station":
		peerRole = "viewer"
	case "viewer":
		peerRole = "station"
	}

	peer := h.router.GetPeer(peerRole)
	if peer == nil {
		return
	}

	switch role {
	case "station":
		stationMsg, _ := protocol.NewViewerReady().Marshal()
		station := h.router.GetPeer("station")
		if station != nil {
			station.Send(stationMsg)
		}
		viewerMsg, _ := protocol.NewStationReady().Marshal()
		peer.Send(viewerMsg)
	case "viewer":
		viewerMsg, _ := protocol.NewStationReady().Marshal()
		viewer := h.router.GetPeer("viewer")
		if viewer != nil {
			viewer.Send(viewerMsg)
		}
		stationMsg, _ := protocol.NewViewerReady().Marshal()
		peer.Send(stationMsg)
	}
}
