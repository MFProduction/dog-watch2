package room

import (
	"errors"
	"sync"
)

var (
	ErrStationAlreadyConnected = errors.New("station already connected")
	ErrViewerAlreadyConnected  = errors.New("viewer already connected")
)

type Connection interface {
	Send(msg []byte) error
}

type Room struct {
	mu       sync.RWMutex
	station  Connection
	viewer   Connection
}

func New() *Room {
	return &Room{}
}

func (r *Room) RegisterStation(conn Connection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.station != nil {
		return ErrStationAlreadyConnected
	}

	r.station = conn
	return nil
}

func (r *Room) RegisterViewer(conn Connection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.viewer != nil {
		return ErrViewerAlreadyConnected
	}

	r.viewer = conn
	return nil
}

func (r *Room) Remove(conn Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.station == conn {
		r.station = nil
	}
	if r.viewer == conn {
		r.viewer = nil
	}
}

func (r *Room) GetStation() Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.station
}

func (r *Room) GetViewer() Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.viewer
}

func (r *Room) GetState() (hasStation, hasViewer bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.station != nil, r.viewer != nil
}
