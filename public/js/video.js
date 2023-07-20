class CamPlayer {
    constructor() {
        this.socket = io();

        this.pc = new RTCPeerConnection({
            iceServers: [{
            urls: 'stun:stun.l.google.com:19302'
            }]
        })

        this.pc.onicecandidateerror = e => {
            //log("ICE Candidate Error: "+JSON.stringify(e))
            console.log("Connection State: "+JSON.stringify(e))
        }
        
        this.pc.onconnectionstatechange = e => {
            //log("Connection State: "+pc.iceConnectionState)
            console.log("Connection State: "+this.pc.iceGatheringState)
        }
        
        this.pc.onicegatheringstatechange = e => {
            //log("Ice Gathering State: "+pc.iceConnectionState)
            console.log("Ice Gathering State: "+this.pc.iceGatheringState)
        }
        
        this.pc.oniceconnectionstatechange = e => {
            //log("Ice Connection State: "+pc.iceConnectionState)
            console.log("Ice Connection State: "+this.pc.iceConnectionState)
        }

        this.pc.onicecandidate = event => {
            if (event.candidate === null) {
                console.log("Emmiting offer");
                this.socket.emit('offer', btoa(JSON.stringify(this.pc.localDescription)));
            }else{
                console.log("Found Candidate");
                this.socket.emit('candidate', btoa(JSON.stringify(event.candidate)));
            }
        }
        
        this.pc.ontrack = (event) => {
            console.log("Track Added");
            const el = document.createElement(event.track.kind);
            el.srcObject = event.streams[0];
            el.autoplay = true;
            el.controls = true;
            document.getElementById('videoDiv').appendChild(el);
        }
        
        //Offer to receive 1 audio, and 1 video track
        this.pc.addTransceiver('video', {
            direction: 'recvonly'
        })
        // this.pc.addTransceiver('audio', {
        //     direction: 'recvonly'
        // })
        
        this.pc.createOffer().then(d => this.pc.setLocalDescription(d)).catch(log)

        this.socket.on('answer', (answer) => {
            let decodedAnswer = JSON.parse(atob(answer));
            console.log("Setting Remote Description");        
            this.pc.setRemoteDescription(decodedAnswer)
                .then(() => {
                    console.log("Set Remote Description");
                    console.log(JSON.stringify(this.pc.remoteDescription));
                })
                .catch((error) => {
                    console.error("Error setting remote description:", error);
                    alert("Error setting remote description: " + error.message);
                });
        });

        this.socket.on('candidate', async(candidate) => {
            try {
                setTimeout(async() => {
                    const decodedCandidate = JSON.parse(atob(candidate));
                    console.log(JSON.stringify(decodedCandidate))
                    await this.pc.addIceCandidate(decodedCandidate);
                    console.log("Added ICE candidate");
                }, 1000);
            } catch (e) {
                alert(e);
            }
        });
    }

    getSocket() {
        return this.socket
    }
}