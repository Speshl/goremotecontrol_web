#!/bin/sh

sudo aplay ./internal/caraudio/starwars.wav 2>&1 | tee ./playOutputLog.txt