# PRD: Recording Feature for Dog Watch

**Status:** Ready for Implementation  
**Date:** 2026-05-29  
**Labels:** ready-for-agent

## Problem Statement

As a dog owner using Dog Watch, I want to record video and audio from the station so that I can review what my dog did while I was away, even if I wasn't actively watching the live stream. Currently, the system only supports live monitoring with no way to capture or replay sessions.

## Solution

Add browser-side recording capability to the station page that captures the local media stream using the MediaRecorder API. Recordings are manually started and stopped by the user, then uploaded to the server as WebM files. A new recordings page provides a UI to list, play, and delete recorded sessions. The server stores recordings on the local filesystem with metadata derived from the files themselves.

## User Stories

1. As a dog owner, I want to start recording from the station page, so that I can capture video and audio of my dog while I'm away
2. As a dog owner, I want to stop recording manually, so that I can control when the recording ends
3. As a dog owner, I want recordings to be saved automatically when I stop, so that I don't lose the captured footage
4. As a dog owner, I want recordings to be uploaded to the server, so that they persist after I close the browser
5. As a dog owner, I want to record even when no viewer is connected, so that I can capture footage without actively monitoring
6. As a dog owner, I want recordings to be saved if I accidentally close the station page, so that I don't lose footage due to browser crashes or navigation
7. As a dog owner, I want to see a list of all recordings, so that I can review past sessions
8. As a dog owner, I want to see the filename, size, and timestamp of each recording, so that I can identify when it was recorded
9. As a dog owner, I want to play recordings directly in the browser, so that I can watch them without downloading
10. As a dog owner, I want to delete recordings I no longer need, so that I can manage storage space
11. As a dog owner, I want recordings to be named with timestamps, so that I can easily identify them chronologically
12. As a dog owner, I want to configure where recordings are stored on the server, so that I can use a dedicated storage drive
13. As a dog owner, I want the recordings directory to be created automatically, so that I don't need to manually set it up
14. As a dog owner, I want to access the recordings page from the index page, so that I can easily navigate to my recordings
15. As a dog owner, I want recordings to be in WebM format, so that they play natively in browsers without codec issues
16. As a dog owner, I want to see visual feedback when recording is in progress, so that I know the system is capturing
17. As a dog owner, I want the recording to use the same audio/video as the live stream, so that what I record matches what I see live
18. As a dog owner, I want recordings to be served with proper video headers, so that they stream smoothly in the browser
19. As a dog owner, I want the recording API to validate filenames, so that I can't accidentally delete or access files outside the recordings directory
20. As a dog owner, I want to see an empty state when there are no recordings, so that I know the system is working correctly

## Implementation Decisions

### Module: Recorder Storage (internal/recorder)

**Purpose:** Deep module encapsulating all recording file operations. Provides a simple interface for saving, listing, retrieving, and deleting recordings.

**Interface:**
- `NewStore(dir string) (*Store, error)` — Initialize storage with directory path, creates directory if it doesn't exist
- `Save(data []byte, filename string) error` — Save recording data to file
- `List() ([]Recording, error)` — List all recordings with metadata
- `Get(filename string) (*os.File, error)` — Open recording file for streaming
- `Delete(filename string) error` — Delete recording file

**Recording struct:**
```go
type Recording struct {
    Filename  string    `json:"filename"`
    Size      int64     `json:"size"`
    CreatedAt time.Time `json:"createdAt"`
}
```

**Key behaviors:**
- Validates filenames to prevent directory traversal attacks (must not contain `/`, `..`, or start with `.`)
- Only accepts `.webm` extension
- Sorts recordings by creation time (newest first)
- Derives metadata from filesystem (filename, size, mtime)
- Thread-safe for concurrent access

**Filename format:** `recording-2026-05-29T22-30-00.webm` (ISO 8601 timestamp with colons replaced by hyphens)

### Module: Recordings API (internal/api)

**Purpose:** HTTP handlers for recording endpoints. Thin wiring layer that delegates to Recorder Storage.

**Endpoints:**

1. **POST /api/recordings**
   - Accepts multipart form data with file field named "recording"
   - Extracts filename from Content-Disposition header
   - Validates filename format
   - Saves to Recorder Storage
   - Returns 201 Created with JSON: `{"filename": "recording-2026-05-29T22-30-00.webm"}`

2. **GET /api/recordings**
   - Returns JSON array of Recording objects
   - Sorted by createdAt descending (newest first)
   - Returns 200 OK with JSON: `[{"filename": "...", "size": 12345, "createdAt": "..."}]`

3. **GET /api/recordings/{filename}**
   - Streams recording file with proper Content-Type: `video/webm`
   - Sets Content-Length header
   - Supports HTTP Range requests for seeking
   - Returns 404 if file not found

4. **DELETE /api/recordings/{filename}**
   - Deletes recording file
   - Returns 204 No Content on success
   - Returns 404 if file not found

**Error handling:**
- Invalid filename: 400 Bad Request
- File not found: 404 Not Found
- Storage error: 500 Internal Server Error

### Module: Browser Recorder (web/static/recorder.js)

**Purpose:** Deep module wrapping MediaRecorder API. Handles chunk accumulation, lifecycle management, and auto-stop on disconnect.

**Interface:**
- `new Recorder(stream)` — Initialize with MediaStream
- `start()` — Start recording, begins accumulating chunks
- `stop()` — Stop recording, returns Promise that resolves with Blob
- `isRecording()` — Returns boolean indicating recording state
- `getBlob()` — Returns accumulated Blob (only valid after stop)

**Key behaviors:**
- Uses `MediaRecorder` with `video/webm` MIME type
- Accumulates chunks in array via `ondataavailable` event
- Listens for `beforeunload` event to auto-stop and upload on page close
- Provides callback hooks: `onStart`, `onStop`, `onError`
- Handles browser compatibility (checks `MediaRecorder.isTypeSupported`)

**Auto-stop on disconnect:**
```javascript
window.addEventListener('beforeunload', async () => {
    if (this.isRecording()) {
        this.stop();
        // Attempt upload (best-effort, may not complete)
        await this.uploadBlob();
    }
});
```

### Module: Recordings Page (web/static/recordings.html)

**Purpose:** UI for managing recordings. Lists all recordings with play and delete options.

**Layout:**
- Header with title "Dog Watch Recordings"
- Link back to index page
- Empty state message when no recordings exist
- Grid/list of recording cards, each showing:
  - Filename (as title)
  - File size (formatted as MB/GB)
  - Created timestamp (formatted as human-readable date/time)
  - Play button (opens video player)
  - Delete button (with confirmation)

**Interactions:**
- Uses htmx to fetch `/api/recordings` on page load
- Play button opens modal with `<video>` element, sets src to `/api/recordings/{filename}`
- Delete button shows confirmation dialog, then sends DELETE request
- After delete, removes card from list with htmx
- Shows loading spinner while fetching

**Styling:**
- Bootstrap 5 for layout and components
- Responsive grid (3 columns on desktop, 1 on mobile)
- Video player modal with full-width video element

### Module: Media Manager (Modified)

**Changes:**
- Add `startRecording()` method — creates Recorder instance with local stream, calls `start()`
- Add `stopRecording()` method — calls `stop()` on Recorder, returns Blob
- Add `isRecording()` method — delegates to Recorder
- Add `getRecordingBlob()` method — returns Blob from Recorder

**Interface additions:**
```javascript
class MediaManager {
    // ... existing methods ...
    
    startRecording() {
        if (!this.localStream) throw new Error('No local stream');
        this.recorder = new Recorder(this.localStream);
        this.recorder.start();
    }
    
    async stopRecording() {
        if (!this.recorder) throw new Error('Not recording');
        const blob = await this.recorder.stop();
        return blob;
    }
    
    isRecording() {
        return this.recorder && this.recorder.isRecording();
    }
}
```

### Module: Station Page (Modified)

**Changes:**
- Add "Start Recording" button (visible when streaming but not recording)
- Add "Stop Recording" button (visible when recording)
- Wire up button click handlers to MediaManager recording methods
- Add `beforeunload` event listener to auto-stop and upload on page close
- Show recording indicator (red dot or badge) when recording is active
- Update button visibility based on recording state

**Button states:**
- Before streaming: Show "Start Streaming" only
- Streaming, not recording: Show "Stop Streaming" and "Start Recording"
- Streaming and recording: Show "Stop Streaming" and "Stop Recording"

**Upload logic:**
```javascript
async function stopAndUploadRecording() {
    const blob = await client.media.stopRecording();
    const formData = new FormData();
    formData.append('recording', blob, generateFilename());
    
    await fetch('/api/recordings', {
        method: 'POST',
        body: formData
    });
}

function generateFilename() {
    const now = new Date();
    const timestamp = now.toISOString().replace(/[:.]/g, '-').slice(0, -5);
    return `recording-${timestamp}.webm`;
}
```

### Module: Index Page (Modified)

**Changes:**
- Add third card for "Recordings" with link to `/recordings`
- Use Bootstrap icon (e.g., `bi-camera-video`) for visual consistency

### Module: Main (Modified)

**Changes:**
- Add `-recordings` flag with default `./recordings/`
- Initialize Recorder Storage with configured directory
- Register new API routes:
  - `POST /api/recordings` → `api.UploadRecording`
  - `GET /api/recordings` → `api.ListRecordings`
  - `GET /api/recordings/{filename}` → `api.GetRecording`
  - `DELETE /api/recordings/{filename}` → `api.DeleteRecording`
- Register recordings page route:
  - `GET /recordings` → serve `recordings.html`

**Flag definition:**
```go
recordingsDir := flag.String("recordings", "./recordings", "Directory to store recordings")
```

### Recording Filename Generation

**Format:** `recording-YYYY-MM-DDTHH-MM-SS.webm`

**Generation logic (browser-side):**
```javascript
function generateFilename() {
    const now = new Date();
    const iso = now.toISOString(); // 2026-05-29T22:30:00.000Z
    const timestamp = iso.replace(/[:.]/g, '-').slice(0, -5); // Remove milliseconds and Z
    return `recording-${timestamp}.webm`;
}
```

**Example:** Station starts recording at 2026-05-29 22:30:00 UTC → filename: `recording-2026-05-29T22-30-00.webm`

### API Contract

**POST /api/recordings**

Request:
```
POST /api/recordings
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="recording"; filename="recording-2026-05-29T22-30-00.webm"
Content-Type: video/webm

[binary data]
------WebKitFormBoundary--
```

Response:
```
HTTP/1.1 201 Created
Content-Type: application/json

{
  "filename": "recording-2026-05-29T22-30-00.webm"
}
```

**GET /api/recordings**

Response:
```
HTTP/1.1 200 OK
Content-Type: application/json

[
  {
    "filename": "recording-2026-05-29T22-30-00.webm",
    "size": 12345678,
    "createdAt": "2026-05-29T22:30:00Z"
  },
  {
    "filename": "recording-2026-05-29T21-15-00.webm",
    "size": 9876543,
    "createdAt": "2026-05-29T21:15:00Z"
  }
]
```

**GET /api/recordings/{filename}**

Response:
```
HTTP/1.1 200 OK
Content-Type: video/webm
Content-Length: 12345678
Accept-Ranges: bytes

[binary video data]
```

**DELETE /api/recordings/{filename}**

Response:
```
HTTP/1.1 204 No Content
```

### Security Considerations

**Filename validation:**
- Must not contain `/` or `\` (directory separators)
- Must not contain `..` (directory traversal)
- Must not start with `.` (hidden files)
- Must end with `.webm` extension
- Must match pattern: `^recording-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}\.webm$`

**File size limits:**
- No explicit limit (relies on filesystem capacity)
- Browser memory limits practical recordings to ~1-2 hours at 720p

**Access control:**
- No authentication (consistent with rest of application)
- Assumes trusted local network

## Testing Decisions

### Testing Philosophy

Tests should verify external behavior and public interfaces, not internal implementation details. Tests should be fast, deterministic, and isolated. Integration tests should verify end-to-end flows.

### Modules to Test

**Recorder Storage (unit tests)**
- Test that Save creates file with correct name and content
- Test that List returns all recordings sorted by creation time
- Test that List returns empty array when no recordings exist
- Test that Get opens file for reading
- Test that Get returns error for non-existent file
- Test that Delete removes file from filesystem
- Test that Delete returns error for non-existent file
- Test that Save rejects invalid filenames (directory traversal, wrong extension)
- Test that Get rejects invalid filenames
- Test that Delete rejects invalid filenames
- Test that Store creates directory if it doesn't exist
- Test concurrent Save operations don't corrupt files

**Recordings API (integration tests)**
- Test POST /api/recordings uploads file successfully
- Test POST /api/recordings returns 400 for invalid filename
- Test POST /api/recordings returns 400 for missing file
- Test GET /api/recordings returns list of recordings
- Test GET /api/recordings returns empty array when no recordings
- Test GET /api/recordings/{filename} streams file with correct Content-Type
- Test GET /api/recordings/{filename} returns 404 for non-existent file
- Test GET /api/recordings/{filename} returns 400 for invalid filename
- Test DELETE /api/recordings/{filename} deletes file successfully
- Test DELETE /api/recordings/{filename} returns 404 for non-existent file
- Test DELETE /api/recordings/{filename} returns 400 for invalid filename
- Test HTTP Range requests for video seeking

**Browser Recorder (browser tests)**
- Test that start() begins recording
- Test that stop() returns Blob with recorded data
- Test that isRecording() returns correct state
- Test that chunks are accumulated correctly
- Test that auto-stop triggers on beforeunload
- Test error handling when MediaRecorder fails

### Prior Art

The codebase has existing test patterns:
- `internal/room/room_test.go` — Unit tests with mock connections
- `internal/signaling/hub_test.go` — Integration tests with httptest server and WebSocket clients
- `internal/certs/certs_test.go` — Unit tests with temporary directories

Follow these patterns:
- Use `t.TempDir()` for isolated filesystem tests
- Use `httptest.NewServer` for API integration tests
- Use table-driven tests for validation scenarios
- Mock external dependencies (filesystem, HTTP clients)

## Out of Scope

- Server-side recording (all recording happens in browser)
- Automatic recording (user must manually start/stop)
- Motion detection or smart recording triggers
- Recording quality selection (uses browser defaults)
- Transcoding to other formats (MP4, HLS)
- Cloud storage integration (S3, etc.)
- Recording scheduling or automation
- Multi-station recording (only one station supported)
- Recording metadata beyond filename/size/timestamp
- Recording search or filtering
- Recording export or download (only streaming playback)
- Recording encryption or access control
- Recording retention policies or auto-deletion
- Thumbnail generation for recordings
- Recording duration display in list (browser player shows this)
- Concurrent recording from multiple stations

## Further Notes

**Browser compatibility:** MediaRecorder API is supported in Chrome, Firefox, Edge, and Safari 14.1+. The application should check `MediaRecorder.isTypeSupported('video/webm')` and show an error if not supported.

**Storage capacity:** A 1-hour 720p WebM recording is approximately 200-500MB depending on motion and audio. Users should monitor disk usage and delete old recordings manually.

**Upload reliability:** The `beforeunload` auto-upload is best-effort. If the browser crashes or network fails, the recording may be lost. For critical recordings, users should manually stop and upload before closing the page.

**Performance:** Large recordings (>500MB) may take several seconds to upload depending on network speed. The UI should show upload progress or a loading indicator.

**File locking:** The Recorder Storage module should handle concurrent access safely. Multiple simultaneous uploads should not corrupt files. Use appropriate file locking or atomic writes if needed.

**Timestamp timezone:** Filenames use UTC timestamps to avoid ambiguity. The recordings page should display timestamps in the user's local timezone.

**Video seeking:** The GET endpoint should support HTTP Range requests to allow seeking within recordings. This is standard for video streaming and expected by browser video players.

**Cleanup strategy:** No automatic cleanup is implemented. Users are responsible for managing storage by deleting old recordings through the UI. A future enhancement could add retention policies (e.g., "delete recordings older than 30 days").
