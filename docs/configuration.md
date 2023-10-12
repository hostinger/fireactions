# Configuration

Fireactions uses YAML for configuration. Both the client and server search for a configuration file in the following locations:

1. `/etc/fireactions/config.yaml`
2. `~/.fireactions/config.yaml`
3. `./config.yaml`

## Client configuration

Example configuration for a client:

```yaml
---
# The Fireactions server URL to connect to for the client.
server-url: http://127.0.0.1:8080

# Data directory for the client. This is where the client will store its state, images, etc.
data-dir: /var/lib/fireactions

# The name of the organisation that the client belongs to.
organisation: hostinger
# The name of the group that the client belongs to. This is used to
# group clients together for scheduling.
group: group1

# Mem overcommit ratio is the ratio of RAM to allocate to the client
# compared to the total memory available on the host.
# CPU overcommit ratio is the ratio of CPU to allocate to the client
# compared to the total CPU available on the host.
cpu-overcommit-ratio: 3.0
mem-overcommit-ratio: 1.0

# Image syncer configuration options. (optional)
image-syncer:
# The interval at which to sync images. 
  interval: 1m
# The maximum number of images to sync concurrently.
  max-concurrent: 10
# List of images to sync. Leave empty to sync all images.
  images:
  - 48233fc0-8c16-491b-8666-970ba3ce771e
  - ubuntu-22.04

enable-image-gc: true
image-gc:
# The interval at which to garbage collect images.
  interval: 1m

# Log level for the agent. Must be one of: debug, info, warn, error, fatal, panic, trace.
log-level: debug
```

## Server configuration

Example configuration for a server:

```yaml
---
# Listen address for the HTTP server. This is where the GitHub webhook should be configured to send events.
listen-addr: 0.0.0.0:8080

# Data directory for the server. This is where the server will store its state.
data-dir: /var/lib/fireactions

# GitHub configuration options.
github:
# Job label prefix to search for in received GitHub events.
  job-label-prefix: fireactions
# The secret used to verify GitHub webhook payloads.
  webhook-secret: SECRET
# The GitHub App ID and PEM-encoded private key.
# See: https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#generating-a-private-key
  app-id: 123456
  app-private-key: |
    -----BEGIN RSA PRIVATE KEY-----

# Scheduler configuration options (optional).
scheduler:
# Multiplier for the CPU and RAM scores. This is used to adjust the scheduler's preference for free CPU and RAM.
  free-cpu-scorer-multiplier: 1.0
  free-ram-scorer-multiplier: 1.0

# The default flavor to use for jobs if no flavor is specified.
default-flavor: 1vcpu-1gb

# Flavors are used to define the resources available to a job. Atleast one flavor must be defined.
# The name of the flavor must be unique.
# The disk size is in GB, the memory size is in MB, and the CPU count is the number of vCPUs.
flavors:
  - name: 1vcpu-1gb
    disk: 50
    image: ubuntu-22.04
    mem: 1024
    cpu: 1
  - name: 2vcpu-2gb
    disk: 50
    image: ubuntu-22.04
    mem: 2048
    cpu: 2
  - name: 4vcpu-4gb
    disk: 50
    image: ubuntu-22.04
    mem: 4096
    cpu: 4

# The default group to use for jobs if no group is specified in GitHub job label.
# The group must be defined in the 'groups' section.
default-group: us-east-2

# Groups are used to separate clients into logical groups, e.g. by region, datacenter, etc. Atleast one group must be defined.
groups:
  - name: us-east-2
  - name: us-west-1
    enabled: false # (optional) Whether the group is enabled or not. Defaults to true.

# Images are virtual machine disk images that can be used to create MicroVMs with Firecracker. The images are synced
# by the clients. Atleast one image must be defined.
# The ID of the image must be unique, otherwise it will be overwritten by the next image with the same ID.
images:
  - id: 48233fc0-8c16-491b-8666-970ba3ce771e
    name: ubuntu-22.04
    url: https://storage.googleapis.com/fireactions/images/ubuntu-22.04.ext4
  - id: aa01e575-4259-48ed-aa24-f9885a67a11a
    name: ubuntu-20.04
    url: https://storage.googleapis.com/fireactions/images/ubuntu-20.04.ext4

# Log level. Must be one of: debug, info, warn, error, fatal, panic, trace.
log-level: debug
```
