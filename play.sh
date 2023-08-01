#!/bin/sh

sudo aplay ~/scripts/starwars.wav 2>&1 | tee ./playOutputLog.txt