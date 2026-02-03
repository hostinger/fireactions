# Images

Fireactions images are OCI compliant Docker images that are used to run GitHub Actions runner in Firecracker microVM. The images are built using Docker and contain all the necessary tools and dependencies.

Each image must contain the Fireactions binary and `/etc/systemd/system/fireactions-agent.service` file:

```systemd
[Unit]
Description=Fireactions Agent
Documentation=https://github.com/hostinger/fireactions
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/bin/fireactions agent --log-level=info
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

The Fireactions agent is started as a systemd service when the container is run. The Fireactions agent manages the lifecycle of GitHub runner inside the Firecracker microVM. Once the workflow job completes (or GitHub runner exits), the Fireactions agent will shut down the Firecracker microVM.

> Optionally, the shutdown can be disabled in order to keep the microVM running for debugging purposes using the `shutdown_on_exit` option of a Pool.

## Available Images

The following images are available [in this repository](https://github.com/hostinger/fireactions-images):

| Name | Description | OS |
|------|-------------|----|
| ubuntu20.04 | Full Ubuntu 20.04 image with Docker, Docker Compose, and other tools | Ubuntu 20.04 |
| ubuntu22.04 | Full Ubuntu 22.04 image with Docker, Docker Compose, and other tools | Ubuntu 22.04 |
| ubuntu24.04 | Full Ubuntu 24.04 image with Docker, Docker Compose, and other tools | Ubuntu 24.04 |

To build a custom image, see the [custom image example](../examples/custom-image.md)
