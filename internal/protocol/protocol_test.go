package protocol

import (
	"encoding/json"
	"testing"
)

func TestNewOffer(t *testing.T) {
	data := map[string]string{"sdp": "test-offer"}
	msg, err := NewOffer(data)
	if err != nil {
		t.Fatalf("NewOffer failed: %v", err)
	}

	if msg.Type != TypeOffer {
		t.Errorf("expected type %s, got %s", TypeOffer, msg.Type)
	}

	if len(msg.Data) == 0 {
		t.Error("expected non-empty data")
	}
}

func TestNewAnswer(t *testing.T) {
	data := map[string]string{"sdp": "test-answer"}
	msg, err := NewAnswer(data)
	if err != nil {
		t.Fatalf("NewAnswer failed: %v", err)
	}

	if msg.Type != TypeAnswer {
		t.Errorf("expected type %s, got %s", TypeAnswer, msg.Type)
	}
}

func TestNewIceCandidate(t *testing.T) {
	data := map[string]string{"candidate": "test-candidate"}
	msg, err := NewIceCandidate(data)
	if err != nil {
		t.Fatalf("NewIceCandidate failed: %v", err)
	}

	if msg.Type != TypeIceCandidate {
		t.Errorf("expected type %s, got %s", TypeIceCandidate, msg.Type)
	}
}

func TestNewStationReady(t *testing.T) {
	msg := NewStationReady()

	if msg.Type != TypeStationReady {
		t.Errorf("expected type %s, got %s", TypeStationReady, msg.Type)
	}

	if len(msg.Data) != 0 {
		t.Error("expected empty data for station-ready")
	}
}

func TestNewViewerReady(t *testing.T) {
	msg := NewViewerReady()

	if msg.Type != TypeViewerReady {
		t.Errorf("expected type %s, got %s", TypeViewerReady, msg.Type)
	}

	if len(msg.Data) != 0 {
		t.Error("expected empty data for viewer-ready")
	}
}

func TestNewError(t *testing.T) {
	msg := NewError("test error")

	if msg.Type != TypeError {
		t.Errorf("expected type %s, got %s", TypeError, msg.Type)
	}

	var errMsg string
	if err := json.Unmarshal(msg.Data, &errMsg); err != nil {
		t.Fatalf("failed to unmarshal error data: %v", err)
	}

	if errMsg != "test error" {
		t.Errorf("expected error message 'test error', got '%s'", errMsg)
	}
}

func TestMessageMarshal(t *testing.T) {
	msg := NewStationReady()
	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty marshaled data")
	}

	var unmarshaled Message
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Type != TypeStationReady {
		t.Errorf("expected type %s, got %s", TypeStationReady, unmarshaled.Type)
	}
}

func TestUnmarshal(t *testing.T) {
	original := NewViewerReady()
	data, _ := original.Marshal()

	msg, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if msg.Type != TypeViewerReady {
		t.Errorf("expected type %s, got %s", TypeViewerReady, msg.Type)
	}
}

func TestUnmarshalInvalidJSON(t *testing.T) {
	_, err := Unmarshal([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMessageConstants(t *testing.T) {
	if TypeOffer != "offer" {
		t.Errorf("TypeOffer should be 'offer', got '%s'", TypeOffer)
	}
	if TypeAnswer != "answer" {
		t.Errorf("TypeAnswer should be 'answer', got '%s'", TypeAnswer)
	}
	if TypeIceCandidate != "ice-candidate" {
		t.Errorf("TypeIceCandidate should be 'ice-candidate', got '%s'", TypeIceCandidate)
	}
	if TypeStationReady != "station-ready" {
		t.Errorf("TypeStationReady should be 'station-ready', got '%s'", TypeStationReady)
	}
	if TypeViewerReady != "viewer-ready" {
		t.Errorf("TypeViewerReady should be 'viewer-ready', got '%s'", TypeViewerReady)
	}
	if TypeError != "error" {
		t.Errorf("TypeError should be 'error', got '%s'", TypeError)
	}
}
