#!/bin/bash

set -e

export CONTAINERD_VERSION=1.7.0
export CNI_VERSION=1.6.0
export DEBIAN_FRONTEND=noninteractive
export FIREACTIONS_VERSION=0.2.5
export FIRECRACKER_VERSION=1.4.1
export KERNEL_VERSION=5.10
export GITHUB_SKIP_TLS_VERIFY=false

usage() {
  echo "This script installs Fireactions on a Linux machine."
  echo
  echo "Usage: $0 [options]"
  echo
  echo "Options:"
  echo "  --github-app-id                     Sepcify the ID of the GitHub App                          (required)"
  echo "  --github-app-key-file               Specify the path to the GitHub App private key file       (required)"
  echo "  --github-enterprise-api-url         Specify the enterprise GitHub api url                     (optional default: https://api.github.com)"
  echo "  --github-skip-tls-verify            Specify whether or not to verify GitHub certificate       (optional default: false)"
  echo "  --github-organization               Specify the name of the GitHub organization               (required)"
  echo "  --fireactions-version               Specify the Fireactions version to install                (default: $FIREACTIONS_VERSION)"
  echo "  --firecracker-version               Specify the Firecracker version to install                (default: $FIRECRACKER_VERSION)"
  echo "  --kernel-version                    Specify the kernel version to install                     (default: $KERNEL_VERSION)"
  echo "  --containerd-snapshotter-device     Specify the device to use for Containerd snapshot storage (required)"
  echo "  --containerd-version                Specify the Containerd version to install                 (default: $CONTAINERD_VERSION)"
  echo "  --cni-version                       Specify the CNI plugin version to install                 (default: $CNI_VERSION)"
  echo "  -h, --help                          Show this help message"
  echo
}

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt() {
  [ -n "$(command -v apt-get)" ]
}

print_error() {
  echo -e "\033[31mERROR:\033[0m $1"
}

check_kvm() {
  if [[ ! -e /dev/kvm ]]; then
    print_error "Virtualization is not available on this machine, /dev/kvm is missing. Enable virtualization and try again."
    exit 1
  fi
}

install_dependencies() {
  if has_apt; then
    apt-get update -qq -y
    apt-get install -qq -y \
      curl \
      gnupg \
      lvm2 \
      tar
  elif has_yum; then
    yum install -q -y \
      curl \
      lvm2 \
      tar
  else
    print_error "Unsupported package manager"
    exit 1
  fi
}

install_firecracker() {
  if [[ -e /usr/local/bin/firecracker ]]; then
    return
  fi

  TEMP_DIR=$(mktemp -d)

  curl -sL -o "$TEMP_DIR/firecracker-v$FIRECRACKER_VERSION.tgz" \
    "https://github.com/firecracker-microvm/firecracker/releases/download/v$FIRECRACKER_VERSION/firecracker-v$FIRECRACKER_VERSION-$(uname -m).tgz"

  tar -xf "$TEMP_DIR/firecracker-v$FIRECRACKER_VERSION.tgz" -C "$TEMP_DIR" --strip-components=1

  mv "$TEMP_DIR/firecracker-v$FIRECRACKER_VERSION-$(uname -m)" /usr/local/bin/firecracker
  chmod +x /usr/local/bin/firecracker

  rm -rf "$TEMP_DIR"
}

install_containerd() {
  if [[ -e /usr/local/bin/containerd ]]; then
    return
  fi

  TEMP_DIR=$(mktemp -d)

  curl -sL -o "$TEMP_DIR/containerd-$CONTAINERD_VERSION-linux-$ARCH.tar.gz" \
    "https://github.com/containerd/containerd/releases/download/v$CONTAINERD_VERSION/containerd-$CONTAINERD_VERSION-linux-$ARCH.tar.gz"

  tar -zxf "$TEMP_DIR/containerd-$CONTAINERD_VERSION-linux-$ARCH.tar.gz" -C "$TEMP_DIR"

  mv "$TEMP_DIR/bin/containerd" /usr/local/bin/containerd
  mv "$TEMP_DIR/bin/ctr" /usr/local/bin/ctr

  cat <<EOF >/etc/systemd/system/containerd.service
[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target local-fs.target

[Service]
Type=notify
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/local/bin/containerd
Delegate=yes
KillMode=process
Restart=always
RestartSec=5
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNPROC=infinity
LimitCORE=infinity
LimitNOFILE=infinity
# Comment TasksMax if your systemd version does not supports it.
# Only systemd 226 and above support this version.
TasksMax=infinity
OOMScoreAdjust=-999

[Install]
WantedBy=multi-user.target
EOF

  mkdir -p /etc/containerd
  cat <<EOF >/etc/containerd/config.toml
version = 2

root      = "/var/lib/containerd"
imports   = []
state     = "/run/containerd"
oom_score = 0

[grpc]
  address = "/run/containerd/containerd.sock"
  uid     = 0
  gid     = 0

[plugins]
  [plugins."io.containerd.snapshotter.v1.devmapper"]
    pool_name       = "containerd-thinpool"
    root_path       = "/var/lib/containerd/devmapper"
    base_image_size = "30GB"
    discard_blocks  = true
EOF

  # Setup LVM thin pool for Containerd
  pvcreate -f $CONTAINERD_SNAPSHOTTER_DEVICE
  vgcreate containerd $CONTAINERD_SNAPSHOTTER_DEVICE
  cat <<EOF | tee /etc/lvm/profile/containerd.profile
activation {
  thin_pool_autoextend_threshold=80
  thin_pool_autoextend_percent=20
}
EOF
  lvcreate --type thin-pool -q -n thinpool --poolmetadatasize 1G --profile containerd --monitor y -l "95%VG" containerd

  systemctl daemon-reload
  systemctl enable containerd && systemctl start containerd

  rm -rf "$TEMP_DIR"
}

install_cni() {
  if [[ -e /opt/cni/bin/bridge ]]; then
    return
  fi

  TEMP_DIR=$(mktemp -d)

  curl -sL -o "$TEMP_DIR/cni-plugins-linux-$ARCH-v$CNI_VERSION.tgz" \
    "https://github.com/containernetworking/plugins/releases/download/v$CNI_VERSION/cni-plugins-linux-$ARCH-v$CNI_VERSION.tgz"

  mkdir -p /opt/cni/bin
  tar -zxf "$TEMP_DIR/cni-plugins-linux-$ARCH-v$CNI_VERSION.tgz" -C /opt/cni/bin

  curl -sL -o /opt/cni/bin/tc-redirect-tap \
    "https://github.com/hostinger/tc-redirect-tap/releases/download/v0.0.1/tc-redirect-tap-$ARCH"
  chmod +x /opt/cni/bin/tc-redirect-tap

  mkdir -p /etc/cni/net.d
  cat <<EOF >/etc/cni/net.d/10-fireactions.conflist
{
  "cniVersion": "0.4.0",
  "name": "fireactions",
  "plugins": [
    {
      "bridge": "fireactions-br0",
      "forceAddress": false,
      "hairpinMode": true,
      "ipMasq": true,
      "ipam": {
        "dataDir": "/var/run/cni",
        "resolvConf": "/etc/resolv.conf",
        "subnet": "192.168.128.0/24",
        "type": "host-local"
      },
      "isDefaultGateway": true,
      "mtu": 1500,
      "type": "bridge"
    },
    {
      "type": "firewall"
    },
    {
      "type": "tc-redirect-tap"
    }
  ]
}
EOF

  rm -rf "$TEMP_DIR"
}

install_fireactions() {
  if [[ -e /usr/local/bin/fireactions ]]; then
    return
  fi

  TEMP_DIR=$(mktemp -d)

  curl -sL -o "$TEMP_DIR/fireactions-v$FIREACTIONS_VERSION.tar.gz" \
    "https://github.com/hostinger/fireactions/releases/download/v$FIREACTIONS_VERSION/fireactions-v$FIREACTIONS_VERSION-linux-$ARCH.tar.gz"

  tar -zxf "$TEMP_DIR/fireactions-v$FIREACTIONS_VERSION.tar.gz" -C "$TEMP_DIR"

  mv "$TEMP_DIR/fireactions" /usr/local/bin/fireactions
  chmod +x /usr/local/bin/fireactions

  cat <<EOF >/etc/sysctl.d/99-fireactions.conf
net.ipv4.conf.all.forwarding=1
net.ipv4.ip_forward=1
EOF
  sysctl -p /etc/sysctl.d/99-fireactions.conf >/dev/null

  mkdir -p /etc/fireactions
  cat <<EOF >/etc/fireactions/config.yaml
bind_address: 127.0.0.1:8080

metrics:
  enabled: true
  address: 127.0.0.1:8081

github:
  app_id: $GITHUB_APP_ID
  enterprise_api_url: $GITHUB_ENTERPRISE_API_URL
  skip_tls_verify: $GITHUB_SKIP_TLS_VERIFY
  app_private_key: |
    $GITHUB_APP_PRIVATE_KEY

pools:
- name: default
  max_runners: 5
  min_runners: 1
  runner:
    name: default
    image: ghcr.io/hostinger/fireactions-images/ubuntu22.04:v0.7.0
    image_pull_policy: IfNotPresent
    group_id: 1
    organization: $GITHUB_ORGANIZATION
    labels:
    - self-hosted
    - fireactions
  firecracker:
    binary_path: firecracker
    kernel_image_path: /var/lib/fireactions/kernels/$KERNEL_VERSION/vmlinux
    kernel_args: "console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw"
    machine_config:
      mem_size_mib: 2048
      vcpu_count: 2
    metadata:
      example1: value1
      example2: value2

log_level: debug
EOF

  cat <<EOF >/etc/systemd/system/fireactions.service
[Unit]
Description=Fireactions
Documentation=https://github.com/hostinger/fireactions
After=network.target

[Service]
User=root
Type=simple
KillMode=process
ExecStartPre=/usr/bin/which firecracker
ExecStartPre=/usr/bin/which containerd
ExecStart=fireactions server --config /etc/fireactions/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable fireactions && systemctl start fireactions

  rm -rf "$TEMP_DIR"
}

install_kernel() {
  if [[ -e /var/lib/fireactions/kernels/$KERNEL_VERSION/vmlinux ]]; then
    return
  fi

  mkdir -p /var/lib/fireactions/kernels/$KERNEL_VERSION

  curl -sL -o "/var/lib/fireactions/kernels/$KERNEL_VERSION/vmlinux" \
    "https://storage.googleapis.com/fireactions/kernels/$ARCH/$KERNEL_VERSION/vmlinux"
}

main() {
  if [ "$#" -eq 0 ]; then
    usage
    exit 1
  fi

  while [ "$1" != "" ]; do
    case $1 in
    --github-app-id)
      shift
      GITHUB_APP_ID=$1
      ;;
    --github-app-id=*)
      GITHUB_APP_ID="${1#*=}"
      ;;
    --github-app-key-file)
      shift
      GITHUB_APP_PRIVATE_KEY_FILE=$1
      ;;
    --github-app-key-file=*)
      GITHUB_APP_PRIVATE_KEY_FILE="${1#*=}"
      ;;
    --github-enterprise-api-url)
      shift
      GITHUB_ENTERPRISE_API_URL=$1
      ;;
    --github-enterprise-api-url=*)
      GITHUB_ENTERPRISE_API_URL="${1#*=}"
      ;;
    --github-skip-tls-verify)
      shift
      GITHUB_SKIP_TLS_VERIFY=$1
      ;;
    --github-skip-tls-verify=*)
      GITHUB_SKIP_TLS_VERIFY="${1#*=}"
      ;;
    --github-organization)
      shift
      GITHUB_ORGANIZATION=$1
      ;;
    --github-organization=*)
      GITHUB_ORGANIZATION="${1#*=}"
      ;;
    --fireactions-version)
      shift
      FIREACTIONS_VERSION=$1
      ;;
    --fireactions-version=*)
      FIREACTIONS_VERSION="${1#*=}"
      ;;
    --firecracker-version)
      shift
      FIRECRACKER_VERSION=$1
      ;;
    --firecracker-version=*)
      FIRECRACKER_VERSION="${1#*=}"
      ;;
    --kernel-version)
      shift
      KERNEL_VERSION=$1
      ;;
    --kernel-version=*)
      KERNEL_VERSION="${1#*=}"
      ;;
    --containerd-snapshotter-device)
      shift
      CONTAINERD_SNAPSHOTTER_DEVICE=$1
      ;;
    --containerd-snapshotter-device=*)
      export CONTAINERD_SNAPSHOTTER_DEVICE="${1#*=}"
      ;;
    --containerd-version)
      shift
      CONTAINERD_VERSION=$1
      ;;
    --containerd-version=*)
      CONTAINERD_VERSION="${1#*=}"
      ;;
    --cni-version)
      shift
      CNI_VERSION=$1
      ;;
    --cni-version=*)
      CNI_VERSION="${1#*=}"
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 1
      ;;
    esac
    shift
  done

  if [ -z "$GITHUB_APP_ID" ]; then
    print_error "Option --github-app-id is required"
    usage
    exit 1
  fi

  if [ -z "$GITHUB_APP_PRIVATE_KEY_FILE" ]; then
    print_error "Option --github-app-key-file is required"
    usage
    exit 1
  else
    if [ ! -f "$GITHUB_APP_PRIVATE_KEY_FILE" ]; then
      print_error "GitHub App private key file not found: $GITHUB_APP_PRIVATE_KEY_FILE"
      exit 1
    fi
    export GITHUB_APP_PRIVATE_KEY=$(cat $GITHUB_APP_PRIVATE_KEY_FILE | sed '1!s/^/    /')
  fi

  if [ -z "$GITHUB_ORGANIZATION" ]; then
    print_error "Option --github-organization is required"
    usage
    exit 1
  fi

  if [ -z "$CONTAINERD_SNAPSHOTTER_DEVICE" ]; then
    print_error "Option --containerd-snapshotter-device is required"
    usage
    exit 1
  fi

  if [ "$(id -u)" -ne 0 ]; then
    print_error "The installation must be run as sudo or root!"
    exit 1
  fi

  case $(uname -m) in
  x86_64)
    export ARCH=amd64
    ;;
  aarch64)
    export ARCH=arm64
    ;;
  *)
    print_error "Unsupported architecture: $(uname -m)"
    exit 1
    ;;
  esac

  if [[ -e /usr/local/bin/fireactions ]]; then
    echo "Fireactions is already installed. Exiting..."
    exit 1
  fi

  echo "Installing Fireactions v$FIREACTIONS_VERSION..."

  check_kvm
  install_dependencies
  install_kernel
  install_containerd
  install_cni
  install_firecracker
  install_fireactions

  echo "Fireactions v$FIREACTIONS_VERSION has been installed successfully! 🎉"
}

main "$@"
