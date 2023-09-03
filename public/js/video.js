class CamPlayer {
    constructor() {
        this.socket = io();
        
        this.gotAnswer = false;

        this.pc = new RTCPeerConnection({
            iceServers: [{
            urls: 'stun:stun.l.google.com:19302'
            }]
        })
    }

    setupListeners() {
        this.pc.onicecandidateerror = e => {
            //log("ICE Candidate Error: "+JSON.stringify(e))
            console.log("Connection State: "+JSON.stringify(e))
            document.getElementById('statusMsg').innerHTML = "ERROR";
        }
        
        this.pc.onconnectionstatechange = e => {
            //log("Connection State: "+pc.iceConnectionState)
            console.log("Connection State: "+this.pc.iceGatheringState)
            document.getElementById('statusMsg').innerHTML = +this.pc.iceGatheringState;
        }
        
        this.pc.onicegatheringstatechange = e => {
            //log("Ice Gathering State: "+pc.iceConnectionState)
            console.log("Ice Gathering State: "+this.pc.iceGatheringState)
            //document.getElementById('statusMsg').innerHTML = +this.pc.iceGatheringState;
        }
        
        this.pc.oniceconnectionstatechange = e => {
            //log("Ice Connection State: "+pc.iceConnectionState)
            console.log("Ice Connection State: "+this.pc.iceConnectionState)
            document.getElementById('statusMsg').innerHTML = +this.pc.iceGatheringState;
        }

        this.pc.onicecandidate = event => {
            if (event.candidate === null) {
                console.log("Emmiting offer");
                this.socket.emit('offer', btoa(JSON.stringify(this.pc.localDescription)));
            } else{
                console.log("Found Candidate");
                this.socket.emit('candidate', btoa(JSON.stringify(event.candidate)));
            }
        }
        
        this.pc.ontrack = (event) => {
            if(event.track.kind == "video"){
                console.log("Creating Video Track");
                //const el = document.createElement("video");
                const el = document.getElementById('videoElement');

                el.id = "videoElement";
                el.srcObject = event.streams[0];
                el.autoplay = true;
                el.muted = true;
                el.playsinline = true;
                el.controls = true;
                //el.style.setProperty("width", "60vw");
                
                //document.getElementById('videoDiv').appendChild(el);                

                el.addEventListener("play", () => {
                    this.playMedia();
                });

                el.addEventListener("pause", () => {
                    this.pauseMedia();
                });

                el.addEventListener("volumechange", () =>{
                    const audio = document.getElementById('audioElement');
                    const video = document.getElementById('videoElement');
                    audio.volume = video.volume;
                });

               

                el.addEventListener("loadeddata", () => {
                    const canvas = document.getElementById('videoCanvas');
                    const videoElement = document.getElementById('videoElement');
                    canvas.width = videoElement.videoWidth;
                    canvas.height = videoElement.videoHeight;
                    
                    console.log("Canvas Size: ",canvas.width, canvas.height);
                    drawVideo();
                });



                console.log("Video Track Added");
            }else{
                console.log("Creating Audio Track");
                const el = document.createElement("audio");
                el.id = "audioElement";
                el.srcObject = event.streams[0];
                el.autoplay = true;
                el.muted = false;
                el.playsinline = true;
                el.controls = false;
                document.getElementById('videoDiv').appendChild(el);
                console.log("Audio Track Added");
            }
            
        }
        
        // Offer to receive 1 audio, and 1 video track
        this.pc.addTransceiver('video', {
            direction: 'recvonly'
        })
        this.pc.addTransceiver('audio', {
            direction: 'recvonly'
        })

        this.socket.on('answer', (answer) => {
            let decodedAnswer = JSON.parse(atob(answer));
            console.log("Setting Remote Description");        
            this.pc.setRemoteDescription(decodedAnswer)
                .then(() => {
                    this.gotAnswer = true;
                    console.log("Set Remote Description");
                    console.log(JSON.stringify(this.pc.remoteDescription));
                })
                .catch((error) => {
                    document.getElementById('statusMsg').innerHTML = "ERROR";
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
                document.getElementById('statusMsg').innerHTML = "ERROR";
                alert(e);
            }
        });

    }

    sendOffer() {
        document.getElementById('statusMsg').innerHTML = "Sending Offer...";
        this.pc.createOffer().then(d => this.pc.setLocalDescription(d)).catch();
    }

    sendOfferWithDelay(delay) {
        setTimeout(this.sendOffer(),delay);
    }

    async startMicrophone() {
        try{
            if(navigator.mediaDevices != null){
                const mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true });
                mediaStream.getTracks().forEach(track => this.pc.addTrack(track, mediaStream));
            }else{
                console.log("No media devices found");
            }
        }
        catch (error) {
            console.log("Error accessing microphone:", error);
        }
    }

    getSocket() {
        return this.socket;
    }

    gotRemoteDescription() {
        return this.gotAnswer;
    }

    pauseMedia() {
        console.log("Pausing...");
        const video = document.getElementById('videoElement');
        video.pause();

        const audio = document.getElementById('audioElement');
        audio.pause();
    }

    playMedia() {
        console.log("Playing...");
        const video = document.getElementById('videoElement');
        video.play();

        const audio = document.getElementById('audioElement');
        audio.play();
    }
}

function drawVideo() {
    const canvas = document.getElementById('videoCanvas');
    const videoContext = canvas.getContext('2d');
    const videoElement = document.getElementById('videoElement');

    const status = document.getElementById('statusMsg').innerHTML;
    const controllerType = document.getElementById('controllerType').innerHTML;
    const escAndGear = document.getElementById('escAndGear').innerHTML;
    const steerAndTrim = document.getElementById('steerAndTrim').innerHTML;
    const pandAndTilt = document.getElementById('panAndTilt').innerHTML;


    videoContext.drawImage(videoElement, 0, 0, 320,180); //TODO Make this dynamic

    ctx.font = "24px monospace";
    ctx.fillText("Status", 10, 160);
    ctx.fillText("Controller", 70, 160);
    ctx.fillText("Esc&Gear", 130, 160);
    ctx.fillText("Steer&Trim", 190, 160);
    ctx.fillText("Pan&Tilt", 250, 160);
    window.requestAnimationFrame(drawVideo);
}