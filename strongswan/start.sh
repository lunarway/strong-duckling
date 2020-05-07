#!/bin/bash

export IP=$(hostname -I)

# Add internal IP
ip address add $VPN_LOCAL_NETWORK/32 dev eth0
# Add route for remote network over internal IP
ip route add $VPN_REMOTE_NETWORK/32 via $VPN_LOCAL_NETWORK

cat /config/swanctl.conf | gomplate >/etc/swanctl/swanctl.conf
cat /config/supervisord.conf | gomplate >/supervisord.conf

/usr/bin/supervisord -c /supervisord.conf
