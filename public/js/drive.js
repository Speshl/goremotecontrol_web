//Startup all the processes we need
const camPlayer = new CamPlayer();
camPlayer.setupListeners();

setTimeout(() => {
    camPlayer.startMicrophone().then(() => {
        camPlayer.sendOffer();
    });
    //camPlayer.sendOffer();
},1000);

const keyPressTracker = new KeyPressTracker();
const gamePadTracker = new GamePadTracker();

//Start listener loop for input commands
setInterval(() => {
    gamePad = gamePadTracker.getGamePad();
    trim = 0;
    
    let command = [127,0,127,127,127,0];
    if(gamePad != null){
        command = gamePadTracker.getCommand(gamePad);
        trim = gamePadTracker.getTrim();
    }else{
        command = keyPressTracker.getCommand();
        trim = keyPressTracker.getTrim();
    }

    document.getElementById('currentCommand').innerHTML = 'Esc: '+command[0] + 'Gear: ' + command[1]+' Steer: '+command[2] + ' Pan: ' + command[3] + ' Tilt: ' + command[4];
    document.getElementById('steeringTrim').innerHTML = trim;

    //Send the command we generated
    if (camPlayer.gotRemoteDescription()) {
        camPlayer.getSocket().emit('command', command);
    }
}, 10); 