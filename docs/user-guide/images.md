# Images

Fireactions images are OCI compliant Docker images that are used to run GitHub Actions runner in Firecracker microVM. The images are built using Docker and contain all the necessary tools and dependencies.

Each image must contain the Fireactions binary and `systemd` service file:

```systemd
[Unit]
Description=Fireactions
Documentation=https://github.com/hostinger/fireactions
After=network.target
SuccessAction=reboot

[Service]
Type=simple
User=root
ExecStart=/usr/bin/fireactions runner --log-level=info
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

The Fireactions binary is started as a systemd service when the container is run. The `SuccessAction` option is used to reboot the microVM when the Fireactions binary exits successfully, forcing the microVM to be recreated for the next job.

## Available Images

The following images are available [in this repository](https://github.com/hostinger/fireactions-images):

| Name | Description | OS |
|------|-------------|----|
| ubuntu20.04 | Full Ubuntu 20.04 image with Docker, Docker Compose, and other tools | Ubuntu 20.04 |
| ubuntu22.04 | Full Ubuntu 22.04 image with Docker, Docker Compose, and other tools | Ubuntu 22.04 |

To build a custom image, see the [custom image example](../examples/custom-image.md)
