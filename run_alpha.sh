#!/bin/sh

file="./goremotecontrol_web"

if [ -f "$file" ] ; then
    rm "$file"
fi

echo Compiling...
go build .

export $(grep -v '^#' alpha_car.env | xargs)
export XDG_RUNTIME_DIR=""

sudo pactl load-module module-echo-cancel source_master=alsa_input.usb-C-Media_Electronics_Inc._USB_PnP_Sound_Device-00.mono-fallback aec_method=webrtc source_name=echocancel sink_name=echocancel1

sudo -E ./goremotecontrol_web