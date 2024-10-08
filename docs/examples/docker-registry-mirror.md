# Setting up a Docker registry mirror

Using DockerHub directly to pull images can be a pain due to rate limits. To avoid this, there's an option of setting up a custom registry mirror, or in other words a pull-through cache. 

This way, you can pull images from the mirror instead of DockerHub, which can be faster and more reliable.

To set up the Docker registry mirror, include the following Ansible role in the Ansible playbook that you've used to install Fireactions:

```yaml
- role: hostinger.common.registry
  vars:
    registry_name: docker.io
    registry_config: "{{ registry_config_docker_io }}"
  tags:
    - registry_docker_io
    - registry
```

The `registry_config_docker_io` variable should be defined in the `group_vars/all.yaml` file. Here's an example of how it can look like:

```yaml
registry_config_docker_io:
  version: 0.1
  http:
    addr: 192.168.128.1:5003 # fireactions-br0 network
    relativeurls: false
    draintimeout: 60s
  storage:
    filesystem:
      rootdirectory: /var/lib/registry/docker.io
  proxy:
    remoteurl: https://registry-1.docker.io
  log:
    level: info
    formatter: text
    accesslog:
      disabled: false
```

Then run the Ansible playbook to apply the changes:

```bash
ansible-playbook -i <inventory> <playbook>.yml --tags registry_docker_io
```

This configuration will set up a registry mirror for `docker.io` images. The mirror will be available at `http://192.168.128.1:5003`.

After setting up the Docker registry mirror, configure the GitHub workflow to use the mirror:

```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3
  with:
    install: true
    driver: docker-container
    buildkitd-flags: --config /etc/buildkit/buildkitd.toml
    config-inline: |
      [registry."docker.io"]
        mirrors = ["192.168.128.1:5003"]
        http = true
        insecure = true

- name: Pull image
  run: |
    docker pull alpine:latest
```

To check if it worked, run the following command:

```bash
curl --silent http://192.168.128.1:5003/v2/_catalog
```

If the output is similar to the following, congratulations, the Docker registry mirror is set up correctly!

```json
{"repositories":["library/alpine","moby/buildkit"]}
```
