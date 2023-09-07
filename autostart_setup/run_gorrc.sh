#!/bin/sh
export $(grep -v '^#' gorrc.env | xargs)
export XDG_RUNTIME_DIR=""

sudo -E /usr/local/bin/gorrc