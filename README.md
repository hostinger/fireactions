# fireactions [![test](https://github.com/hostinger/fireactions/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/hostinger/fireactions/actions/workflows/test.yaml)

<img src="./docs/logo.png" alt="logo" width="150"/>

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
- actions-<GROUP>-<FLAVOR>
```

Job labels identify the type of virtual machine to create for the Job. Label must begin with prefix `actions-` and must be followed by the group and flavor name. If the flavor name is not specified, the default configured flavor will be used.

The label format is based on the following template:

`<PREFIX>-<GROUP>-<FLAVOR>`

Flavors and groups are defined in the `flavors` and `groups` section of the [configuration](./docs/configuration.md) file. The default flavor and group can be specified in the `default-flavor`, `default-group` options.

## Metrics

[Prometheus](https://prometheus.io/) metrics are exposed on the same port and can be accessed at `/metrics` endpoint.

### Server

Currently, the following metrics are exposed:

| Metric name                | Description                                                                 |
|----------------------------|-----------------------------------------------------------------------------|
| fireactions_server_up | Whether the Fireactions server is up and running. |
| fireactions_store_scrape_errors_total | Total number of errors encountered while scraping the store. |
| fireactions_store_resources_total | Number of resources in the store (nodes, jobs, runners, groups, flavors) |

Example Prometheus scrape configuration:

```yaml
scrape_configs:
- job_name: fireactions-server
  static_configs:
  - targets:
      - 127.0.0.1:8080
```

## Roadmap

- Support for right-sizing virtual machines based on actual (historical) GitHub job resource usage via [Prometheus](https://prometheus.io/)

## License

[Apache License, Version 2.0](LICENSE)
