package room

import (
	"testing"
)

type mockConnection struct {
	messages [][]byte
}

func (m *mockConnection) Send(msg []byte) error {
	m.messages = append(m.messages, msg)
	return nil
}

func TestRegisterStation(t *testing.T) {
	r := New()
	conn := &mockConnection{}

	err := r.RegisterStation(conn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasStation, _ := r.GetState()
	if !hasStation {
		t.Error("expected station to be registered")
	}
}

func TestRegisterViewer(t *testing.T) {
	r := New()
	conn := &mockConnection{}

	err := r.RegisterViewer(conn)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	_, hasViewer := r.GetState()
	if !hasViewer {
		t.Error("expected viewer to be registered")
	}
}

func TestRejectSecondStation(t *testing.T) {
	r := New()
	conn1 := &mockConnection{}
	conn2 := &mockConnection{}

	err := r.RegisterStation(conn1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = r.RegisterStation(conn2)
	if err != ErrStationAlreadyConnected {
		t.Errorf("expected ErrStationAlreadyConnected, got %v", err)
	}
}

func TestRejectSecondViewer(t *testing.T) {
	r := New()
	conn1 := &mockConnection{}
	conn2 := &mockConnection{}

	err := r.RegisterViewer(conn1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = r.RegisterViewer(conn2)
	if err != ErrViewerAlreadyConnected {
		t.Errorf("expected ErrViewerAlreadyConnected, got %v", err)
	}
}

func TestRemoveStation(t *testing.T) {
	r := New()
	conn := &mockConnection{}

	r.RegisterStation(conn)
	r.Remove(conn)

	hasStation, _ := r.GetState()
	if hasStation {
		t.Error("expected station to be removed")
	}
}

func TestRemoveViewer(t *testing.T) {
	r := New()
	conn := &mockConnection{}

	r.RegisterViewer(conn)
	r.Remove(conn)

	_, hasViewer := r.GetState()
	if hasViewer {
		t.Error("expected viewer to be removed")
	}
}

func TestAllowNewStationAfterRemove(t *testing.T) {
	r := New()
	conn1 := &mockConnection{}
	conn2 := &mockConnection{}

	r.RegisterStation(conn1)
	r.Remove(conn1)

	err := r.RegisterStation(conn2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAllowNewViewerAfterRemove(t *testing.T) {
	r := New()
	conn1 := &mockConnection{}
	conn2 := &mockConnection{}

	r.RegisterViewer(conn1)
	r.Remove(conn1)

	err := r.RegisterViewer(conn2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGetStation(t *testing.T) {
	r := New()
	conn := &mockConnection{}

	r.RegisterStation(conn)

	if r.GetStation() != conn {
		t.Error("expected GetStation to return registered connection")
	}
}

func TestGetViewer(t *testing.T) {
	r := New()
	conn := &mockConnection{}

	r.RegisterViewer(conn)

	if r.GetViewer() != conn {
		t.Error("expected GetViewer to return registered connection")
	}
}

func TestGetStateEmpty(t *testing.T) {
	r := New()

	hasStation, hasViewer := r.GetState()
	if hasStation || hasViewer {
		t.Error("expected empty state")
	}
}

func TestGetStateBothConnected(t *testing.T) {
	r := New()
	station := &mockConnection{}
	viewer := &mockConnection{}

	r.RegisterStation(station)
	r.RegisterViewer(viewer)

	hasStation, hasViewer := r.GetState()
	if !hasStation || !hasViewer {
		t.Error("expected both station and viewer to be connected")
	}
}
