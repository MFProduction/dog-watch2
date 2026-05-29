class Recorder {
    constructor(stream) {
        this.stream = stream;
        this.mediaRecorder = null;
        this.chunks = [];
        this.recording = false;
        this.blob = null;
        this.sessionId = null;
        this.chunkIndex = 0;
        this.uploadInterval = null;
        
        this.onStart = null;
        this.onStop = null;
        this.onError = null;
        this.onChunkUploaded = null;
    }

    generateSessionId() {
        const now = new Date();
        const iso = now.toISOString();
        return iso.replace(/[:.]/g, '-').slice(0, -5);
    }

    async uploadChunk(chunk, index, isFinal = false) {
        if (!this.sessionId) return;

        const formData = new FormData();
        formData.append('chunk', chunk);
        formData.append('sessionId', this.sessionId);
        formData.append('index', index.toString());
        formData.append('final', isFinal.toString());

        try {
            const response = await fetch('/api/recordings/chunk', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                console.error('Failed to upload chunk:', response.statusText);
            } else if (this.onChunkUploaded) {
                this.onChunkUploaded(index);
            }
        } catch (error) {
            console.error('Error uploading chunk:', error);
        }
    }

    async uploadPendingChunks() {
        if (this.chunks.length === 0) return;

        const chunksToUpload = [...this.chunks];
        this.chunks = [];

        for (const chunk of chunksToUpload) {
            await this.uploadChunk(chunk, this.chunkIndex, false);
            this.chunkIndex++;
        }
    }

    start() {
        if (this.recording) {
            throw new Error('Already recording');
        }

        if (!MediaRecorder.isTypeSupported('video/webm')) {
            throw new Error('WebM recording not supported in this browser');
        }

        this.chunks = [];
        this.blob = null;
        this.sessionId = this.generateSessionId();
        this.chunkIndex = 0;

        this.mediaRecorder = new MediaRecorder(this.stream, {
            mimeType: 'video/webm'
        });

        this.mediaRecorder.ondataavailable = (event) => {
            if (event.data.size > 0) {
                this.chunks.push(event.data);
            }
        };

        this.mediaRecorder.onstop = () => {
            this.blob = new Blob(this.chunks, { type: 'video/webm' });
            if (this.onStop) {
                this.onStop(this.blob);
            }
        };

        this.mediaRecorder.onerror = (event) => {
            if (this.onError) {
                this.onError(event.error);
            }
        };

        this.mediaRecorder.start(1000);
        this.recording = true;

        // Upload chunks every 5 seconds
        this.uploadInterval = setInterval(() => {
            this.uploadPendingChunks();
        }, 5000);

        if (this.onStart) {
            this.onStart();
        }
    }

    async stop() {
        if (!this.recording) {
            throw new Error('Not recording');
        }

        // Stop periodic uploads
        if (this.uploadInterval) {
            clearInterval(this.uploadInterval);
            this.uploadInterval = null;
        }

        return new Promise((resolve) => {
            this.mediaRecorder.onstop = async () => {
                // Upload any remaining chunks
                const remainingChunks = [...this.chunks];
                this.chunks = [];

                for (let i = 0; i < remainingChunks.length; i++) {
                    const isLast = (i === remainingChunks.length - 1);
                    await this.uploadChunk(remainingChunks[i], this.chunkIndex, isLast);
                    this.chunkIndex++;
                }

                // If no remaining chunks, send empty final marker
                if (remainingChunks.length === 0) {
                    await this.uploadChunk(new Blob(), this.chunkIndex, true);
                }

                this.recording = false;
                if (this.onStop) {
                    this.onStop(null);
                }
                resolve(null);
            };
            this.mediaRecorder.stop();
        });
    }

    isRecording() {
        return this.recording;
    }

    getBlob() {
        return this.blob;
    }

    getSessionId() {
        return this.sessionId;
    }
}
