class GamePadTracker {
    constructor() {
        this.gamepadIndex = -1
    
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
            }
        }
        return null;
    }

    getCommand(myGamepad) {
        let command = [127,127,127,127];
        if(myGamepad != null) {
            document.getElementById('controllerType').innerHTML = myGamepad.id; //show gamepad type

            if(myGamepad.id.toLowerCase().includes("xbox")){
                command = this.commandFromXbox(myGamepad);
            }else if(myGamepad.id.toLowerCase().includes("g27")){
                command = this.commandFromG27(myGamepad);
            }else{
                document.getElementById('controllerType').innerHTML = "Unsupported - " + myGamepad.id;
            }
        }
        return command
    }

    mapToRange(value, min, max, minReturn, maxReturn) {
        return Math.floor((maxReturn-minReturn) * (value-min)/(max-min) + minReturn)
    }

    commandFromG27(myGamepad) {
        let command = [127,127,127,127];
        //esc
        if(myGamepad.axes[2] < .9){
            command[0] = mapToRange(myGamepad.axes[5], -1, 1, 127, 255);
        }else if(myGamepad.axes[5] < .9){
            command[0] = mapToRange(myGamepad.axes[2], -1, 1, 0, 127);
        }else{
            command[0] = 127;
        }
    
        //servo
        if(myGamepad.axes[0] > .05){
            command[1] = mapToRange(myGamepad.axes[0], .05, 1, 127, 255);
        }else if(myGamepad.axes[0] < -.05){
            command[1] = 127 - mapToRange(myGamepad.axes[0], -.05, -1, 0, 127);
        }else{
            command[1] = 127;
        }
    
        //dpad    
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
            command[0] = 127 - mapToRange(myGamepad.buttons[6].value, .1, 1, 0, 127);
        }else if(myGamepad.buttons[7].value > .1){
            //gas
            command[0] = mapToRange(myGamepad.buttons[7].value, .1, 1, 127, 255);
        }else{
            //neutral
            command[0] = 127;
        }
        //servo
        if(myGamepad.axes[0] > .1){
            command[1] = mapToRange(myGamepad.axes[0], .1, 1, 127, 255);
        }else if(myGamepad.axes[0] < -.1){
            command[1] = mapToRange(myGamepad.axes[0], -1, -.1, 0, 127);
        }else{
            command[1] = 127;
        }
    
        //pan
        if(myGamepad.axes[2] > .1){
            command[2] = mapToRange(myGamepad.axes[2], .1, 1, 127, 255);
        }else if(myGamepad.axes[2] < -.1){
            command[2] = mapToRange(myGamepad.axes[2], -1, -.1, 0, 127);
        }else{
            command[2] = 127;
        }
    
        //tilt
        if(myGamepad.axes[3] > .1){
            command[3] = mapToRange(myGamepad.axes[3], .1, 1, 127, 255);
        }else if(myGamepad.axes[3] < -.1){
            command[3] = mapToRange(myGamepad.axes[3], -1, -.1, 0, 127);
        }else{
            command[3] = 127;
        }
        return command;
    }    
}