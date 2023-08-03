#!/bin/sh

aplay -D hw:CARD=wm8960soundcard,DEV=0 $1
