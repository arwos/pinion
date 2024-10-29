#!/bin/bash


if [ -f "/etc/systemd/system/pinion.service" ]; then
    systemctl stop pinion
    systemctl disable pinion
    systemctl daemon-reload
fi
