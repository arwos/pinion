#!/bin/bash


if [ -f "/etc/systemd/system/pinion.service" ]; then
    systemctl start pinion
    systemctl enable pinion
    systemctl daemon-reload
fi
