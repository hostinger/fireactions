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

# Name of the client. This is used to identify the client on the server.
# Defaults to the $HOSTNAME environment variable.
name: ""

# CPU and RAM overcommit ratios. These are used to calculate the resources required for a job.
cpu_overcommit_ratio: 1.0
ram_overcommit_ratio: 1.0

# Labels to apply for the client. These are used for affinity rules.
labels:
  fireactions/region: default

# The client will reconcile the state of the server with the local state every `reconcile_interval` seconds.
reconcile_interval: 5s

# Log level. Can be one of: debug, info, warn, error, fatal, panic, trace.
log_level: debug
```

## Server configuration

Example configuration for a server:

```yaml
---
# Path to the directory where the server will store its data.
data-dir: /var/lib/fireactions

# HTTP server configuration.
http:
# The address to listen on for HTTP requests.
  listen_addr: 0.0.0.0:8081

# GitHub configuration.
github:
# The GitHub webhook secret. This should be the same as the one configured in the GitHub App settings.
  webhook_secret: SECRET
# The GitHub App ID and key. This can be found in the GitHub App settings.
  app_id: 123456
  app_private_key: |
    -----BEGIN RSA PRIVATE KEY-----

# Metrics server configuration. This is used to expose Prometheus metrics on endpoint `/metrics`.
metrics:
# Whether to enable the metrics server.
  enabled: true
# The address to listen on for HTTP requests.
  address: 0.0.0.0:8082

# List of job labels.
job_labels:
- name: fireactions-2vcpu-2gb
# List of allowed repositories. Regular expressions are supported.
  allowed_repositories:
  - ".*"
# Template for the GitHub runner name. The following variables are supported:
# - {{ .ID }}: The ID of the runner.
  runner_name_template: fireactions-2vcpu-2gb-{{ .ID }}
# Additional labels to apply to the GitHub runner. The default labels include only the job_label name.
  runner_labels:
  - ubuntu-22.04
# The GitHub runner image to use. This should be a Docker image that contains the GitHub runner and the Fireactions agent.
  runner_image: ghcr.io/hostinger/fireactions/runner:ubuntu-20.04-x64-2.310.2
# The GitHub runner image pull policy. Can be one of: Always, Never, IfNotPresent.
  runner_image_pull_policy: IfNotPresent
# The GitHub runner resources. These are used to calculate the resources required for a job.
  runner_resources:
    memory_mb: 2048
    vcpus: 2
# Affinity rules. These are used to schedule GitHub runners to specific clients. The following operators are supported:
# - NotIn: The value must not be in the list of values.
# - In: The value must be in the list of values.
  runner_affinity:
  - { key: fireactions/region, operator: In, values: [default] }
# Metadata to apply to the GitHub runner. Use with MMDS service to provide metadata to the virtual machine.
  runner_metadata:
    example1: value1
    example2: value2

# Log level. Can be one of: debug, info, warn, error, fatal, panic, trace.
log_level: debug
```
