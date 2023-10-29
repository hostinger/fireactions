# Configuration

Fireactions uses YAML for configuration. Both the client and server search for a configuration file in the following locations:

1. `/etc/fireactions/config.yaml`
2. `~/.fireactions/config.yaml`
3. `./config.yaml`

## Client configuration

Example configuration for a client:

```yaml
---
fireactions-server-url: http://127.0.0.1:8080

node:
  name: ""
  cpu-overcommit-ratio: 1.0
  ram-overcommit-ratio: 1.0
  labels:
    fireactions/region: default

poll-interval: 5s

heartbeat-success-threshold: 1
heartbeat-failure-threshold: 1
heartbeat-interval: 1s

firecracker:
  binary-path: ./firecracker
  kernel-image-path: vmlinux.bin
  kernel-args: console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw
  socket-path: /var/run/fireactions/%s.sock
  log-file-path: /var/log/fireactions/%s.log
  log-level: debug

containerd:
  address: /run/containerd/containerd.sock

cni:
  conf-dir: ./cni/conf.d
  bin-dirs:
  - ./cni/bin

metrics:
  listen-addr: 127.0.0.1:8080
  enabled: true

log-level: debug
```

## Server configuration

Example configuration for a server:

```yaml
---
data-dir: /var/lib/fireactions

http:
  listen-addr: 0.0.0.0:8081

github:
  webhook-secret: SECRET
  job-label-prefix: fireactions-
  job-labels:
  - name: 2vcpu-2gb
    allowed-repositories:
    - *
    runner:
      image: ghcr.io/hostinger/fireactions/runner:ubuntu-20.04-x64-2.310.2
      image-pull-policy: IfNotPresent
      resources:
        memory-mb: 2048
        vcpus: 2
      affinity:
      - { key: fireactions/region, operator: In, values: [default] }
  app-id: 123456
  app-private-key: |
    -----BEGIN RSA PRIVATE KEY-----

log-level: debug
```
