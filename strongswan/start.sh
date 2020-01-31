#!/bin/bash

# turn on bash's job control
set -m
ip=$(hostname -I)

# Add internal IP
ip address add $VPN_LOCAL_NETWORK/32 dev eth0
# Add route for remote network over internal IP
ip route add $VPN_REMOTE_NETWORK/32 via $VPN_LOCAL_NETWORK

cat <<EOF >/etc/swanctl/swanctl.conf
connections {
  gw-gw {
    local_addrs  = $ip,$VPN_LOCAL_PEER
    remote_addrs = $VPN_REMOTE_PEER

    local {
      auth = psk
      id = $VPN_LOCAL_PEER
    }
    remote {
      auth = psk
      id = $VPN_REMOTE_PEER
    }
    children {
      net-net-0 {
        local_ts = $VPN_LOCAL_NETWORK/32
        remote_ts = $VPN_REMOTE_NETWORK/32
        updown = /usr/libexec/ipsec/_updown iptables

        rekey_time = 5400
        rekey_bytes = 500000000
        rekey_packets = 1000000
        esp_proposals = aes256-sha256-ecp384

        start_action = start
        close_action = start
        dpd_action = start
      }
    }
    version = 2
    mobike = no
    reauth_time = 10800
    proposals = aes256-sha256-ecp384
  }
}

secrets {
  ike-1 {
    id-local = $VPN_LOCAL_PEER
    id-remote = $VPN_REMOTE_PEER
    secret = "123456"
  }
}
EOF

mkdir -p /etc/supervisor/conf.d/
cat <<EOF >/supervisord.conf
[supervisord]
nodaemon=true

[program:charon]
command=/prefix-log /usr/sbin/charon-systemd
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0

[program:http-server]
command=/prefix-log node /server.js
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0

[program:strong-duckling]
command=/prefix-log nodemon --signal SIGTERM --watch /strong-duckling -x "/strong-duckling $STRONG_DUCKLING_ARGS || exit 1"
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
EOF

/usr/bin/supervisord -c /supervisord.conf &
sleep 2
/usr/sbin/swanctl --load-all --noprompt
fg %1
