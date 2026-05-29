# Dog Watch

A simple WebRTC-based dog monitoring system for local networks. Stream video and audio from a computer (station) to a phone (viewer) with zero configuration.

## Features

- Browser-based camera and microphone capture
- Low-latency WebRTC peer-to-peer streaming
- **Video/audio recording** with browser-side capture and server storage
- HTTPS with auto-generated self-signed certificates
- Single binary deployment with embedded frontend
- No authentication required (local network only)
- Auto-discovery: just open `/watch` on your phone

## Why HTTPS?

Camera and microphone access requires a secure context (HTTPS) in modern browsers. The server automatically generates a self-signed certificate on first run. You'll need to accept the certificate warning in your browser.

## Usage

1. Build and run the server on your homeserver:
   ```bash
   go build -o dog-watch
   ./dog-watch
   ```

2. Open `https://<server-ip>:8443/station` on the computer with camera/mic
   - Accept the self-signed certificate warning
   - Click "Start Streaming" and grant camera/mic permissions

3. Open `https://<server-ip>:8443/watch` on your phone to view the stream
   - Accept the self-signed certificate warning
   - Stream appears automatically

## Recording

Record video and audio from the station for later playback:

1. On the station page, click "Start Recording" to begin capturing
2. A red "REC" badge indicates recording is in progress
3. Click "Stop Recording" to stop and upload the recording to the server
4. Open `https://<server-ip>:8443/recordings` to view, play, and delete recordings

Recordings are stored as WebM files in the recordings directory (default: `./recordings/`). If you close the station page while recording, the recording is automatically stopped and uploaded (best-effort).

## Building

```bash
go build -o dog-watch
./dog-watch
```

Server runs on port 8443 by default. Use `-port` flag or `PORT` environment variable to change.

## Command Line Options

```
-port int
    Server port (default 8443)
-certs string
    Directory to store certificates (default ".")
-recordings string
    Directory to store recordings (default "./recordings")
```

## Network Setup

The server automatically detects your local network IP addresses and includes them in the self-signed certificate. This allows access from any device on your local network.
