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
  job-label-prefix: fireactions-
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

# Log level. Must be one of: debug, info, warn, error, fatal, panic, trace.
log-level: debug
```
