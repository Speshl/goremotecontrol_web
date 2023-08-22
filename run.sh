#!/bin/sh

rm ./goremotecontrol_web

go build .

export $(grep -v '^#' alpha_car.env | xargs)
export XDG_RUNTIME_DIR=""

sudo -E ./goremotecontrol_web