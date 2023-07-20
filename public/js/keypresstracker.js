class KeyPressTracker {
    constructor() {
        this.pressedKeys = {};

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

    getCommand() {
        let command = [127,127,127,127];
        if(this.pressedKeys['s'] === true) {
            command[0] = 0;
        }else if(this.pressedKeys['w'] === true) {
            command[0] = 255;
        }else{
            command[0] = 127;
        }

        if(this.pressedKeys['a'] === true) {
            command[1] = 0;
        }else if(this.pressedKeys['d'] === true) {
            command[1] = 255;
        }else{
            command[1] = 127;
        }

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