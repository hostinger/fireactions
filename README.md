[![test](https://github.com/hostinger/fireactions/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/hostinger/fireactions/actions/workflows/test.yaml)

![Banner](docs/banner.png)

Fireactions is an orchestrator for GitHub runners. BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure [Firecracker](https://firecracker-microvm.github.io/) based virtual machines.

Several key features:

- **Autoscaling**

  Robust autoscaling based on GitHub webhooks for `workflow_job` events, cost-effective with fast GitHub runner startup time of 20s~.

- **Ephemeral**

  Each virtual machine is created from scratch and destroyed after the job is finished, no state is preserved between jobs, just like with GitHub hosted runners.

- **Customizable**

  Define Your own virtual machine Flavors, Groups, Images and more to match Your needs. See [Configuration](./docs/configuration.md) for more information.

## Quickstart

To start using self-hosted GitHub runners, add the label to your workflow jobs:

```yaml
<...>
runs-on:
- self-hosted
# e.g. fireactions.group1.1vcpu-1gb, fireactions.group1, fireactions
- <PREFIX>[.GROUP][.FLAVOR]
```

See [Configuration](./docs//configuration.md) for more information on how to configure the default job label prefix, groups and flavors.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to contribute to Fireactions.

## License

See [LICENSE](LICENSE)
