[![test](https://github.com/hostinger/fireactions/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/hostinger/fireactions/actions/workflows/test.yaml)

![Banner](docs/banner.png)

BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure [Firecracker](https://firecracker-microvm.github.io/) based virtual machines.

![Architecture](docs/architecture.png)

## Features

- Autoscaling based on received GitHub webhooks for `workflow_job` events.
- Support for multiple GitHub organisations
- Fast startup time of 15-30 seconds (from webhook received event to job running).
- Security by design: ephemeral virtual machines, no persistent storage. No need to worry about secrets left on the virtual machine.
- Support for `x86_64`, `arm64` architectures and multiple Linux kernel (LTS) versions.
- Support for multiple Linux distributions (Ubuntu 20.04 and Ubuntu 22.04).
- VM resource allocation (vCPUs, RAM) based on GitHub job labels.

## Usage

To start using self-hosted GitHub runners, add the label to your workflow jobs:

```yaml
<...>
runs-on:
- self-hosted
- <PREFIX>[.GROUP][.FLAVOR] # e.g. fireactions.group1.1vcpu-1gb, fireactions.group1, fireactions
```

Job labels identify the type of virtual machine to create for the Job. Label must begin with prefix and must be followed by the group and flavor name, separated by a dot. If neither group nor flavor is specified, the default group and flavor will be used.

See [Configuration](./docs//configuration.md) for more information on how to configure the default job label prefix, groups and flavors.

## Roadmap

- Support for right-sizing virtual machines based on actual (historical) GitHub job resource usage via [Prometheus](https://prometheus.io/)

## License

[Apache License, Version 2.0](LICENSE)
