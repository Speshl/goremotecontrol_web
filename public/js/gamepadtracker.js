class GamePadTracker {
    constructor() {
        this.maxPosition = 255;
        this.midPosition = 127;
        this.minPosition = 0;

        this.panSpeed = 2;
        this.tiltSpeed = 2;

        this.neutralGear = 0;
        this.reverseGear = 255;
        this.maxGears = 6;

        this.neutralCommand = [this.midPosition, this.neutralGear, this.midPosition, this.midPosition, this.midPosition, 0];
        this.panPos = this.midPosition;
        this.tiltPos = this.midPosition;
        this.currentGear = this.neutralGear;
        this.inManualGear = false;

        this.gamepadIndex = -1;
        this.steeringTrim = 0;

        this.volumeUpPress = false;
        this.volumeDownPress = false;
        this.volumeMutePress = false;

        this.volumeMuted = false;
        this.volumeAtMute = 0;

        this.upShiftPress = false;
        this.downShiftPress = false;

        this.leftTrimPress = false;
        this.rightTrimPress = false;

        this.minTrim = -50;
        this.maxTrim = 50;

        window.addEventListener('gamepadconnected', (event) => {
            const myGamepads = navigator.getGamepads();
            if (myGamepads != null && myGamepads[event.gamepad.index] != null) {
                this.gamepadIndex = event.gamepad.index;
            } else {
                console.log("Got event from null gamepad: ", event.gamepad.index);
            }
        });

        window.addEventListener('gamepaddisconnected', (evnet) => {
            this.gamepadIndex = -1;
        });

    }

    getGamePad() {
        if (this.gamepadIndex !== -1) {
            const myGamepads = navigator.getGamepads();
            const myGamepad = myGamepads[this.gamepadIndex];

            // console.log("Gamepad Index: ", this.gamepadIndex);
            // console.log("Number Controllers Found: ", myGamepads.length);
            // for(let i=0; i<myGamepads.length; i++){
            //     if(myGamepads[i] != null){
            //         console.log("GamePad ID: ",myGamepads[i].id);
            //     }else{
            //         console.log("Game pad index" + i + " is null");
            //     }
            // }

            if (myGamepad.id.toLowerCase().includes("xbox")) {
                return myGamepad;
            } else if (myGamepad.id.toLowerCase().includes("g27")) {
                return myGamepad;
            } else if (myGamepad.id.toLowerCase().includes("b684")) { //TGT wheel
                return myGamepad;
            }
        }
        return null;
    }

    getCommand(myGamepad) {
        let command = this.neutralCommand;
        if (myGamepad != null) {
            if (myGamepad.id.toLowerCase().includes("xbox")) {
                command = this.commandFromXbox(myGamepad);
            }
            // else if(myGamepad.id.toLowerCase().includes("g27")){
            //     command = this.commandFromG27(myGamepad);
            // }else if(myGamepad.id.toLowerCase().includes("b684")){ //TGT Wheel
            //     command = this.commandFromTGT(myGamepad);
            // }
            else {
                document.getElementById('controllerType').innerHTML = "Unsupported - " + myGamepad.id;
            }
        }
        return command
    }

    getTrim() {
        return this.steeringTrim;
    }

    getGearString() {
        if(this.currentGear == this.neutralGear){
            return "N";
        }else if(this.currentGear == this.reverseGear){
            return "R";
        }else{
            return ""+this.currentGear;
        }
    }

    getControllerName(myGamepad){
        if (myGamepad.id.toLowerCase().includes("xbox")) {
            return "Xbox"
        } else if (myGamepad.id.toLowerCase().includes("g27")) {
            return "G27"
        } else if (myGamepad.id.toLowerCase().includes("b684")) { //TGT wheel
            return "TGT"
        }else{
            return "Unsupported"
        }
    }

    mapToRange(value, min, max, minReturn, maxReturn) {
        return Math.floor((maxReturn - minReturn) * (value - min) / (max - min) + minReturn)
    }

    upShift() {
        if(this.currentGear == this.reverseGear){
            this.currentGear = this.neutralGear;
        }else if(this.currentGear == this.neutralGear){
            this.currentGear = 1;
        }else if(this.currentGear >=0 && this.currentGear <this.maxGears){
            this.currentGear++;
        }
    }

    downShift() {
        if(this.currentGear == this.neutralGear){
            this.currentGear = this.reverseGear;
        }else if(this.currentGear == 1){
            this.currentGear = this.neutralGear;
        }
        else if(this.currentGear > 1 && this.currentGear <= this.maxGears){
            this.currentGear--;
        }
    }

    volumeUp() {
        const volumeSlider = document.getElementById('streamVolume');
        const audioElement = document.getElementById('audioElement');

        if(this.volumeMuted){
            this.volumeUnMute();
        }

        let currentVolume = volumeSlider.value;
        let newVolume = parseInt(currentVolume) + 10;

        if (newVolume < 100) {
            audioElement.volume = newVolume/100;
            volumeSlider.value = newVolume;
        }else{
            audioElement.volume = 1;
            volumeSlider.value = 100;
        }  
    }

    volumeDown() {
        const volumeSlider = document.getElementById('streamVolume');
        const audioElement = document.getElementById('audioElement');
        let currentVolume = volumeSlider.value;
        let newVolume = parseInt(currentVolume) - 10;

        if (newVolume > 0) {
            audioElement.volume = newVolume/100;
            volumeSlider.value = newVolume;
        }else{
            audioElement.volume = 0;
            volumeSlider.value = 0;
        }        
    }

    volumeMute() {
        const volumeSlider = document.getElementById('streamVolume');
        const audioElement = document.getElementById('audioElement');
        this.volumeAtMute = volumeSlider.value;
        audioElement.volume = 0;
        volumeSlider.value = 0;
        this.volumeMuted = true;
    }

    volumeUnMute() {
        const volumeSlider = document.getElementById('streamVolume');
        const audioElement = document.getElementById('audioElement');
        audioElement.volume = this.volumeAtMute / 100;
        volumeSlider.value = this.volumeAtMute;
        this.volumeAtMute = 0;
        this.volumeMuted = false;
    }

    commandFromXbox(myGamepad) {
        //DPAD UP/Down/Left/Right
        // 12/13/14/15
        let command = this.neutralCommand;
        //esc
        if (myGamepad.buttons[6].value > .1 && myGamepad.buttons[6].value >= myGamepad.buttons[7].value) {
            //brake
            command[0] = this.midPosition - this.mapToRange(myGamepad.buttons[6].value, .1, 1, this.minPosition, this.midPosition);
        } else if (myGamepad.buttons[7].value > .1) {
            //gas
            command[0] = this.mapToRange(myGamepad.buttons[7].value, .1, 1, this.midPosition, this.maxPosition);
        } else {
            //neutral
            command[0] = this.midPosition;
        }

         //Voume Up
         if(myGamepad.buttons[12].pressed && this.volumeUpPress == false){ //new press
            this.volumeUpPress = true;
            this.volumeUp();
            
        }else if (!myGamepad.buttons[12].pressed && this.volumeUpPress == true){
            this.volumeUpPress = false;
        }

        //Voume Down
        if(myGamepad.buttons[13].pressed && this.volumeDownPress == false){ //new press
            this.volumeDownPress = true;
            this.volumeDown();
            
        }else if (!myGamepad.buttons[13].pressed && this.volumeDownPress == true){
            this.volumeDownPress = false;
        }

        //Voume Mute
        if(myGamepad.buttons[9].pressed && this.volumeMutePress == false){ //new press
            this.volumeMutePress = true;
            if(this.volumeMuted){
                this.volumeUnMute();
            }else{
                this.volumeMute();
            } 
        }else if (!myGamepad.buttons[9].pressed && this.volumeMutePress == true){
            this.volumeMutePress = false;
        }

        //Upshift
        if (myGamepad.buttons[5].pressed && this.upShiftPress == false) { //new press
            this.upShiftPress = true;
            this.upShift();
        } else if (!myGamepad.buttons[5].pressed && this.upShiftPress == true) {
            this.upShiftPress = false;
        }

        //Downshift
        if (myGamepad.buttons[4].pressed && this.downShiftPress == false) { //new press
            this.downShiftPress = true;
            this.downShift();
        } else if (!myGamepad.buttons[4].pressed && this.downShiftPress == true) {
            this.downShiftPress = false;
        }

        command[1] = this.currentGear;

        //servo
        let steerCommand = command[1];
        if (myGamepad.axes[0] > .1) {
            steerCommand = this.mapToRange(myGamepad.axes[0], .1, 1, this.midPosition, this.maxPosition);
        } else if (myGamepad.axes[0] < -.1) {
            steerCommand = this.mapToRange(myGamepad.axes[0], -1, -.1, this.minPosition, this.midPosition);
        } else {
            steerCommand = this.midPosition;
        }

        //steering trim
        if (myGamepad.buttons[14].pressed && this.leftTrimPress == false) { //new press
            this.leftTrimPress = true;
            if (this.steeringTrim > this.minTrim) {
                this.steeringTrim -= 2;
            }
        } else if (!myGamepad.buttons[14].pressed && this.leftTrimPress == true) {
            this.leftTrimPress = false;
        }

        if (myGamepad.buttons[15].pressed && this.rightTrimPress == false) { //new press
            this.rightTrimPress = true;
            if (this.steeringTrim < this.maxTrim) {
                this.steeringTrim += 2;
            }
        } else if (!myGamepad.buttons[15].pressed && this.rightTrimPress == true) {
            this.rightTrimPress = false;
        }

        if (steerCommand + this.steeringTrim > this.maxPosition) {
            steerCommand = this.maxPosition;
        } else if (steerCommand + this.steeringTrim < 0) {
            steerCommand = 0;
        } else {
            steerCommand = steerCommand + this.steeringTrim;
        }
        command[2] = steerCommand;

        if (myGamepad.axes[2] > .15 || myGamepad.axes[2] < -.15) {
            this.panPos += this.mapToRange(myGamepad.axes[2], -1, 1, -1 * this.panSpeed, this.panSpeed);
        }
        if (myGamepad.axes[3] > .15 || myGamepad.axes[3] < -.15) {
            this.tiltPos -= this.mapToRange(myGamepad.axes[3], -1, 1, -1 * this.tiltSpeed, this.tiltSpeed);
        }

        if (this.panPos < this.minPosition) {
            this.panPos = this.minPosition;
        }
        if (this.panPos > this.maxPosition) {
            this.panPos = this.maxPosition;
        }

        if (this.tiltPos < this.minPosition) {
            this.tiltPos = this.minPosition;
        }
        if (this.tiltPos > this.maxPosition) {
            this.tiltPos = this.maxPosition;
        }


        //Reset camera
        if (myGamepad.buttons[11].pressed) {
            this.panPos = this.midPosition;
            this.tiltPos = this.midPosition;
        }
        command[3] = this.panPos;
        command[4] = this.tiltPos;

        //Quick Sounds
        if (myGamepad.buttons[0].pressed) {
            command[5] = 1;
        } else if (myGamepad.buttons[1].pressed) {
            command[5] = 2;
        } else if (myGamepad.buttons[2].pressed) {
            command[5] = 3;
        } else if (myGamepad.buttons[3].pressed) {
            command[5] = 4;
        } else {
            command[5] = 0;
        }

        return command;
    }






    //TODO Update for gears
    commandFromG27(myGamepad) {
        let command = this.neutralCommand;
        //esc
        if (myGamepad.axes[2] < .9) {
            command[0] = this.mapToRange(myGamepad.axes[5], -1, 1, this.midPosition, this.maxPosition);
        } else if (myGamepad.axes[5] < .9) {
            command[0] = this.mapToRange(myGamepad.axes[2], -1, 1, this.minPosition, this.midPosition);
        } else {
            command[0] = this.midPosition;
        }

        //gear
        //TODO: Update gear buttons And test manual gear shifting
        //Upshift
        if (myGamepad.buttons[5].pressed && this.upShiftPress == false) { //new press
            this.upShiftPress = true;
            this.upShift();
        } else if (!myGamepad.buttons[5].pressed && this.upShiftPress == true) {
            this.upShiftPress = false;
        }

        //Downshift
        if (myGamepad.buttons[4].pressed && this.downShiftPress == false) { //new press
            this.downShiftPress = true;
            this.downShift();
        } else if (!myGamepad.buttons[4].pressed && this.downShiftPress == true) {
            this.downShiftPress = false;
        }

        let inManualGear = false;
        if (myGamepad.buttons[0].pressed){
            this.currentGear = this.reverseGear
            inManualGear = true;
        }else if (myGamepad.buttons[0].pressed){
            this.currentGear = 1
            inManualGear = true;
        }else if (myGamepad.buttons[0].pressed){
            this.currentGear = 2
            inManualGear = true;
        }else if (myGamepad.buttons[0].pressed){
            this.currentGear = 3
            inManualGear = true;
        }else if (myGamepad.buttons[0].pressed){
            this.currentGear = 4
            inManualGear = true;
        }else if (myGamepad.buttons[0].pressed){
            this.currentGear = 5
            inManualGear = true;
        }else if (myGamepad.buttons[0].pressed){
            this.currentGear = 6
            inManualGear = true;
        }

        if(inManualGear){
            this.inManualGear = true;
        }

        if (!inManualGear && this.inManualGear){
            this.inManualGear = false;
            this.currentGear = this.neutralGear;
        }

        command[1] = this.currentGear;

        //servo
        let steerCommand = command[1];
        if (myGamepad.axes[0] > .05) {
            steerCommand = this.mapToRange(myGamepad.axes[0], .05, 1, this.midPosition, this.maxPosition);
        } else if (myGamepad.axes[0] < -.05) {
            steerCommand = this.midPosition - this.mapToRange(myGamepad.axes[0], -.05, -1, this.minPosition, this.midPosition);
        } else {
            steerCommand = this.midPosition;
        }

        //steering trim
        if (myGamepad.buttons[14] == 1.0 && this.trimLeftPress == false) { //new press
            this.trimLeftPress = true;
            this.steeringTrim--;
            if (this.steeringTrim > this.minTrim) {
                this.steeringTrim--;
            }
        } else if (myGamepad.buttons[14] == 0 && this.trimLeftPress == true) {
            this.trimLeftPress = false;
        }

        if (myGamepad.buttons[15] == 1.0 && this.trimRightPress == false) { //new press
            this.trimRightPress = true;
            if (this.steeringTrim < this.maxTrim) {
                this.steeringTrim++;
            }
        } else if (myGamepad.buttons[15] == 0 && this.trimRightPress == true) {
            this.trimRightPress = false;
        }

        if (steerCommand + this.steeringTrim > this.maxPosition) {
            steerCommand = this.maxPosition;
        } else if (steerCommand + this.steeringTrim < 0) {
            steerCommand = 0;
        } else {
            steerCommand = steerCommand + this.steeringTrim;
        }
        command[1] = steerCommand;

        //dpad    Pan/Tilt
        dpadValue = myGamepad.axes[9].toFixed(2)
        const upkey = -1.00
        const uprightkey = -0.71429
        const rightkey = -0.42857
        const downrightkey = -0.14286
        const downkey = 0.14286
        const downleftkey = 0.42857
        const leftkey = 0.71429
        const upleftkey = 1.00

        if (dpadValue == upkey.toFixed(2)) {
            //up
            this.tiltPos += this.tiltSpeed;
        } else if (dpadValue == uprightkey.toFixed(2)) {
            //up-right
            this.tiltPos += this.tiltSpeed;
            this.panPos += this.panSpeed;
        } else if (dpadValue == rightkey.toFixed(2)) {
            //right
            this.panPos += this.panSpeed;
        } else if (dpadValue == downrightkey.toFixed(2)) {
            //down-right
            this.panPos += this.panSpeed;
            this.tiltPos -= this.tiltSpeed;
        } else if (dpadValue == downkey.toFixed(2)) {
            //down
            this.tiltPos -= this.tiltSpeed;
        } else if (dpadValue == downleftkey.toFixed(2)) {
            //down-left
            this.panPos -= this.panSpeed;
            this.tiltPos -= this.tiltSpeed;
        } else if (dpadValue == leftkey.toFixed(2)) {
            //left
            this.panPos -= this.panSpeed;
        } else if (dpadValue == upleftkey.toFixed(2)) {
            //up-left
            this.panPos -= this.panSpeed;
            this.tiltPos += this.tiltSpeed;
        }

        if (this.panPos < this.minPosition) {
            this.panPos = this.minPosition;
        }
        if (this.panPos > this.maxPosition) {
            this.panPos = this.maxPosition;
        }

        if (this.tiltPos < this.minPosition) {
            this.tiltPos = this.minPosition;
        }
        if (this.tiltPos > this.maxPosition) {
            this.tiltPos = this.maxPosition;
        }

        //Reset camera
        if (myGamepad.buttons[11].pressed) { //TODO: Need to figure out button for this
            this.panPos = this.midPosition;
            this.tiltPos = this.midPosition;
        }
        command[2] = this.panPos;
        command[3] = this.tiltPos;
        return command;
    }
    
    //TODO Update for gears
    commandFromTGT(myGamepad) {
        let command = this.neutralCommand;
        //esc
        if (myGamepad.axes[5] < .9) {
            command[0] = this.mapToRange(myGamepad.axes[5], -1, 1, this.midPosition, this.maxPosition);
        } else if (myGamepad.axes[1] < .9) {
            command[0] = this.mapToRange(myGamepad.axes[1], -1, 1, this.minPosition, this.midPosition);
        } else {
            command[0] = this.midPosition;
        }

        //servo
        let steerCommand = command[1];
        if (myGamepad.axes[0] > .05) {
            steerCommand = this.mapToRange(myGamepad.axes[0], .05, 1, this.midPosition, this.maxPosition);
        } else if (myGamepad.axes[0] < -.05) {
            steerCommand = this.midPosition - this.mapToRange(myGamepad.axes[0], -.05, -1, this.minPosition, this.midPosition);
        } else {
            steerCommand = this.midPosition;
        }

        //steering trim
        if (myGamepad.buttons[3].pressed && this.trimLeftPress == false) { //new press
            this.trimLeftPress = true;
            this.steeringTrim--;
            if (this.steeringTrim > this.minTrim) {
                this.steeringTrim--;
            }
        } else if (!myGamepad.buttons[3].pressed && this.trimLeftPress == true) {
            this.trimLeftPress = false;
        }

        if (myGamepad.buttons[4].pressed && this.trimRightPress == false) { //new press
            this.trimRightPress = true;
            if (this.steeringTrim < this.maxTrim) {
                this.steeringTrim++;
            }
        } else if (!myGamepad.buttons[4].pressed && this.trimRightPress == true) {
            this.trimRightPress = false;
        }

        if (steerCommand + this.steeringTrim > this.maxPosition) {
            steerCommand = this.maxPosition;
        } else if (steerCommand + this.steeringTrim < 0) {
            steerCommand = 0;
        } else {
            steerCommand = steerCommand + this.steeringTrim;
        }
        command[1] = steerCommand;

        // //dpad    Pan/Tilt
        // dpadValue = myGamepad.axes[9].toFixed(2)
        // const upkey = -1.00
        // const uprightkey = -0.71429
        // const rightkey = -0.42857
        // const downrightkey = -0.14286
        // const downkey = 0.14286
        // const downleftkey = 0.42857
        // const leftkey = 0.71429
        // const upleftkey = 1.00

        // if(dpadValue == upkey.toFixed(2)){
        //     //up
        //     command[3] = 255;
        // }else if(dpadValue == uprightkey.toFixed(2)){
        //     //up-right
        //     command[3] = 255;
        //     command[2] = 255;
        // }else if(dpadValue == rightkey.toFixed(2)){
        //     //right
        //     command[2] = 255;
        // }else if(dpadValue == downrightkey.toFixed(2)){
        //     //down-right
        //     command[2] = 255;
        //     command[3] = 0;
        // }else if(dpadValue == downkey.toFixed(2)){
        //     //down
        //     command[3] = 0;
        // }else if(dpadValue == downleftkey.toFixed(2)){
        //     //down-left
        //     command[3] = 0;
        //     command[2] = 0;
        // }else if(dpadValue == leftkey.toFixed(2)){
        //     //left
        //     command[2] = 0;
        // }else if(dpadValue == upleftkey.toFixed(2)){
        //     //up-left
        //     command[2] = 0;
        //     command[3] = 255;
        // }

        return command;
    }
}