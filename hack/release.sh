#!/bin/bash
#
# This script is used to prepare the release artifacts for fireactions.
#
# Usage:
#   hack/release.sh
#

set -e

FIRECRACKER_VERSION="1.4.1"
CNI_PLUGINS_VERSION="1.3.0"

pushd () {
  command pushd "$@" > /dev/null
}

popd () {
  command popd "$@" > /dev/null
}

function download_firecracker_amd64()
{
  mkdir -p tmp
  curl -fsSL -o firecracker.tgz "https://github.com/firecracker-microvm/firecracker/releases/download/v${FIRECRACKER_VERSION}/firecracker-v${FIRECRACKER_VERSION}-x86_64.tgz"
  tar -zxf firecracker.tgz \
    --strip-components=1   \
    -C tmp
  rm -rf firecracker.tgz

  cp tmp/firecracker-v${FIRECRACKER_VERSION}-x86_64 firecracker
  rm -rf tmp
}

function download_firecracker_arm64()
{
  mkdir -p tmp
  curl -fsSL -o firecracker.tgz "https://github.com/firecracker-microvm/firecracker/releases/download/v${FIRECRACKER_VERSION}/firecracker-v${FIRECRACKER_VERSION}-aarch64.tgz"
  tar -zxf firecracker.tgz \
    --strip-components=1   \
    -C tmp
  rm -rf firecracker.tgz

  cp tmp/firecracker-v${FIRECRACKER_VERSION}-aarch64 firecracker
  rm -rf tmp
}

function download_cni_plugins_amd64()
{
  mkdir -p cni/bin tmp
  curl -fsSL -o cni-plugins.tgz https://github.com/containernetworking/plugins/releases/download/v${CNI_PLUGINS_VERSION}/cni-plugins-linux-amd64-v${CNI_PLUGINS_VERSION}.tgz
  tar -zxf cni-plugins.tgz -C tmp && rm -rf cni-plugins.tgz

  curl -fsSL -o tmp/tc-redirect-tap https://github.com/alexellis/tc-tap-redirect-builder/releases/download/2022-04-01-1337/tc-redirect-tap
  chmod +x tmp/tc-redirect-tap

  cp tmp/bridge tmp/host-local tmp/firewall tmp/tc-redirect-tap cni/bin
  rm -rf tmp
}

function download_cni_plugins_arm64()
{
  mkdir -p cni/bin tmp
  curl -fsSL -o cni-plugins.tgz https://github.com/containernetworking/plugins/releases/download/v${CNI_PLUGINS_VERSION}/cni-plugins-linux-arm64-v${CNI_PLUGINS_VERSION}.tgz
  tar -zxf cni-plugins.tgz -C tmp && rm -rf cni-plugins.tgz

  curl -fsSL -o tmp/tc-redirect-tap https://github.com/alexellis/tc-tap-redirect-builder/releases/download/2022-04-01-1337/tc-redirect-tap-arm64
  chmod +x tmp/tc-redirect-tap

  cp tmp/bridge tmp/host-local tmp/firewall tmp/tc-redirect-tap cni/bin
  rm -rf tmp
}

function download_kernel_amd64()
{
  curl -fsSL -o vmlinux https://storage.googleapis.com/fireactions/kernels/amd64/5.10/vmlinux
}

function download_kernel_arm64()
{
  curl -fsSL -o vmlinux https://storage.googleapis.com/fireactions/kernels/arm64/5.10/vmlinux
}

function main()
{
  echo "[+] Preparing dependencies for release..."

  mkdir -p dist release/amd64 release/arm64

  echo "[+] Preparing amd64 release in ./release/amd64 (this may take a while)"
  pushd release/amd64

  download_firecracker_amd64
  download_cni_plugins_amd64
  download_kernel_amd64

  mkdir -p cni/conf.d
  cp ../../contrib/fireactions.conflist cni/conf.d/fireactions.conflist

  popd

  echo "[+] Preparing arm64 release in ./release/arm64 (this may take a while)"
  pushd release/arm64

  download_firecracker_arm64
  download_cni_plugins_arm64
  download_kernel_arm64

  mkdir -p cni/conf.d
  cp ../../contrib/fireactions.conflist cni/conf.d/fireactions.conflist

  popd

  echo "[+] Done."
}

main
