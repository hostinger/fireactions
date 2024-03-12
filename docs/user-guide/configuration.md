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
# Enable basic authentication.
#
# Default: false
#
basic_auth_enabled: true

#
# Map of basic authentication users. The key is the username and the value is the password. Valid only when `basic_auth_enabled` is true.
#
# Default: {}
basic_auth_users:
  user1: password1
  user2: password2

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
  # The secret used to sign incoming webhooks. This is used to verify that the webhook is from GitHub.
  # Not required, but recommended.
  #
  # Default: ""
  #
  webhook_secret: secret

#
# Pools configuration.
#
pools:
  #
  # The name of the pool.
  #
- name: fireactions-2vcpu-2gb
  #
  # The maximum number of GitHub runners that can be created in the pool.
  #
  # Required: true
  #
  max_runners: 20
  #
  # The minimum number of GitHub runners that should be running in the pool.
  #
  # Required: true
  #
  min_runners: 10
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
