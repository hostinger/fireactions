# Getting started

Putting it simply, Fireactions is an orchestrator for GitHub runners.

## Why Fireactions?

Essentially, we needed a reliable and fast way to run self-hosted GitHub runners on our own infrastructure. We wanted to have the same level of control and security as with GitHub hosted runners, but without the limitations of the hosted runners (e.g. long startup times, limited resources, etc.).

Fireactions is designed to be a cost-effective, fast and secure solution for running self-hosted GitHub runners. It is built on top of [Firecracker](https://firecracker-microvm.github.io/) microVMs, which provide a lightweight and secure environment for running workloads.

## Why Firecracker?

Firecracker is a lightweight virtual machine monitor (VMM) that uses the Linux Kernel-based Virtual Machine (KVM) to create and manage microVMs. MicroVMs are lightweight, fast and secure virtual machines that are designed to run a single application or service.

Compared to containers, microVMs provide a higher level of isolation and security, as each microVM runs in its own isolated environment. This makes them ideal for running untrusted workloads, such as CI/CD jobs.

## Is it ready for production?

Fireactions is still in the early stages of development, we are waiting for feedback from the community to improve it further. However, we are already using it fully in production and it has been working well for us so far.

## Requirements

There are a few requirements to run Fireactions:

- Linux machine with KVM support. We recommend using a machine with at least 2 CPU cores and 4GB of RAM.
- GitHub organisation account with PAT or GitHub App installation token.
- Containerd
- Firecracker v1.4.1 or newer

## Quickstart

There are two main ways of installing Fireactions: [manual installation](./manual-installation.md) and [Ansible installation](./ansible-installation.md).

We recommend using the Docker installation method, as it is the easiest and most convenient way to get started.
