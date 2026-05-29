class MediaManager {
    constructor() {
        this.localStream = null;
        this.remoteStream = null;
        this.recorder = null;
    }

    async getLocalStream(constraints = { video: true, audio: true }) {
        this.localStream = await navigator.mediaDevices.getUserMedia(constraints);
        return this.localStream;
    }

    setRemoteStream(stream) {
        this.remoteStream = stream;
    }

    getLocalTracks() {
        if (!this.localStream) {
            return [];
        }
        return this.localStream.getTracks();
    }

    stopLocalStream() {
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
            this.localStream = null;
        }
    }

    muteLocalAudio(muted) {
        if (this.localStream) {
            this.localStream.getAudioTracks().forEach(track => {
                track.enabled = !muted;
            });
        }
    }

    startRecording() {
        if (!this.localStream) {
            throw new Error('No local stream');
        }
        if (this.recorder && this.recorder.isRecording()) {
            throw new Error('Already recording');
        }

        this.recorder = new Recorder(this.localStream);
        this.recorder.start();
    }

    async stopRecording() {
        if (!this.recorder || !this.recorder.isRecording()) {
            throw new Error('Not recording');
        }

        const blob = await this.recorder.stop();
        return blob;
    }

    isRecording() {
        return this.recorder && this.recorder.isRecording();
    }

    getRecordingBlob() {
        return this.recorder ? this.recorder.getBlob() : null;
    }

    async close() {
        if (this.recorder && this.recorder.isRecording()) {
            await this.recorder.stop();
        }
        this.stopLocalStream();
        this.remoteStream = null;
        this.recorder = null;
    }
}
