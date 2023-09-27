# Configuration

Fireactions uses YAML for configuration. Both the client and server search for a configuration file in the following locations:

1. `/etc/fireactions/config.yaml`
2. `~/.fireactions/config.yaml`
3. `./config.yaml`

## Client configuration

Example configuration for a client:

```yaml
---
fireactions_server_url: http://127.0.0.1:8080

node:
  name: ""
  cpu_overcommit_ratio: 1.0
  ram_overcommit_ratio: 1.0
  labels:
    fireactions/region: default

poll_interval: 5s

heartbeat_success_threshold: 1
heartbeat_failure_threshold: 1
heartbeat_interval: 1s

firecracker:
  binary_path: ./firecracker
  kernel_image_path: vmlinux.bin
  kernel_args: console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw
  socket_path: /var/run/fireactions/%s.sock
  log_file_path: /var/log/fireactions/%s.log
  log_level: debug

containerd:
  address: /run/containerd/containerd.sock

cni:
  conf_dir: ./cni/conf.d
  bin_dirs:
  - ./cni/bin

metrics:
  listen_addr: 127.0.0.1:8080
  enabled: true

log_level: debug
```

## Server configuration

Example configuration for a server:

```yaml
---
data-dir: /var/lib/fireactions

http:
  listen_addr: 0.0.0.0:8081

github:
  webhook_secret: SECRET
  job_label_prefix: fireactions-
  job_labels:
  - name: 2vcpu-2gb
    allowed_repositories:
    - *
    runner:
      image: ghcr.io/hostinger/fireactions/runner:ubuntu-20.04-x64-2.310.2
      image_pull_policy: IfNotPresent
      resources:
        memory_mb: 2048
        vcpus: 2
      affinity:
      - { key: fireactions/region, operator: In, values: [default] }
  app_id: 123456
  app_private_key: |
    -----BEGIN RSA PRIVATE KEY-----

log_level: debug
```
