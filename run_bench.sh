#!/bin/sh

file="./goremotecontrol_web"

if [ -f "$file" ] ; then
    rm "$file"
fi

echo Compiling...
go build .

export $(grep -v '^#' bench_car.env | xargs)
export XDG_RUNTIME_DIR=""

sudo -E ./goremotecontrol_web