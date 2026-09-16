# Configuration

Fireactions is configured using a YAML file. The default configuration file is located at `/etc/fireactions/config.yaml`.

You can also specify a custom configuration file using the `--config` flag when starting the Fireactions server.

Example configuration file with all available options:

```yaml
---
#
# The address to listen on for HTTP requests.
#
# Default: :8080
#
bind_address: 0.0.0.0:8080

#
# Metrics server configuration. This is used to expose Prometheus metrics on endpoint `/metrics`.
#
metrics:
  #
  # Enable Prometheus metrics.
  #
  enabled: true

  #
  # The address to listen on for HTTP requests.
  #
  address: 127.0.0.1:8081

#
# GitHub configuration.
#
github:
  #
  # The GitHub App private key. This is used to authenticate with GitHub.
  #
  # Default: ""
  #
  app_private_key: |
    -----BEGIN RSA PRIVATE KEY-----
  #
  # The GitHub App ID.
  #
  # Default: 0
  app_id: 12345

#
# Pools configuration.
#
pools:
  #
  # The name of the pool.
  #
- name: fireactions-2vcpu-2gb
  #
  # The number of replicas for the GitHub runners in the pool.
  #
  # Required: true
  #
  replicas: 5
  #
  # Shutdown the Firecracker VM when the runner exits.
  #
  # Required: false, Default: true
  #
  shutdown_on_exit: true
  #
  # GitHub runner configuration.
  #
  runner:
    #
    # The name of the GitHub runner. This is used to identify the runner in GitHub and is suffixed with a unique identifier.
    #
    # Required: true
    name: fireactions-2vcpu-2gb
    #
    # Container image to use for the Firecracker VM as the root device.
    #
    # Required: true
    #
    image: ghcr.io/hostinger/fireactions/runner:ubuntu-20.04-x64-2.310.2
    #
    # The pull policy for the container image. Can be one of: Always, IfNotPresent, Never.
    #
    # Required: true
    image_pull_policy: IfNotPresent
    #
    # GitHub runner group ID. 1 is the default group.
    #
    # Required: true
    group_id: 1
    #
    # Organization name.
    #
    # Required: true
    #
    organization: hostinger
    #
    # Labels to apply to the GitHub runner.
    #
    # Required: true
    #
    labels:
    - self-hosted
    - fireactions-2vcpu-2gb
    - fireactions
  #
  # Firecracker configuration.
  #
  firecracker:
    #
    # The path to the Firecracker binary.
    #
    # Default: firecracker
    #
    binary_path: firecracker
    #
    # The path to the kernel image.
    #
    # Required: true
    #
    kernel_image_path: /var/lib/fireactions/vmlinux
    #
    # Kernel command line arguments.
    #
    # Default: "console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw"
    #
    kernel_args: "console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw"
    #
    # Firecracker machine configuration.
    #
    # Required: true
    #
    machine_config:
      #
      # The amount of memory in MiB.
      #
      # Required: true
      #
      mem_size_mib: 2048
      #
      # The number of vCPUs.
      #
      # Required: true
      #
      vcpu_count: 2
    #
    # Firecracker network interface configuration.
    #
    # Default: {} (no rate limiting)
    #
    network_interface:
      #
      # Rate limiter for incoming (ingress) traffic. Maps to the Firecracker
      # network interface `rx_rate_limiter`. Both the `bandwidth` and `ops`
      # token buckets are optional, an omitted bucket means unlimited.
      #
      # The resulting rate is `size` / `refill_time`.
      #
      # Default: {} (unlimited)
      #
      in_rate_limiter:
        #
        # Token bucket with bytes as tokens.
        #
        # Default: {} (unlimited)
        #
        bandwidth:
          #
          # The total number of tokens (bytes) the bucket can hold.
          #
          # Required: true
          #
          size: 131072000
          #
          # The amount of milliseconds it takes for the bucket to refill.
          # 131072000 bytes per 1000 ms is ~125 MiB/s.
          #
          # Required: true
          #
          refill_time: 1000
          #
          # The initial burst size (bytes). Consumed before the refill process
          # starts happening.
          #
          # Default: 0
          #
          one_time_burst: 262144000
        #
        # Token bucket with operations (packets) as tokens.
        #
        # Default: {} (unlimited)
        #
        ops:
          size: 10000
          refill_time: 1000
      #
      # Rate limiter for outgoing (egress) traffic. Maps to the Firecracker
      # network interface `tx_rate_limiter`. Same structure as
      # `in_rate_limiter`.
      #
      # Default: {} (unlimited)
      #
      out_rate_limiter:
        bandwidth:
          size: 26214400
          refill_time: 1000
    #
    # Firecracker root block device configuration.
    #
    # Default: {} (no rate limiting)
    #
    rootfs:
      #
      # Rate limiter for the root block device. Maps to the Firecracker drive
      # `rate_limiter`. Both the `bandwidth` and `ops` token buckets are
      # optional, an omitted bucket means unlimited.
      #
      # The resulting rate is `size` / `refill_time`.
      #
      # Default: {} (unlimited)
      #
      rate_limiter:
        #
        # Token bucket with bytes as tokens. 52428800 bytes per 1000 ms is
        # ~50 MiB/s of disk throughput.
        #
        # Default: {} (unlimited)
        #
        bandwidth:
          size: 52428800
          refill_time: 1000
        #
        # Token bucket with operations (IOPS) as tokens.
        #
        # Default: {} (unlimited)
        #
        ops:
          size: 2000
          refill_time: 1000
    #
    # Metadata to pass to the Firecracker VM via MMDS.
    #
    # Default: {}
    #
    metadata:
      example1: value1
      example2: value2

#
# Log level. Can be one of: debug, info, warn, error, fatal, panic, trace.
#
# Default: info
#
log_level: debug
```
