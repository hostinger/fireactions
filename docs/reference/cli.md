# Interacting with Fireactions via CLI

Fireactions provides a CLI for interacting with the server.

```bash
$ fireactions --help
BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure Firecracker based virtual machines.

Usage:
  fireactions [command]

Main application commands:
  server      Starts the Fireactions server
  agent       Starts the Fireactions agent
  validate    Validates the server configuration file

Pool management commands:
  pools       Manage pools

Machine management commands:
  ps          List all running machines across all pools
  login       SSH into a running VM as root user
  logs        Stream logs from the fireactions-agent service inside a machine

Image management commands:
  image       Manage images

Additional Commands:
  version     Show version information
  help        Help about any command
  completion  Generate the autocompletion script for the specified shell

Flags:
  -h, --help      help for fireactions
  -v, --version   version for fireactions

Use "fireactions [command] --help" for more information about a command.
```

## Server Endpoint

Most commands that interact with the Fireactions server accept an `--endpoint` (or `-e`) flag to specify the server address.

```bash
fireactions --endpoint https://fireactions.example.com:8080 pools list
# or using shorthand
fireactions -e https://fireactions.example.com:8080 pools list
```

The default endpoint is `127.0.0.1:8080`.

## Commands

### Main Application Commands

#### `server`

Starts the Fireactions server.

```bash
fireactions server
```

By default, the server looks for a configuration file at `/etc/fireactions/config.yaml`. You can specify a different path using the `--config` (or `-f`) flag:

```bash
fireactions server --config /path/to/config.yaml
```

#### `agent`

Starts the Fireactions agent. This command should be run inside the virtual machine and is automatically executed by the VM image.

```bash
fireactions agent
```

The agent reads metadata from the Firecracker MMDS service to configure itself. You can optionally specify the log level:

```bash
fireactions agent --log-level debug
```

Available log levels: `debug`, `info`, `warn`, `error`, `fatal`, `panic`, `trace` (default: `info`)

#### `validate`

Validates a server configuration file without starting the server. This is useful for checking configuration syntax and validating settings before deployment.

```bash
fireactions validate /path/to/config.yaml
```

### Pool Management Commands

All pool management commands accept an `--endpoint` (or `-e`) flag to specify the server address (default: `127.0.0.1:8080`).

#### `pools list` (alias: `pools ls`)

List all configured pools with their current status.

```bash
fireactions pools list
```

#### `pools pause <NAME>`

Pause a pool, preventing it from scaling up. Running VMs continue to operate, but no new VMs will be started.

```bash
fireactions pools pause default
```

#### `pools resume <NAME>`

Resume a paused pool, enabling it to scale up again.

```bash
fireactions pools resume default
```

#### `pools scale <NAME> --replicas <N>`

Scale a pool to the specified number of replicas. The pool will scale up or down to match the desired number.

```bash
# Scale to 5 replicas
fireactions pools scale default --replicas 5

# Scale down to 0 (stop all VMs)
fireactions pools scale default --replicas 0
```

**Note**: The `--replicas` flag is required and you can scale down to 0 to stop all VMs in a pool.

### Machine Management Commands

All machine management commands accept an `--endpoint` (or `-e`) flag to specify the server address (default: `127.0.0.1:8080`).

#### `ps` (alias: `ls`)

List all running machines across all pools.

```bash
fireactions ps
```

#### `login <VMID>`

SSH into a running VM as the root user. This is useful for debugging or inspecting VM state.

```bash
fireactions login default-abc123
```

The command will automatically:
- Look up the VM's IP address
- Establish an SSH connection with appropriate options
- Drop you into a root shell

**Requirements**: SSH must be installed and accessible in your PATH.

#### `logs <MACHINE_ID>`

Stream logs from the fireactions-agent gRPC service running inside a machine. This shows the zerolog output from the agent service itself, including agent startup, status changes, and any errors from the agent.

```bash
# Show all buffered logs
fireactions logs default-abc123

# Follow logs in real-time (like tail -f)
fireactions logs default-abc123 --follow

# Show last 50 lines and follow
fireactions logs default-abc123 --follow --tail 50
```

**Flags:**
- `-f, --follow`: Follow log output (stream continuously like tail -f)
- `--tail N`: Number of lines to show from end (0 = all buffered logs)

### Image Management Commands

All image management commands accept an `--endpoint` (or `-e`) flag to specify the server address (default: `127.0.0.1:8080`).

#### `image list` (alias: `image ls`)

List all container images managed by Fireactions.

```bash
fireactions image list
```

#### `image remove <NAME>` (alias: `image rm`)

Remove a container image from the Fireactions server.

```bash
fireactions image remove ghcr.io/myorg/myimage:latest
```

### Additional Commands

#### `version`

Show version information for Fireactions.

```bash
fireactions version
```

## Examples

### Basic Workflow

```bash
# Validate configuration before starting
fireactions validate /etc/fireactions/config.yaml

# Start the server
fireactions server --config /etc/fireactions/config.yaml

# List all pools
fireactions pools list

# Scale up a pool
fireactions pools scale production --replicas 10

# Check machine status
fireactions ps

# View logs from a specific machine
fireactions logs production-abc123 --follow

# SSH into a machine for debugging
fireactions login production-abc123

# Scale down when done
fireactions pools scale production --replicas 0

# List images
fireactions image list

# Remove an unused image
fireactions image remove ghcr.io/example/old-image:v1
```

### Using Remote Server

```bash
# Connect to a remote Fireactions server
fireactions -e https://fireactions.example.com:8080 pools list

# All commands support the --endpoint flag
fireactions -e https://fireactions.example.com:8080 ps
fireactions -e https://fireactions.example.com:8080 logs machine-123 -f
```
