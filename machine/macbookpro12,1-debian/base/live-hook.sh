#!/bin/sh
# Install the scripts/config from audiovis-util into Debian's temporary RAM root.
set -eu
data=/media/audiovis-util
[ -r "$data/live-startup.sh" ] || { echo 'audiovis-util startup script is missing.' >&2; exit 0; }

install -d -m 755 /usr/local/sbin /etc/wpa_supplicant
install -d -m 700 /root/.ssh
install -m 755 "$data/console-font.sh" /usr/local/sbin/audiovis-util-console-font
install -m 755 "$data/install-offline.sh" /usr/local/sbin/audiovis-util-install-offline
install -m 755 "$data/audio-utility" /usr/local/bin/audio-utility
install -m 755 "$data/live-startup.sh" /usr/local/sbin/audiovis-util-live-startup
install -m 755 "$data/hardware-test.sh" /usr/local/sbin/audiovis-util-hardware-test
install -m 600 "$data/ssh-authorized-key" /root/.ssh/authorized_keys
install -m 600 "$data/wpa_supplicant.conf" /etc/wpa_supplicant/audiovis-util.conf
install -m 600 "$data/network.conf" /etc/audiovis-util-network.conf

systemctl set-default multi-user.target
mkdir -p /etc/systemd/system/getty@tty1.service.d
cat > /etc/systemd/system/getty@tty1.service.d/zz-audiovis-util.conf <<'EOF'
[Service]
ExecStart=
ExecStart=-/sbin/agetty --autologin root --noclear %I 38400 linux
EOF

cat > /etc/systemd/system/audiovis-util-live-startup.service <<'EOF'
[Unit]
Description=Configure MacBook headless networking, SSH, and hardware report
Wants=console-setup.service
After=local-fs.target console-setup.service
Before=getty@tty1.service

[Service]
Type=oneshot
ExecStart=/usr/local/sbin/audiovis-util-live-startup
RemainAfterExit=yes
TimeoutStartSec=300
StandardOutput=journal+console
StandardError=journal+console

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable audiovis-util-live-startup.service
# live-config runs during early boot. Queue startup without waiting for units
# whose ordering may depend on live-config itself finishing.
systemctl start --no-block audiovis-util-live-startup.service
