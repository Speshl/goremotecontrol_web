#!/bin/sh

go build .

export $(grep -v '^#' alpha_car.env | xargs)

sudo -E ./goremotecontrol_web