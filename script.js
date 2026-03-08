const ws = new WebSocket((location.protocol === "https:" ? "wss://" : "ws://") + location.host + "/ws");
const pc = new RTCPeerConnection({
    iceServers: [{urls: "stun:stun.l.google.com:19302"}]
});

const local = document.getElementById("local");
const remote = document.getElementById("remote");

const mediaReady = navigator.mediaDevices.getUserMedia({video: true, audio: true}).then(stream => {
    local.srcObject = stream;
    stream.getTracks().forEach(track => pc.addTrack(track, stream));
});

pc.ontrack = ({streams: [stream]}) => {
    remote.srcObject = stream;
};

pc.onicecandidate = ({candidate}) => {
    if (candidate) ws.send(JSON.stringify({candidate}));
};

async function handleMessage(data) {
    const message = JSON.parse(data);
    if (message.candidate) {
        pc.addIceCandidate(message.candidate);
    } else if (message.join) {
        await mediaReady;
        const offer = await pc.createOffer();
        await pc.setLocalDescription(offer);
        ws.send(JSON.stringify({description: pc.localDescription}));
    } else if (message.description && message.description.type === "offer") {
        await mediaReady;
        await pc.setRemoteDescription(message.description);
        const answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);
        ws.send(JSON.stringify({description: pc.localDescription}));
    } else if (message.description) {
        await pc.setRemoteDescription(message.description);
    }
}

let queue = Promise.resolve();
ws.onmessage = ({data}) => {
    queue = queue.then(() => handleMessage(data));
};
