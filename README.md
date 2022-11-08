# activity-server
service that polls established ssh connections to determine activity

# Howto

copy activity-server.service to /lib/systemd/system/activity-server.service

copy activity-server to /usr/local/bin/

systemctl enable --now activity-server.service
