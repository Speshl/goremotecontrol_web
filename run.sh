#!/bin/sh

go build .

export $(grep -v '^#' alpha_car.env | xargs)
export XDG_RUNTIME_DIR=""

sudo -E ./goremotecontrol_web