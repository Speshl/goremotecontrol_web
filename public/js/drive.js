const log = msg => {
    document.getElementById('div').innerHTML += msg + '<br>'
}

//Startup all the processes we need
const camPlayer = new CamPlayer();
const keyPressTracker = new KeyPressTracker();
const gamePadTracker = new GamePadTracker();

//Start listener loop for input commands
setInterval(() => {
    gamePad = gamePadTracker.getGamePad()
    if(gamePad != null){
        command = gamePadTracker.getCommand(gamePad)
    }else{
        command = keyPressTracker.getCommand();
    }
    document.getElementById('currentCommand').innerHTML = 'Esc: '+command[0] + ' Servo: '+command[1] + ' Pan: ' + command[2] + ' Tilt: ' + command[3];
        //Send the command we generated
        if (camPlayer.gotRemoteDescription()) {
            camPlayer.getSocket().emit('command', command);
        }
}, 10); 