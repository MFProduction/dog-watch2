package protocol

import "encoding/json"

const (
	TypeOffer         = "offer"
	TypeAnswer        = "answer"
	TypeIceCandidate  = "ice-candidate"
	TypeStationReady  = "station-ready"
	TypeViewerReady   = "viewer-ready"
	TypeError         = "error"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

func NewOffer(data interface{}) (Message, error) {
	return newMessage(TypeOffer, data)
}

func NewAnswer(data interface{}) (Message, error) {
	return newMessage(TypeAnswer, data)
}

func NewIceCandidate(data interface{}) (Message, error) {
	return newMessage(TypeIceCandidate, data)
}

func NewStationReady() Message {
	return Message{Type: TypeStationReady}
}

func NewViewerReady() Message {
	return Message{Type: TypeViewerReady}
}

func NewError(errMsg string) Message {
	data, _ := json.Marshal(errMsg)
	return Message{Type: TypeError, Data: data}
}

func newMessage(msgType string, data interface{}) (Message, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return Message{}, err
	}
	return Message{Type: msgType, Data: rawData}, nil
}

func (m Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func Unmarshal(data []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return msg, err
}
