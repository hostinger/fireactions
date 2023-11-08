#!/bin/sh
set -e

FIREACTIONS_CONFIG=${FIREACTIONS_CONFIG:-/etc/fireactions/config.yaml}

# If the user is trying to run Fireactions directly with some arguments, then
# pass them to Fireactions.
if [ "$(printf "%s" "$1" | cut -c 1)" = '-' ]; then
  set -- fireactions "$@"
fi

if [ "$1" = '' ]; then
  set -- fireactions --help
fi

if [ "$1" = 'server' ]; then
  shift
  set -- fireactions server "$@"
  if [ ! -f "$FIREACTIONS_CONFIG" ]; then
    echo "ERR: Config file $FIREACTIONS_CONFIG does not exist. Exiting."
    exit 1
  fi

  # Get data_dir from config file
  DATA_DIR=$(grep -E '^data_dir:' "$FIREACTIONS_CONFIG" | awk '{print $2}')

  if [ -z "$DATA_DIR" ]; then
    echo "ERR: data_dir not found in config file $FIREACTIONS_CONFIG. Exiting."
    exit 1
  fi

  # Create data_dir if it doesn't exist
  if [ ! -d "$DATA_DIR" ]; then
    mkdir -p "$DATA_DIR" && chown fireactions:fireactions "$DATA_DIR"
  fi
elif [ "$1" = 'client' ]; then
  set -- fireactions client "$@"
fi

if [ "$1" = 'fireactions' ]; then
  exec runuser -u fireactions -- "$@"
fi

exec "$@"
