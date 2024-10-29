#!/bin/bash


if ! [ -d /var/lib/pinion/ ]; then
    mkdir /var/lib/pinion
fi

if [ -f "/etc/systemd/system/pinion.service" ]; then
    systemctl stop pinion
    systemctl disable pinion
    systemctl daemon-reload
fi
