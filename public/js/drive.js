//Startup all the processes we need
const camPlayer = new CamPlayer();
const keyPressTracker = new KeyPressTracker();
const gamePadTracker = new GamePadTracker();

//Start listener loop for input commands
setInterval(() => {
    gamePad = gamePadTracker.getGamePad();
    trim = 0;
    
    let command = [127,127,127,127];
    if(gamePad != null){
        command = gamePadTracker.getCommand(gamePad);
        trim = gamePadTracker.getTrim();
    }else{
        command = keyPressTracker.getCommand();
        trim = keyPressTracker.getTrim();
    }

    document.getElementById('currentCommand').innerHTML = 'Esc: '+command[0] + ' Servo: '+command[1]/* + ' Pan: ' + command[2] + ' Tilt: ' + command[3]*/;
    document.getElementById('steeringTrim').innerHTML = trim;

    //Send the command we generated
    if (camPlayer.gotRemoteDescription()) {
        camPlayer.getSocket().emit('command', command);
    }
}, 10); 