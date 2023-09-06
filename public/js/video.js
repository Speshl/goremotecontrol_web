class CamPlayer {
    constructor(forceLocal) {
        this.socket = io();

        this.lastVolume = 0;
        this.timesToShowVolume = 0;
        
        this.gotAnswer = false;

        if (forceLocal == true) {
            this.pc = new RTCPeerConnection()
        }else{
            this.pc = new RTCPeerConnection({
                iceServers: [{
                urls: 'stun:stun.l.google.com:19302'
                }]
            })
        }
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

                const canvas = document.getElementById('videoCanvas');
                canvas.addEventListener("click", () =>{
                    const canvas = document.getElementById('videoCanvas');
                    if (canvas.requestFullscreen) {
                        canvas.requestFullscreen();
                    } else if (canvas.webkitRequestFullscreen) { /* Safari */
                        canvas.webkitRequestFullscreen();
                    } else if (canvas.msRequestFullscreen) { /* IE11 */
                        canvas.msRequestFullscreen();
                    }  
                })

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
                const volumeSlider = document.getElementById('streamVolume');
                const el = document.getElementById('audioElement');
                el.srcObject = event.streams[0];
                el.autoplay = true;
                el.muted = false;
                el.playsinline = true;
                el.controls = false;
                el.volume = volumeSlider.value/100;
                this.lastVolume = volumeSlider.value/100;

                volumeSlider.addEventListener('input', (e) => {
                    el.volume = e.target.value/100;
                })
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

    showVolume(volume) {
        if(volume != this.lastVolume){
            this.lastVolume = volume;
            this.timesToShowVolume = 60;
        }
        if(this.timesToShowVolume > 0){
            this.timesToShowVolume--;
            return true;
        }
        return false;
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
}

function drawVideo() {
    const canvas = document.getElementById('videoCanvas');
    const videoContext = canvas.getContext('2d');
    const videoElement = document.getElementById('videoElement');

    const audioElement = document.getElementById('audioElement');
    let currentVolume = audioElement.volume;

    const escAndGear = document.getElementById('escAndGear').innerHTML;
    const steerAndTrim = document.getElementById('steerAndTrim').innerHTML;
    const panAndTilt = document.getElementById('panAndTilt').innerHTML;
    const combined = escAndGear + " " +steerAndTrim + " "+ panAndTilt;

    videoContext.drawImage(videoElement, 0, 0, 320,180); //TODO Make this dynamic

    videoContext.fillStyle = "white";
    videoContext.font = "10px monospace";

    if(camPlayer.showVolume(currentVolume)){
        videoContext.fillText("Volume: "+currentVolume, 140, 150)
    }

    videoContext.fillText(combined, 10, 175);
    window.requestAnimationFrame(drawVideo);
}