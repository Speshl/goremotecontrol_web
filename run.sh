#!/bin/sh

go build .

export $(grep -v '^#' alpha_car.env | xargs)

sudo ./goremotecontrol_web