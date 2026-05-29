# Dog Watch Protocol

This document defines the WebSocket message protocol used for signaling between the station and viewer clients.

## Message Format

All messages are JSON objects with the following structure:

```json
{
  "type": "<message-type>",
  "data": <optional-payload>
}
```

## Message Types

### offer

Sent by the station to initiate a WebRTC connection.

```json
{
  "type": "offer",
  "data": {
    "type": "offer",
    "sdp": "<SDP-offer-string>"
  }
}
```

### answer

Sent by the viewer in response to an offer.

```json
{
  "type": "answer",
  "data": {
    "type": "answer",
    "sdp": "<SDP-answer-string>"
  }
}
```

### ice-candidate

Sent by either peer during ICE negotiation.

```json
{
  "type": "ice-candidate",
  "data": {
    "candidate": "<ICE-candidate-string>",
    "sdpMid": "<SDP-media-id>",
    "sdpMLineIndex": <index-number>
  }
}
```

### station-ready

Sent by the server to the viewer when a station connects.

```json
{
  "type": "station-ready"
}
```

### viewer-ready

Sent by the server to the station when a viewer connects.

```json
{
  "type": "viewer-ready"
}
```

### error

Sent by the server when an error occurs.

```json
{
  "type": "error",
  "data": "<error-message-string>"
}
```

## Implementation

### Go

The protocol is implemented in `internal/protocol/protocol.go` with type-safe constructors:

```go
import "dog-watch/internal/protocol"

// Create messages
offer, err := protocol.NewOffer(sdpData)
answer, err := protocol.NewAnswer(sdpData)
candidate, err := protocol.NewIceCandidate(candidateData)
stationReady := protocol.NewStationReady()
viewerReady := protocol.NewViewerReady()
errorMsg := protocol.NewError("error message")

// Serialize
data, err := msg.Marshal()

// Deserialize
msg, err := protocol.Unmarshal(data)
```

### JavaScript

Messages are created as plain objects and serialized with JSON:

```javascript
// Create messages
const offer = { type: 'offer', data: sdpOffer };
const answer = { type: 'answer', data: sdpAnswer };
const candidate = { type: 'ice-candidate', data: iceCandidate };

// Serialize
const json = JSON.stringify(message);

// Deserialize
const msg = JSON.parse(json);
```

## Connection Flow

1. Station connects to `/ws?role=station`
2. Viewer connects to `/ws?role=viewer`
3. Server sends `viewer-ready` to station (if viewer already connected)
4. Server sends `station-ready` to viewer (if station already connected)
5. Station creates WebRTC offer and sends to server
6. Server forwards offer to viewer
7. Viewer creates answer and sends to server
8. Server forwards answer to station
9. Both peers exchange `ice-candidate` messages via server
10. WebRTC connection established, media streams flow peer-to-peer

## Error Handling

The server sends error messages in these cases:

- Invalid role parameter
- Station already connected (when second station tries to connect)
- Viewer already connected (when second viewer tries to connect)

Clients should handle error messages and display appropriate feedback to users.
