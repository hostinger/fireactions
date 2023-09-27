#!/bin/bash
set -Eeuo pipefail

echo "Running Job Completed Hooks"
for hook in /home/runner/hooks/job-completed.d/*; do
  echo "Running hook: $hook"
  "$hook" "$@"
done
