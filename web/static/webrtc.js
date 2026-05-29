class WebRTCClient {
    constructor(role) {
        this.role = role;
        this.peerConnection = new PeerConnectionManager();
        this.signaling = new SignalingClient(role);
        this.media = new MediaManager();
        
        this.onRemoteStream = null;
        this.onStatusChange = null;

        this.setupPeerConnectionCallbacks();
        this.setupSignalingCallbacks();
    }

    setupPeerConnectionCallbacks() {
        this.peerConnection.onIceCandidate = (candidate) => {
            this.signaling.send({
                type: 'ice-candidate',
                data: candidate
            });
        };

        this.peerConnection.onTrack = (stream) => {
            this.media.setRemoteStream(stream);
            if (this.onRemoteStream) {
                this.onRemoteStream(stream);
            }
        };

        this.peerConnection.onConnectionStateChange = (state) => {
            if (this.onStatusChange) {
                this.onStatusChange(state);
            }
        };
    }

    setupSignalingCallbacks() {
        this.signaling.onOpen = () => {
            if (this.onStatusChange) {
                this.onStatusChange('signaling');
            }
        };

        this.signaling.onMessage = async (msg) => {
            await this.handleMessage(msg);
        };

        this.signaling.onClose = () => {
            if (this.onStatusChange) {
                this.onStatusChange('disconnected');
            }
        };

        this.signaling.onError = () => {
            if (this.onStatusChange) {
                this.onStatusChange('error');
            }
        };
    }

    createPeerConnection() {
        this.peerConnection.create();
    }

    addLocalStream(stream) {
        this.media.localStream = stream;
        stream.getTracks().forEach(track => {
            this.peerConnection.addTrack(track, stream);
        });
    }

    connectWebSocket() {
        this.signaling.connect();
    }

    async handleMessage(msg) {
        switch (msg.type) {
            case 'offer':
                const answer = await this.peerConnection.createAnswer(msg.data);
                this.signaling.send({
                    type: 'answer',
                    data: answer
                });
                break;
            case 'answer':
                await this.peerConnection.setRemoteAnswer(msg.data);
                break;
            case 'ice-candidate':
                await this.peerConnection.addIceCandidate(msg.data);
                break;
            case 'station-ready':
                console.log('Station is ready');
                break;
            case 'viewer-ready':
                console.log('Viewer is ready');
                if (this.role === 'station' && this.peerConnection.isStable()) {
                    const offer = await this.peerConnection.createOffer();
                    this.signaling.send({
                        type: 'offer',
                        data: offer
                    });
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

    close() {
        this.peerConnection.close();
        this.signaling.close();
        this.media.close();
    }
}
