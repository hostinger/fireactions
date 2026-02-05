# Frequently Asked Questions

## How does Fireactions compare to other solutions?

Self hosted GitHub runners are a great way to run your CI/CD jobs on your own infrastructure. However, setting up and managing self-hosted runners can be a complex and time-consuming process. Fireactions aims to simplify this process by providing a lightweight and secure solution for orchestrating self-hosted GitHub runners.

Compared to [ARC(Actions Runner Controller)](https://github.com/actions/actions-runner-controller), ARC is a Kubernetes operator with primary use case of managing self-hosted GitHub runners in Kubernetes clusters, while Fireactions is a standalone application that can be run on any regular bare metal server or a VM that supports nested virtualization.

Kubernetes can be an overkill for most situations and managing a Kubernetes cluster can be complex and resource intensive task.

## Do I need to use a bare metal server or a VM to run Fireactions?

Fireactions can be run on any regular bare metal server or a VM that supports nested virtualization. Do note that running Fireactions on a VM might have performance implications, especially if you are running multiple concurrent jobs.

## Is there GPU support in Fireactions?

Currently, Fireactions **does not support GPU workloads**. Fireactions uses [Firecracker](https://firecracker-microvm.github.io/) under the hood and it does not have GPU support, but it is in the [roadmap](https://github.com/firecracker-microvm/firecracker/discussions/4845). We are actively tracking this feature and will add support for GPU workloads as soon as it is available.
