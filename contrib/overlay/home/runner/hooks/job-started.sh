#!/bin/bash
set -Eeuo pipefail

echo "Running Job Started Hooks"
for hook in /home/runner/hooks/job-started.d/*; do
  echo "Running hook: $hook"
  "$hook" "$@"
done
