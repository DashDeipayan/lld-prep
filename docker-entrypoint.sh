#!/bin/sh
# Hosted disks (e.g. Render) are mounted root-owned; fix ownership, then drop
# to the unprivileged user so exercise code never runs as root.
set -e
if [ "$(id -u)" = "0" ]; then
  mkdir -p /data
  chown -R app:app /data
  exec setpriv --reuid=app --regid=app --init-groups env HOME=/home/app "$@"
fi
exec "$@"
