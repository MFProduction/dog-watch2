class PeerConnectionManager {
    constructor(iceServers = [{ urls: 'stun:stun.l.google.com:19302' }]) {
        this.pc = null;
        this.iceServers = iceServers;
        this.onIceCandidate = null;
        this.onTrack = null;
        this.onConnectionStateChange = null;
    }

    create() {
        this.pc = new RTCPeerConnection({ iceServers: this.iceServers });

        this.pc.onicecandidate = (event) => {
            if (event.candidate && this.onIceCandidate) {
                this.onIceCandidate(event.candidate);
            }
        };

        this.pc.ontrack = (event) => {
            if (this.onTrack) {
                this.onTrack(event.streams[0]);
            }
        };

        this.pc.onconnectionstatechange = () => {
            if (this.onConnectionStateChange) {
                this.onConnectionStateChange(this.pc.connectionState);
            }
        };

        this.pc.oniceconnectionstatechange = () => {
            console.log('ICE connection state:', this.pc.iceConnectionState);
        };
    }

    addTrack(track, stream) {
        if (this.pc) {
            this.pc.addTrack(track, stream);
        }
    }

    async createOffer() {
        const offer = await this.pc.createOffer();
        await this.pc.setLocalDescription(offer);
        return offer;
    }

    async createAnswer(offer) {
        await this.pc.setRemoteDescription(new RTCSessionDescription(offer));
        const answer = await this.pc.createAnswer();
        await this.pc.setLocalDescription(answer);
        return answer;
    }

    async setRemoteAnswer(answer) {
        await this.pc.setRemoteDescription(new RTCSessionDescription(answer));
    }

    async addIceCandidate(candidate) {
        await this.pc.addIceCandidate(new RTCIceCandidate(candidate));
    }

    isStable() {
        return this.pc && this.pc.signalingState === 'stable';
    }

    close() {
        if (this.pc) {
            this.pc.close();
            this.pc = null;
        }
    }
}
