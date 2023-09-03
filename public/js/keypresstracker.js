class KeyPressTracker {
    constructor() {
        this.maxPosition = 255;
        this.midPosition = 127;
        this.minPosition = 0;

        this.panSpeed = 1;
        this.tiltSpeed = 1;

        this.neutralGear = 0;
        this.reverseGear = 255;
        this.maxGears = 6;

        this.neutralCommand = [this.midPosition,this.neutralGear,this.midPosition,this.midPosition,this.midPosition,0];
        this.panPos = this.midPosition;
        this.tiltPos = this.midPosition;
        this.currentGear = this.neutralGear;

        this.pressedKeys = {};
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

        // Event listener for keydown event
        document.addEventListener('keydown', (event) => {
            const key = event.key;
            this.pressedKeys[key] = true;
            if(key == "ArrowUp" || key == "ArrowDown" || key == "ArrowLeft" || key == "ArrowRight" || key == " "){
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

    getGearString() {
        if(this.currentGear == this.neutralGear){
            return "N";
        }else if(this.currentGear == this.reverseGear){
            return "R";
        }else{
            return ""+this.currentGear;
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

    volumeSync() {
        const volumeSlider = document.getElementById('streamVolume');
        if(this.volumeMuted && volumeSlider.value > 0) {
            this.volumeMuted = false;
        }
    }

    upShift() {
        if(this.currentGear == this.reverseGear){
            this.currentGear = this.neutralGear;
        }else if(this.currentGear == this.neutralGear){
            this.currentGear = 1;
        }else if(this.currentGear >=0 && this.currentGear <this.maxGears){
            this.currentGear ++;
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

    getCommand() {
        let command = this.neutralCommand;
        if(this.pressedKeys['s'] === true) {
            command[0] = this.minPosition;
        }else if(this.pressedKeys['w'] === true) {
            command[0] = this.maxPosition;
        }else{
            command[0] = this.midPosition;
        }


        let steerCommand = command[2];
        if(this.pressedKeys['a'] === true) {
            steerCommand = this.minPosition;
        }else if(this.pressedKeys['d'] === true) {
            steerCommand = this.maxPosition;
        }else{
            steerCommand = this.midPosition;
        }

        this.volumeSync();

        //Voume Up
        if(this.pressedKeys[']'] && this.volumeUpPress == false){ //new press
            this.volumeUpPress = true;
            this.volumeUp();
            
        }else if (!this.pressedKeys[']'] && this.volumeUpPress == true){
            this.volumeUpPress = false;
        }

        //Voume Down
        if(this.pressedKeys['['] && this.volumeDownPress == false){ //new press
            this.volumeDownPress = true;
            this.volumeDown();
            
        }else if (!this.pressedKeys['['] && this.volumeDownPress == true){
            this.volumeDownPress = false;
        }

        //Voume Mute
        if(this.pressedKeys['m'] && this.volumeMutePress == false){ //new press
            this.volumeMutePress = true;
            if(this.volumeMuted){
                this.volumeUnMute();
            }else{
                this.volumeMute();
            } 
        }else if (!this.pressedKeys['m'] && this.volumeMutePress == true){
            this.volumeMutePress = false;
        }

        //Upshift
        if(this.pressedKeys['e'] && this.upShiftPress == false){ //new press
            this.upShiftPress = true;
            this.upShift();
            
        }else if (!this.pressedKeys['e'] && this.upShiftPress == true){
            this.upShiftPress = false;
        }

         //Downshift
         if(this.pressedKeys['q'] && this.downShiftPress == false){ //new press
            this.downShiftPress = true;
            this.downShift();
        }else if (!this.pressedKeys['q'] && this.downShiftPress == true){
            this.downShiftPress = false;
        }

        command[1] = this.currentGear;

        //steering trim
        if(this.pressedKeys[','] && this.leftTrimPress == false){ //new press
            this.leftTrimPress = true;
            if(this.steeringTrim > this.minTrim){
                this.steeringTrim-=2;
            }
        }else if (!this.pressedKeys[','] && this.leftTrimPress == true){
            this.leftTrimPress = false;
        }

        if(this.pressedKeys['.'] && this.rightTrimPress == false){ //new press
            this.rightTrimPress = true;
            if(this.steeringTrim < this.maxTrim){
                this.steeringTrim+=2;
            }
        }else if (!this.pressedKeys['.'] && this.rightTrimPress == true){
            this.rightTrimPress = false;
        }
        
        if(steerCommand + this.steeringTrim > this.maxPosition){
            steerCommand = this.maxPosition;
        }else if(steerCommand + this.steeringTrim < this.minPosition) {
            steerCommand = this.minPosition;
        }else{
            steerCommand = steerCommand + this.steeringTrim;
        }
        command[2] = steerCommand;


        if(this.pressedKeys['ArrowLeft'] === true) {
            this.panPos -= this.panSpeed;
            if(this.panPos < this.minPosition){
                this.panPos = this.minPosition;
            }
        }else if(this.pressedKeys['ArrowRight'] === true) {
            this.panPos += this.panSpeed;
            if(this.panPos > this.maxPosition){
                this.panPos = this.maxPosition;
            }
        }else{
            //auto recenter
            // if(this.panPos > this.midPosition){
            //     this.panPos -= this.panSpeed;
            // }
            // if(this.panPos < this.midPosition){
            //     this.panPos += this.panSpeed;
            // }
            
            // let diffrence = this.panPos - this.midPosition
            // if(Math.abs(diffrence) > this.panSpeed){
            //     this.panPos = this.midPosition
            // }
        }

        if(this.pressedKeys['ArrowDown'] === true) {
            this.tiltPos -= this.tiltSpeed;
            if(this.tiltPos < this.minPosition){
                this.tiltPos = this.minPosition;
            }
        }else if(this.pressedKeys['ArrowUp'] === true) {
            this.tiltPos += this.tiltSpeed;
            if(this.tiltPos > this.maxPosition){
                this.tiltPos = this.maxPosition;
            }
        }else{
            //autorecenter
            // if(this.tiltPos > this.midPosition){
            //     this.tiltPos -= this.tiltSpeed;
            // }
            // if(this.tiltPos < this.midPosition){
            //     this.tiltPos += this.tiltSpeed;
            // }
            
            // let diffrence = this.tiltPos - this.midPosition
            // if(Math.abs(diffrence) > this.tiltSpeed){
            //     this.tiltPos = this.midPosition
            // }
        }

        //Resent camera on spacebar
        if(this.pressedKeys[' '] === true) {
            this.tiltPos = this.midPosition;
            this.panPos = this.midPosition;
        }

        command[3] = this.panPos;
        command[4] = this.tiltPos;



        return command
    }
}