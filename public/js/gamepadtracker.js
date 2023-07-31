class GamePadTracker {
    constructor() {
        this.neutralCommand = [127,127,127,127];
        this.gamepadIndex = -1;
        this.steeringTrim = 0;

        this.leftTrimPress = false;
        this.rightTrimPress = false;

        this.minTrim = -50;
        this.maxTrim = 50;
    
        window.addEventListener('gamepadconnected', (event) => {
            this.gamepadIndex = event.gamepad.index;
        });

    }

    getGamePad() {
        if(this.gamepadIndex !== -1) {
            const myGamepad = navigator.getGamepads()[this.gamepadIndex];
            if(myGamepad.id.toLowerCase().includes("xbox")){
                return myGamepad;
            }else if(myGamepad.id.toLowerCase().includes("g27")){
                return myGamepad;
            }else if(myGamepad.id.toLowerCase().includes("B684")){
                return myGamepad;
            }
        }
        return null;
    }

    getCommand(myGamepad) {
        let command = this.neutralCommand;
        if(myGamepad != null) {
            document.getElementById('controllerType').innerHTML = myGamepad.id; //show gamepad type

            if(myGamepad.id.toLowerCase().includes("xbox")){
                command = this.commandFromXbox(myGamepad);
            }else if(myGamepad.id.toLowerCase().includes("g27")){
                command = this.commandFromG27(myGamepad);
            }
            else if(myGamepad.id.toLowerCase().includes("B684")){
                command = this.commandFromTGT(myGamepad);
            }else{
                document.getElementById('controllerType').innerHTML = "Unsupported - " + myGamepad.id;
            }
        }
        return command
    }

    getTrim() {
        return this.steeringTrim;
    }

    mapToRange(value, min, max, minReturn, maxReturn) {
        return Math.floor((maxReturn-minReturn) * (value-min)/(max-min) + minReturn)
    }

    commandFromG27(myGamepad) {
        let command = this.neutralCommand;
        //esc
        if(myGamepad.axes[2] < .9){
            command[0] = this.mapToRange(myGamepad.axes[5], -1, 1, 127, 255);
        }else if(myGamepad.axes[5] < .9){
            command[0] = this.mapToRange(myGamepad.axes[2], -1, 1, 0, 127);
        }else{
            command[0] = 127;
        }
    
        //servo
        let steerCommand = command[1];
        if(myGamepad.axes[0] > .05){
            steerCommand = this.mapToRange(myGamepad.axes[0], .05, 1, 127, 255);
        }else if(myGamepad.axes[0] < -.05){
            steerCommand = 127 - this.mapToRange(myGamepad.axes[0], -.05, -1, 0, 127);
        }else{
            steerCommand = 127;
        }

        //steering trim
        if(myGamepad.buttons[14] == 1.0 && this.trimLeftPress == false){ //new press
            this.trimLeftPress = true;
            this.steeringTrim--;
            if(this.steeringTrim > this.minTrim){
                this.steeringTrim--;
            }
        }else if (myGamepad.buttons[14] == 0 && this.trimLeftPress == true){
            this.trimLeftPress = false;
        }

        if(myGamepad.buttons[15] == 1.0 && this.trimRightPress == false){ //new press
            this.trimRightPress = true;
            if(this.steeringTrim < this.maxTrim){
                this.steeringTrim++;
            }
        }else if (myGamepad.buttons[15] == 0 && this.trimRightPress == true){
            this.trimRightPress = false;
        }
        
        if(steerCommand + this.steeringTrim > 255){
            steerCommand = 255;
        }else if(steerCommand + this.steeringTrim < 0) {
            steerCommand = 0;
        }else{
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
    
        if(dpadValue == upkey.toFixed(2)){
            //up
            command[3] = 255;
        }else if(dpadValue == uprightkey.toFixed(2)){
            //up-right
            command[3] = 255;
            command[2] = 255;
        }else if(dpadValue == rightkey.toFixed(2)){
            //right
            command[2] = 255;
        }else if(dpadValue == downrightkey.toFixed(2)){
            //down-right
            command[2] = 255;
            command[3] = 0;
        }else if(dpadValue == downkey.toFixed(2)){
            //down
            command[3] = 0;
        }else if(dpadValue == downleftkey.toFixed(2)){
            //down-left
            command[3] = 0;
            command[2] = 0;
        }else if(dpadValue == leftkey.toFixed(2)){
            //left
            command[2] = 0;
        }else if(dpadValue == upleftkey.toFixed(2)){
            //up-left
            command[2] = 0;
            command[3] = 255;
        }
        
        return command;
    }

    commandFromTGT(myGamepad) {
        let command = this.neutralCommand;
        //esc
        if(myGamepad.axes[2] < .9){
            command[0] = this.mapToRange(myGamepad.axes[5], -1, 1, 127, 255);
        }else if(myGamepad.axes[5] < .9){
            command[0] = this.mapToRange(myGamepad.axes[1], -1, 1, 0, 127);
        }else{
            command[0] = 127;
        }
    
        //servo
        let steerCommand = command[1];
        if(myGamepad.axes[0] > .05){
            steerCommand = this.mapToRange(myGamepad.axes[0], .05, 1, 127, 255);
        }else if(myGamepad.axes[0] < -.05){
            steerCommand = 127 - this.mapToRange(myGamepad.axes[0], -.05, -1, 0, 127);
        }else{
            steerCommand = 127;
        }

        //steering trim
        if(myGamepad.buttons[14].pressed  && this.trimLeftPress == false){ //new press
            this.trimLeftPress = true;
            this.steeringTrim--;
            if(this.steeringTrim > this.minTrim){
                this.steeringTrim--;
            }
        }else if (!myGamepad.buttons[14].pressed  && this.trimLeftPress == true){
            this.trimLeftPress = false;
        }

        if(myGamepad.buttons[3].pressed  && this.trimRightPress == false){ //new press
            this.trimRightPress = true;
            if(this.steeringTrim < this.maxTrim){
                this.steeringTrim++;
            }
        }else if (!myGamepad.buttons[15].pressed  && this.trimRightPress == true){
            this.trimRightPress = false;
        }
        
        if(steerCommand + this.steeringTrim > 255){
            steerCommand = 255;
        }else if(steerCommand + this.steeringTrim < 0) {
            steerCommand = 0;
        }else{
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
    
        if(dpadValue == upkey.toFixed(2)){
            //up
            command[3] = 255;
        }else if(dpadValue == uprightkey.toFixed(2)){
            //up-right
            command[3] = 255;
            command[2] = 255;
        }else if(dpadValue == rightkey.toFixed(2)){
            //right
            command[2] = 255;
        }else if(dpadValue == downrightkey.toFixed(2)){
            //down-right
            command[2] = 255;
            command[3] = 0;
        }else if(dpadValue == downkey.toFixed(2)){
            //down
            command[3] = 0;
        }else if(dpadValue == downleftkey.toFixed(2)){
            //down-left
            command[3] = 0;
            command[2] = 0;
        }else if(dpadValue == leftkey.toFixed(2)){
            //left
            command[2] = 0;
        }else if(dpadValue == upleftkey.toFixed(2)){
            //up-left
            command[2] = 0;
            command[3] = 255;
        }
        
        return command;
    }

    commandFromXbox(myGamepad) {
        let command = [127,127,127,127];
        //esc
        if(myGamepad.buttons[6].value > .1 &&  myGamepad.buttons[6].value >= myGamepad.buttons[7].value){
            //brake
            command[0] = 127 - this.mapToRange(myGamepad.buttons[6].value, .1, 1, 0, 127);
        }else if(myGamepad.buttons[7].value > .1){
            //gas
            command[0] = this.mapToRange(myGamepad.buttons[7].value, .1, 1, 127, 255);
        }else{
            //neutral
            command[0] = 127;
        }
        //servo
        let steerCommand = command[1];
        if(myGamepad.axes[0] > .1){
            steerCommand = this.mapToRange(myGamepad.axes[0], .1, 1, 127, 255);
        }else if(myGamepad.axes[0] < -.1){
            steerCommand = this.mapToRange(myGamepad.axes[0], -1, -.1, 0, 127);
        }else{
            steerCommand = 127;
        }

        //steering trim
        if(myGamepad.buttons[14].pressed && this.leftTrimPress == false){ //new press
            this.leftTrimPress = true;
            if(this.steeringTrim > this.minTrim){
                this.steeringTrim-=2;
            }
        }else if (!myGamepad.buttons[14].pressed && this.leftTrimPress == true){
            this.leftTrimPress = false;
        }

        if(myGamepad.buttons[15].pressed && this.rightTrimPress == false){ //new press
            this.rightTrimPress = true;
            if(this.steeringTrim < this.maxTrim){
                this.steeringTrim+=2;
            }
        }else if (!myGamepad.buttons[15].pressed && this.rightTrimPress == true){
            this.rightTrimPress = false;
        }
        
        if(steerCommand + this.steeringTrim > 255){
            steerCommand = 255;
        }else if(steerCommand + this.steeringTrim < 0) {
            steerCommand = 0;
        }else{
            steerCommand = steerCommand + this.steeringTrim;
        }
        command[1] = steerCommand;
    
        //pan
        if(myGamepad.axes[2] > .1){
            command[2] = this.mapToRange(myGamepad.axes[2], .1, 1, 127, 255);
        }else if(myGamepad.axes[2] < -.1){
            command[2] = this.mapToRange(myGamepad.axes[2], -1, -.1, 0, 127);
        }else{
            command[2] = 127;
        }
    
        //tilt
        if(myGamepad.axes[3] > .1){
            command[3] = this.mapToRange(myGamepad.axes[3], .1, 1, 127, 255);
        }else if(myGamepad.axes[3] < -.1){
            command[3] = this.mapToRange(myGamepad.axes[3], -1, -.1, 0, 127);
        }else{
            command[3] = 127;
        }
        return command;
    }    
}