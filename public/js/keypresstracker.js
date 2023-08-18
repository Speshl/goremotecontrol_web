class KeyPressTracker {
    constructor() {
        this.maxPosition = 255;
        this.midPosition = 127;
        this.minPosition = 0;

        this.panSpeed = 5;
        this.tiltSpeed = 5;

        this.neutralCommand = [this.midPosition,this.midPosition,this.midPosition,this.midPosition,0];
        this.panPos = this.midPosition;
        this.tiltPos = this.midPosition;

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
        let command = this.neutralCommand;
        if(this.pressedKeys['s'] === true) {
            command[0] = this.minPosition;
        }else if(this.pressedKeys['w'] === true) {
            command[0] = this.maxPosition;
        }else{
            command[0] = this.midPosition;
        }


        let steerCommand = command[1];
        if(this.pressedKeys['a'] === true) {
            steerCommand = this.minPosition;
        }else if(this.pressedKeys['d'] === true) {
            steerCommand = this.maxPosition;
        }else{
            steerCommand = this.midPosition;
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
        
        if(steerCommand + this.steeringTrim > this.maxPosition){
            steerCommand = this.maxPosition;
        }else if(steerCommand + this.steeringTrim < this.minPosition) {
            steerCommand = this.minPosition;
        }else{
            steerCommand = steerCommand + this.steeringTrim;
        }
        command[1] = steerCommand;


        if(this.pressedKeys['ArrowLeft'] === true) {
            this.panPos -= this.panSpeed;
            if(this.panPos < this.minPosition){
                this.panPos = this.minPosition
            }
        }else if(this.pressedKeys['ArrowRight'] === true) {
            this.panPos += this.panSpeed;
            if(this.panPos > this.maxPosition){
                this.panPos = this.maxPosition
            }
        }else{
            this.panPos = this.midPosition
        }

        if(this.pressedKeys['ArrowDown'] === true) {
            this.tiltPos -= this.tiltSpeed;
            if(this.tiltPos < this.minPosition){
                this.tiltPos = this.minPosition
            }
        }else if(this.pressedKeys['ArrowUp'] === true) {
            this.tiltPos += this.tiltSpeed;
            if(this.tiltPos > this.maxPosition){
                this.tiltPos = this.maxPosition
            }
        }else{
            this.tiltPos = this.midPosition
        }

        //Resent camera on spacebar
        if(this.pressedKeys['Space'] === true) {
            this.tiltPos = this.midPosition;
            this.panPos = this.midPosition;
        }

        command[2] = this.panPos;
        command[3] = this.tiltPos;



        return command
    }
}