# Manual Installation

This guide will walk you through manually installing and configuring Fireactions on a Linux machine.

## Overview

Fireactions consists of several components working together:

- **Firecracker**: A lightweight virtual machine monitor that runs the runner environments
- **Containerd**: Container runtime that pulls and manages runner images
- **CNI Plugins**: Networking layer that connects VMs to the host network
- **Linux Kernel**: A minimal kernel image optimized for Firecracker VMs
- **Fireactions**: The orchestration service that ties everything together

## Prerequisites

Before you begin, ensure you have:

### Required

- **Linux server** with x86_64 (amd64) or aarch64 (arm64) architecture
- **Root or sudo access** for system-level configuration
- **KVM virtualization support** - Firecracker requires hardware virtualization
- **GitHub App credentials**:
    - App ID
    - App must be installed on your target organization
    - Private key (PEM format)
    - See [Creating GitHub Apps](https://docs.github.com/en/apps/creating-github-apps) for setup instructions
- **Dedicated block device** for Containerd storage (e.g., `/dev/nvme1n1`, `/dev/sdb`)
    - This will be used exclusively for container image storage via LVM
    - Minimum 50GB recommended, though this depends on your image sizes

### Verify Hardware Virtualization

Firecracker requires KVM (Kernel-based Virtual Machine) support:

```bash
# Check if KVM device exists
ls -la /dev/kvm

# Check if your CPU supports virtualization
grep -E '(vmx|svm)' /proc/cpuinfo
```

If `/dev/kvm` doesn't exist: Enable VT-x in BIOS (Intel) or AMD-V in BIOS (AMD) or ensure nested virtualization is enabled (Cloud providers).

## Step 1: Install System Dependencies

Install the base packages needed for downloading and managing components.

**For Ubuntu/Debian:**
```bash
apt-get update
apt-get install -y curl gnupg lvm2 tar
```

**For RHEL/CentOS/Rocky Linux:**
```bash
yum install -y curl lvm2 tar
```

## Step 2: Install Firecracker

[Firecracker](https://github.com/firecracker-microvm/firecracker) is an open-source virtualization technology that enables lightweight virtual machines. It provides the isolation layer for each GitHub Actions runner.

First, determine your system architecture and find the latest version:

```bash
# Set architecture variable
export ARCH=$(case $(uname -m) in
  x86_64) echo "amd64" ;;
  aarch64) echo "arm64" ;;
esac)

# Check latest release at: https://github.com/firecracker-microvm/firecracker/releases
export FIRECRACKER_VERSION=1.9.1  # Replace with latest version
```

Download and install the Firecracker binary:

```bash
curl -fsSL -o firecracker.tgz \
  "https://github.com/firecracker-microvm/firecracker/releases/download/v${FIRECRACKER_VERSION}/firecracker-v${FIRECRACKER_VERSION}-$(uname -m).tgz"

tar -xzf firecracker.tgz --strip-components=1
mv firecracker-v${FIRECRACKER_VERSION}-$(uname -m) /usr/local/bin/firecracker
chmod +x /usr/local/bin/firecracker
rm firecracker.tgz
```

Verify the installation:

```bash
firecracker --version
```

You should see output like: `Firecracker v1.9.1`

## Step 3: Install Containerd

[Containerd](https://containerd.io) is an industry-standard container runtime that manages the lifecycle of containers. Fireactions uses it to pull runner images from container registries and prepare root filesystems for the VMs.

### Download and Install Containerd

Check for the latest release at https://github.com/containerd/containerd/releases:

```bash
export CONTAINERD_VERSION=1.7.24  # Replace with latest 1.7.x version

curl -fsSL -o containerd.tar.gz \
  "https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-${ARCH}.tar.gz"

tar -xzf containerd.tar.gz
mv bin/containerd /usr/local/bin/containerd
mv bin/ctr /usr/local/bin/ctr
rm -rf bin containerd.tar.gz
```

### Create Containerd Systemd Service

Systemd will manage the Containerd daemon lifecycle:

```bash
cat > /etc/systemd/system/containerd.service << 'EOF'
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
LimitNPROC=infinity
LimitCORE=infinity
LimitNOFILE=infinity
TasksMax=infinity
OOMScoreAdjust=-999

[Install]
WantedBy=multi-user.target
EOF
```

### Configure Containerd with Devmapper Snapshotter

Containerd uses snapshotters to manage container filesystem layers. The `devmapper` snapshotter with LVM thin provisioning provides efficient storage management and better performance for our use case.

Create the configuration directory and file:

```bash
mkdir -p /etc/containerd

cat > /etc/containerd/config.toml << 'EOF'
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
```

Adjust `base_image_size` as needed depending on your container image sizes.

### Setup LVM Thin Pool for Containerd

LVM thin provisioning allows efficient storage allocation. Instead of pre-allocating disk space for each container, space is allocated on-demand.

**Important**: This will wipe the specified device. Ensure you're using the correct device and it contains no important data.

```bash
# Set your dedicated device path
export CONTAINERD_SNAPSHOTTER_DEVICE=/dev/nvme1n1  # CHANGE THIS

# Verify the device exists and is unmounted
lsblk ${CONTAINERD_SNAPSHOTTER_DEVICE}
```

Create the LVM physical volume and volume group:

```bash
# Create physical volume
pvcreate -f ${CONTAINERD_SNAPSHOTTER_DEVICE}

# Create volume group named 'containerd'
vgcreate containerd ${CONTAINERD_SNAPSHOTTER_DEVICE}
```

Configure thin pool auto-extension (prevents space exhaustion):

```bash
cat > /etc/lvm/profile/containerd.profile << 'EOF'
activation {
  thin_pool_autoextend_threshold=80
  thin_pool_autoextend_percent=20
}
EOF
```

This configuration enables automatic extension of the thin pool when it reaches 80% capacity, increasing its size by 20% each time.

Create the thin pool:

```bash
lvcreate --type thin-pool -q -n thinpool \
  --poolmetadatasize 1G \
  --profile containerd \
  --monitor y \
  -l "95%VG" containerd
```

This creates the thin pool named `thinpool` in the `containerd` volume group.

### Start Containerd

Enable and start the Containerd service:

```bash
systemctl daemon-reload
systemctl enable containerd
systemctl start containerd

# Verify it's running
systemctl status containerd

# Test with a simple command
ctr version
```

## Step 4: Install CNI Plugins

The [Container Network Interface (CNI)](https://github.com/containernetworking/cni) provides networking for containers and VMs. Fireactions uses several CNI plugins to create isolated network namespaces and connect VMs to the host network.

### Install Standard CNI Plugins

Check for the latest release at https://github.com/containernetworking/plugins/releases:

```bash
export CNI_VERSION=1.6.1  # Replace with latest version

curl -fsSL -o cni-plugins.tgz \
  "https://github.com/containernetworking/plugins/releases/download/v${CNI_VERSION}/cni-plugins-linux-${ARCH}-v${CNI_VERSION}.tgz"

mkdir -p /opt/cni/bin
tar -xzf cni-plugins.tgz -C /opt/cni/bin
rm cni-plugins.tgz

# Verify installation
ls -lh /opt/cni/bin/
```

This installs `bridge`, `host-local`, and `firewall` plugins, which are essential for networking functionality.

### Install tc-redirect-tap Plugin

The [tc-redirect-tap](https://github.com/awslabs/tc-redirect-tap) plugin uses Traffic Control (tc) to redirect packets between the VM's tap device and the host, providing better performance than traditional bridging.

```bash
curl -fsSL -o /opt/cni/bin/tc-redirect-tap \
  "https://github.com/hostinger/tc-redirect-tap/releases/download/v0.0.1/tc-redirect-tap-${ARCH}"
chmod +x /opt/cni/bin/tc-redirect-tap

# Verify
/opt/cni/bin/tc-redirect-tap --version
```

### Configure CNI Network

Create the network configuration that Fireactions will use:

```bash
mkdir -p /etc/cni/net.d

cat > /etc/cni/net.d/10-fireactions.conflist << 'EOF'
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
```

Adjust the configuration as needed, especially the subnet. We recommend using a big subnet range (e.g., /23) to accommodate future growth.

## Step 5: Download Kernel Image

Firecracker VMs require a Linux kernel. We provide pre-built, optimized kernel images, which you can download as follows:

```bash
export KERNEL_VERSION=5.10  # or 6.1 for newer kernel

mkdir -p /var/lib/fireactions/kernels/${KERNEL_VERSION}
curl -fsSL -o /var/lib/fireactions/kernels/${KERNEL_VERSION}/vmlinux \
  "https://storage.googleapis.com/fireactions/kernels/${ARCH}/${KERNEL_VERSION}/vmlinux"

# Verify download
ls -lh /var/lib/fireactions/kernels/${KERNEL_VERSION}/vmlinux
```

The kernel is configured with minimal modules to reduce the attack surface and improve boot times.

## Step 6: Install Fireactions

Now install the Fireactions orchestrator that coordinates all these components.

### Download Fireactions Binary

Check for the latest release at https://github.com/hostinger/fireactions/releases:

```bash
export FIREACTIONS_VERSION=1.0.0  # Replace with latest version

curl -fsSL -o fireactions.tar.gz \
  "https://github.com/hostinger/fireactions/releases/download/v${FIREACTIONS_VERSION}/fireactions-v${FIREACTIONS_VERSION}-linux-${ARCH}.tar.gz"

tar -xzf fireactions.tar.gz
mv fireactions /usr/local/bin/fireactions
chmod +x /usr/local/bin/fireactions
rm fireactions.tar.gz

# Verify installation
fireactions version
```

### Configure Sysctl for IP Forwarding

Enable IP forwarding to allow VMs to reach external networks:

```bash
cat > /etc/sysctl.d/99-fireactions.conf << 'EOF'
net.ipv4.conf.all.forwarding=1
net.ipv4.ip_forward=1
EOF

# Apply immediately
sysctl -p /etc/sysctl.d/99-fireactions.conf
```

### Create Fireactions Configuration

Create the main configuration file. This tells Fireactions how to connect to GitHub and how to provision runners.

```bash
mkdir -p /etc/fireactions
```

Create `/etc/fireactions/config.yaml` with your specific values:

```yaml
# Address where Fireactions will listen (change to 0.0.0.0:8080 for external access)
bind_address: 127.0.0.1:8080

# Prometheus metrics endpoint
metrics:
  enabled: true
  address: 127.0.0.1:8081

# GitHub App authentication
github:
  app_id: YOUR_GITHUB_APP_ID
  app_private_key: |
    -----BEGIN RSA PRIVATE KEY-----
    YOUR_PRIVATE_KEY_CONTENT_HERE
    -----END RSA PRIVATE KEY-----

# Runner pool configuration
pools:
- name: default
  replicas: 1  # Number of concurrent runners in this pool
  runner:
    name: default
    # Runner image - must be compatible with Fireactions
    image: ghcr.io/hostinger/fireactions-images/ubuntu22.04:latest
    image_pull_policy: IfNotPresent  # or Always to pull on every run
    group_id: 1  # Runner group ID in GitHub (1 = default)
    organization: YOUR_GITHUB_ORGANIZATION  # or use 'repository: owner/repo'
    labels:
    - self-hosted
    - fireactions
    # Add more labels to target specific workflows
    # - gpu
    # - large-runner
  firecracker:
    binary_path: firecracker
    kernel_image_path: /var/lib/fireactions/kernels/5.10/vmlinux
    kernel_args: "console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw"
    machine_config:
      mem_size_mib: 2048  # RAM per VM (adjust based on your workloads)
      vcpu_count: 2       # vCPUs per VM
    # Custom metadata passed to VMs (accessible via MMDS)
    metadata:
      pool: default
      environment: production

# Logging level: debug, info, warn, error
log_level: info
```

For all configuration options, see the [configuration guide](../reference/configuration.md).

### Create Fireactions Systemd Service

Set up Fireactions to run as a system service:

```bash
cat > /etc/systemd/system/fireactions.service << 'EOF'
[Unit]
Description=Fireactions
Documentation=https://github.com/hostinger/fireactions
After=network.target containerd.service
Requires=containerd.service

[Service]
User=root
Type=simple
KillMode=process
ExecStartPre=/usr/bin/which firecracker
ExecStartPre=/usr/bin/which containerd
ExecStart=/usr/local/bin/fireactions server --config /etc/fireactions/config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

### Start Fireactions

Enable and start the service:

```bash
systemctl daemon-reload
systemctl enable fireactions
systemctl start fireactions
```

## Step 7: Verify Installation

### Check Service Status

```bash
systemctl status fireactions
```

Expected output:

```
● fireactions.service - Fireactions
     Loaded: loaded (/etc/systemd/system/fireactions.service; enabled; preset: enabled)
     Active: active (running) since Mon 2024-12-09 10:30:15 UTC; 5min ago
       Docs: https://github.com/hostinger/fireactions
    Process: 3564 ExecStartPre=/usr/bin/which firecracker (code=exited, status=0/SUCCESS)
    Process: 3566 ExecStartPre=/usr/bin/which containerd (code=exited, status=0/SUCCESS)
   Main PID: 3571 (fireactions)
      Tasks: 15
     Memory: 45.2M
        CPU: 2.134s
```

Refer to [Troubleshooting Guide](../help/troubleshooting.md) in case of issues.
