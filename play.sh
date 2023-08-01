#!/bin/sh

sudo aplay $1 2>&1 | tee ./playOutputLog.txt