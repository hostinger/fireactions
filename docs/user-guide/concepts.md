# Concepts

## Pool

Pool is a group of GitHub runners that share the same labels and Firecracker virtual machine configuration. Each pool can have a number of replicas for the GitHub runners. Fireactions will automatically scale the number of GitHub runners to match the desired number of replicas.

There can be multiple pools configured, each with different configurations. For example, you can have a pool with runners that have 2 vCPUs and 2 GB of RAM, and another pool with runners that have 4 vCPUs and 4 GB of RAM, each with different labels.

Pools can be paused via CLI, which prevents it from scaling up. This can be useful when you want to prevent new runners from being created, but you don't want to delete the existing runners.

Pools are configured in the `pools` section of the configuration file, e.g.:

```yaml
pools:
- name: example
  replicas: 5
  shutdown_on_exit: true
  runner:
    name: example
    image: <IMAGE>:<IMAGE_TAG>
    image_pull_policy: IfNotPresent
    group_id: 1
    organization: hostinger
    labels:
    - self-hosted
    - fireactions
  firecracker:
    binary_path: firecracker
    kernel_image_path: /usr/local/share/firecracker/vmlinux.bin
    kernel_args: "console=ttyS0 noapic reboot=k panic=1 pci=off nomodules rw"
    machine_config:
      mem_size_mib: 1024
      vcpu_count: 1
    metadata:
      example: example
```

This will create a pool named `example` with 5 replicas for the GitHub runners. The runners will have the labels `self-hosted` and `fireactions`, and will use the specified Firecracker configuration.
