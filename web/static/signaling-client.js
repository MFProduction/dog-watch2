class SignalingClient {
    constructor(role) {
        this.role = role;
        this.ws = null;
        this.onMessage = null;
        this.onOpen = null;
        this.onClose = null;
        this.onError = null;
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?role=${this.role}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            if (this.onOpen) {
                this.onOpen();
            }
        };

        this.ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            if (this.onMessage) {
                this.onMessage(msg);
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket closed');
            if (this.onClose) {
                this.onClose();
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            if (this.onError) {
                this.onError(error);
            }
        };
    }

    send(msg) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(msg));
        }
    }

    close() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}
