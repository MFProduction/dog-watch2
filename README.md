# Dog Watch

A simple WebRTC-based dog monitoring system for local networks. Stream video and audio from a computer (station) to a phone (viewer) with zero configuration.

## Features

- Browser-based camera and microphone capture
- Low-latency WebRTC peer-to-peer streaming
- Single binary deployment with embedded frontend
- No authentication required (local network only)
- Auto-discovery: just open `/watch` on your phone

## Usage

1. Run the server on your homeserver
2. Open `http://<server-ip>:8080/station` on the computer with camera/mic
3. Open `http://<server-ip>:8080/watch` on your phone to view the stream

## Building

```bash
go build -o dog-watch
./dog-watch
```

Server runs on port 8080 by default.
