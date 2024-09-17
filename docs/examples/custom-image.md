# Creating custom image for Firecracker VMs

Fireactions allows you to use custom image for Firecracker virtual machines.

Use cases:

- Pre-installing software that is often used in the organisation.
- Using a custom OS
- Pre-configuring the VM with a specific configuration

Fireactions base images can be found in the [fireactions-images](https://github.com/hostinger/fireactions-images) GitHub repository. You can use them as a base for your custom image. Creating a custom image is as simple as creating a Dockerfile and building it:

```Dockerfile
# Use the base image
FROM --platform=linux/amd64 ghcr.io/hostinger/fireactions-images/ubuntu22.04:v0.5.1

# Install software, e.g. octopilot
COPY --from=ghcr.io/dailymotion-oss/octopilot:v1.6.0 /usr/local/bin/octopilot /usr/local/bin/octopilot
```

Build the image and push it to container registry:

```bash
docker build -t my-custom-image . && docker push my-custom-image
```

The last step is using the custom image in Fireactions by specifying it in the configuration file. The container registry must be accessible from the Fireactions server, so make sure to configure the credentials (optional).
