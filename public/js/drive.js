//Startup all the processes we need
let forceLocalDrive = false;
const forceLocalDiv = document.getElementById('localDrive');
if(forceLocalDiv == null){
    forceLocalDrive = true;
}
const camPlayer = new CamPlayer(forceLocalDrive);
camPlayer.setupListeners();

setTimeout(() => {
    camPlayer.startMicrophone().then(() => {
        camPlayer.sendOffer();
    });
    //camPlayer.sendOffer();
}, 1000);

const keyPressTracker = new KeyPressTracker();
const gamePadTracker = new GamePadTracker();

//Start listener loop for input commands
setInterval(() => {
    let gamePad = gamePadTracker.getGamePad();
    let trim = 0;
    let gear = "N";

    let command = [127, 0, 127, 127, 127, 0];
    if (gamePad != null) {
        command = gamePadTracker.getCommand(gamePad);
        trim = gamePadTracker.getTrim();
        gear = gamePadTracker.getGearString();
    } else {
        command = keyPressTracker.getCommand();
        trim = keyPressTracker.getTrim();
        gear = keyPressTracker.getGearString();
    }

    if(gamePad != null){
        document.getElementById('controllerType').innerHTML = gamePadTracker.getControllerName(gamePad);
    }
    document.getElementById('escAndGear').innerHTML = 'Esc: ' + command[0] + ' Gear: ' + gear;
    document.getElementById('steerAndTrim').innerHTML = 'Steer: ' + command[2] + ' Trim: ' + trim;
    document.getElementById('panAndTilt').innerHTML = 'Pan: ' + command[3] + ' Tilt: ' + command[4];

    //Send the command we generated
    if (camPlayer.gotRemoteDescription()) {
        camPlayer.getSocket().emit('command', command);
    }
}, 5);