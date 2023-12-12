![Banner](banner.png)

Fireactions is an orchestrator for GitHub runners. BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure [Firecracker](https://firecracker-microvm.github.io/) based virtual machines.

![Architecture](architecture.png)

Several key features:

- **Autoscaling**

  Robust pool based scaling, cost-effective with fast GitHub runner startup time of 20s~.

- **Ephemeral**

  Each virtual machine is created from scratch and destroyed after the job is finished, no state is preserved between jobs, just like with GitHub hosted runners.

- **Customizable**

  Define job labels and customize virtual machine resources to fit Your needs. See [Configuration](./user-guide/configuration.md) for more information.
