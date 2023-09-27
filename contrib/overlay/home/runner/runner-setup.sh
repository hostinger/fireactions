#!/bin/bash
set -Eeuo pipefail

echo "Setting up GitHub Actions self-hosted runner..."
echo "Retrieving metadata..."

runnerctl set-configuring

MMDS_TOKEN_TTL=300
MMDS_TOKEN=$(curl "http://169.254.169.254/latest/api/token" -s -X PUT -H "X-Metadata-Token-TTL-Seconds: $MMDS_TOKEN_TTL")
if [ -z "$MMDS_TOKEN" ]; then
  echo "Unable to retrieve X-Metadata-Token from MMDS. Exiting."
  exit 1
fi

RUNNER_NAME=$(curl "http://169.254.169.254/latest/meta-data/runner-name" -s -H "Accept: application/json" -H "X-Metadata-Token: $MMDS_TOKEN" | jq -r)
if [ -z "$RUNNER_NAME" ]; then
  echo "Unable to set RUNNER_NAME from MMDS. Exiting."
  exit 1
fi

RUNNER_TOKEN=$(curl "http://169.254.169.254/latest/meta-data/runner-token" -s -H "Accept: application/json" -H "X-Metadata-Token: $MMDS_TOKEN" | jq -r)
if [ -z "$RUNNER_TOKEN" ]; then
  echo "Unable to set RUNNER_TOKEN from MMDS. Exiting."
  exit 1
fi

RUNNER_LABEL=$(curl "http://169.254.169.254/latest/meta-data/runner-labels" -s -H "Accept: application/json" -H "X-Metadata-Token: $MMDS_TOKEN" | jq -r)
if [ -z "$RUNNER_LABEL" ]; then
  echo "Unable to set RUNNER_LABEL from MMDS. Exiting."
  exit 1
fi

RUNNER_URL=$(curl "http://169.254.169.254/latest/meta-data/runner-url" -s -H "Accept: application/json" -H "X-Metadata-Token: $MMDS_TOKEN" | jq -r)
if [ -z "$RUNNER_URL" ]; then
  echo "Unable to set RUNNER_URL from MMDS. Exiting."
  exit 1
fi

retries=10
while [[ ${retries} -gt 0 ]]; do
  ./config.sh --url $RUNNER_URL --token $RUNNER_TOKEN --name $RUNNER_NAME --labels $RUNNER_LABEL --unattended --ephemeral --replace
  if [ -f .runner ]; then
    break
  fi
  echo "Failed to configure GitHub Actions self-hosted runner. Retrying..."
  retries=$((retries - 1))
  sleep 1
done

if [ ! -f .runner ]; then
  runnerctl set-error
  echo "Failed to configure GitHub Actions self-hosted runner. Exiting."
  exit 1
fi

runnerctl set-idle

echo "GitHub Actions self-hosted runner configured successfully."
exit 0
