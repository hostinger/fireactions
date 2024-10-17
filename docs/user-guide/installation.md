# Installation

Ansible is an open-source automation tool that allows you to automate the deployment and configuration of software on multiple servers. It is a powerful tool that can help you manage your infrastructure more efficiently.

It is the recommended way to install Fireactions. This guide will show you how to install Fireactions using Ansible.

## Pre-requisites

Before you begin, make sure you have the following requirements:

- Ansible set up on local machine
- GitHub App ID and private key, see [Creating GitHub Apps](https://docs.github.com/en/apps/creating-github-apps)

## Step 1: Create an inventory file

```yaml
# hosts.ini
[all]
fireactions-server-1 ansible_host=<IP_ADDRESS> ansible_user=<SSH_USER>
```

## Step 2: Create an Ansible playbook

```yaml
# site.yaml
---
- name: Install Fireactions
  hosts:
  - all
  become: yes
  - role: hostinger.fireactions.fireactions
    vars:
      fireactions_version: 0.2.3
      fireactions_config:
        bind_address: 0.0.0.0:8080
        metrics:
          enabled: true
          address: 127.0.0.1:8081
        github:
          app_id: <APP_ID>
          app_private_key: |
            <APP_PRIVATE_KEY>
        debug: true
        pools:
        - name: example
          max_runners: 10
          min_runners: 1
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

The `fireactions_config` variable contains the configuration for Fireactions. You can customize the configuration to suit your needs. For more information on the available configuration options, see the [configuration](../user-guide/configuration.md)

## Step 3: Create a requirements file

```yaml
# requirements.yaml
---
collections:
  - name: hostinger.fireactions
    version: 0.1.1
  - name: hostinger.common
    version: 0.8.0
```

## Step 4: Install the required Ansible collections

```bash
ansible-galaxy collection install -r requirements.yaml
```

## Step 5: Run the Ansible playbook

```bash
ansible-playbook -i hosts.ini --diff site.yaml
```

The Ansible playbook will do the following:

- Install [Containerd](https://github.com/containerd/containerd)
- Install [CNI plugins](https://github.com/containernetworking/plugins) (bridge, host-local, firewall, [tc-redirect-tap](https://github.com/awslabs/tc-redirect-tap))
- Configure CNI networking
- Install [Firecracker](https://github.com/firecracker-microvm/firecracker)
- Setup required sysctl settings
- Install Fireactions

After the playbook has finished, check the status of the Fireactions service:

```bash
$ systemctl status fireactions
● fireactions.service - Fireactions
     Loaded: loaded (/etc/systemd/system/fireactions.service; enabled; preset: enabled)
     Active: active (running) since Sun 2024-09-08 18:59:08 UTC; 2 days ago
       Docs: https://github.com/hostinger/fireactions
    Process: 3564 ExecStartPre=/usr/bin/which firecracker (code=exited, status=0/SUCCESS)
    Process: 3566 ExecStartPre=/usr/bin/which containerd (code=exited, status=0/SUCCESS)
   Main PID: 3571 (fireactions)
      Tasks: 480 (limit: 618568)
     Memory: 87.8G
        CPU: 5d 6h 21min 36.143s
```

At this point, Fireactions should be up and running. You can now proceed to [running your first build](../user-guide/running-the-first-build.md).
