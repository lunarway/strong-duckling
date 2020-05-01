#!/bin/bash

while true; do
  echo "starting vpn"
  supervisorctl start charon
  sleep 2
  /usr/sbin/swanctl --load-all --noprompt

  sleep 15

  echo "stopping vpn"
  supervisorctl stop charon
  sleep 120

done
