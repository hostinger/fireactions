#!/bin/sh
set -e

# If the user is trying to run Fireactions directly with some arguments, then
# pass them to Fireactions.
if [ "$(printf "%s" "$1" | cut -c 1)" = '-' ]; then
  set -- fireactions "$@"
fi

if [ "$1" = '' ]; then
  set -- fireactions --help
fi

if [ "$1" = 'fireactions' ]; then
  exec runuser -u fireactions -- "$@"
else
  exec "$@"
fi
