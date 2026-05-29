class WebRTCClient {
    constructor(role) {
        this.role = role;
        this.pc = null;
        this.ws = null;
        this.localStream = null;
        this.remoteStream = null;
        this.onRemoteStream = null;
        this.onStatusChange = null;

        this.iceServers = [
            { urls: 'stun:stun.l.google.com:19302' }
        ];
    }

    createPeerConnection() {
        this.pc = new RTCPeerConnection({ iceServers: this.iceServers });

        this.pc.onicecandidate = (event) => {
            if (event.candidate) {
                this.sendMessage({
                    type: 'ice-candidate',
                    data: event.candidate
                });
            }
        };

        this.pc.ontrack = (event) => {
            this.remoteStream = event.streams[0];
            if (this.onRemoteStream) {
                this.onRemoteStream(this.remoteStream);
            }
        };

        this.pc.onconnectionstatechange = () => {
            if (this.onStatusChange) {
                this.onStatusChange(this.pc.connectionState);
            }
        };

        this.pc.oniceconnectionstatechange = () => {
            console.log('ICE connection state:', this.pc.iceConnectionState);
        };
    }

    addLocalStream(stream) {
        this.localStream = stream;
        stream.getTracks().forEach(track => {
            this.pc.addTrack(track, stream);
        });
    }

    async createOffer() {
        const offer = await this.pc.createOffer();
        await this.pc.setLocalDescription(offer);
        this.sendMessage({
            type: 'offer',
            data: offer
        });
    }

    async handleOffer(offer) {
        await this.pc.setRemoteDescription(new RTCSessionDescription(offer));
        const answer = await this.pc.createAnswer();
        await this.pc.setLocalDescription(answer);
        this.sendMessage({
            type: 'answer',
            data: answer
        });
    }

    async handleAnswer(answer) {
        await this.pc.setRemoteDescription(new RTCSessionDescription(answer));
    }

    async handleIceCandidate(candidate) {
        await this.pc.addIceCandidate(new RTCIceCandidate(candidate));
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?role=${this.role}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            if (this.onStatusChange) {
                this.onStatusChange('signaling');
            }
        };

        this.ws.onmessage = async (event) => {
            const msg = JSON.parse(event.data);
            await this.handleMessage(msg);
        };

        this.ws.onclose = () => {
            console.log('WebSocket closed');
            if (this.onStatusChange) {
                this.onStatusChange('disconnected');
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            if (this.onStatusChange) {
                this.onStatusChange('error');
            }
        };
    }

    async handleMessage(msg) {
        switch (msg.type) {
            case 'offer':
                await this.handleOffer(msg.data);
                break;
            case 'answer':
                await this.handleAnswer(msg.data);
                break;
            case 'ice-candidate':
                await this.handleIceCandidate(msg.data);
                break;
            case 'station-ready':
                console.log('Station is ready');
                break;
            case 'viewer-ready':
                console.log('Viewer is ready');
                if (this.role === 'station' && this.pc && this.pc.signalingState === 'stable') {
                    await this.createOffer();
                }
                break;
            case 'error':
                console.error('Server error:', msg.data);
                if (this.onStatusChange) {
                    this.onStatusChange('error');
                }
                break;
        }
    }

    sendMessage(msg) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(msg));
        }
    }

    close() {
        if (this.pc) {
            this.pc.close();
        }
        if (this.ws) {
            this.ws.close();
        }
        if (this.localStream) {
            this.localStream.getTracks().forEach(track => track.stop());
        }
    }
}
