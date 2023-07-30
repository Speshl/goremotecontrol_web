class KeyPressTracker {
    constructor() {
        this.pressedKeys = {};
        this.steeringTrim = 0;

        this.leftTrimPress = false;
        this.rightTrimPress = false;

        this.minTrim = -50;
        this.maxTrim = 50;

        // Event listener for keydown event
        document.addEventListener('keydown', (event) => {
            const key = event.key;
            this.pressedKeys[key] = true;
            if(key == "ArrowUp" || key == "ArrowDown" || key == "ArrowLeft" || key == "ArrowRight"){
                event.preventDefault();
            }
        });

        // Event listener for keyup event
        document.addEventListener('keyup', (event) => {
            const key = event.key;
            delete this.pressedKeys[key];
        });
    }

    getPressedKeys() {
        return Object.keys(this.pressedKeys);
    }

    getTrim() {
        return this.steeringTrim;
    }

    getCommand() {
        let command = [127,127,127,127];
        if(this.pressedKeys['s'] === true) {
            command[0] = 0;
        }else if(this.pressedKeys['w'] === true) {
            command[0] = 255;
        }else{
            command[0] = 127;
        }


        let steerCommand = command[1];
        if(this.pressedKeys['a'] === true) {
            steerCommand = 0;
        }else if(this.pressedKeys['d'] === true) {
            steerCommand = 255;
        }else{
            steerCommand = 127;
        }

        //steering trim
        if(this.pressedKeys['q'] && this.leftTrimPress == false){ //new press
            this.leftTrimPress = true;
            if(this.steeringTrim > this.minTrim){
                this.steeringTrim-=2;
            }
        }else if (!this.pressedKeys['q'] && this.leftTrimPress == true){
            this.leftTrimPress = false;
        }

        if(this.pressedKeys['e'] && this.rightTrimPress == false){ //new press
            this.rightTrimPress = true;
            if(this.steeringTrim < this.maxTrim){
                this.steeringTrim+=2;
            }
        }else if (!this.pressedKeys['e'] && this.rightTrimPress == true){
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


        if(this.pressedKeys['ArrowLeft'] === true) {
            command[2] = 0;
        }else if(this.pressedKeys['ArrowRight'] === true) {
            command[2] = 255;
        }else{
            command[2] = 127;
        }

        if(this.pressedKeys['ArrowUp'] === true) {
            command[3] = 255;
        }else if(this.pressedKeys['ArrowDown'] === true) {
            command[3] = 0;
        }else{
            command[3] = 127;
        }

        return command
    }
}