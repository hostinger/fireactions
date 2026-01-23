# Interacting with Fireactions via CLI

Fireactions provides a CLI for interacting with the server.

```bash
$ fireactions --help
BYOM (Bring Your Own Metal) and run self-hosted GitHub runners in ephemeral, fast and secure Firecracker based virtual machines.

Usage:
  fireactions [command]

Main application commands:
  runner      Starts the virtual machine runner. This command should be run inside the virtual machine.
  server      Start the server
  pools       Manage pools

MicroVM management commands:
  ps          List all running VMs across all pools
  login       SSH into a running VM as root user

Additional Commands:
  reload      Reload the server with the latest configuration (no downtime)

Flags:
  -e, --endpoint string   Endpoint to use for communicating with the Fireactions API. (default "http://127.0.0.1:8080")
  -u, --username string   Username to use for authenticating with the Fireactions API.
  -p, --password string   Password to use for authenticating with the Fireactions API.
  -h, --help              help for fireactions
  -v, --version           version for fireactions

Use "fireactions [command] --help" for more information about a command.
```

## Authentication

If the Fireactions server is configured with basic authentication, you must include the username and password using the `--username` and `--password` flags.

```bash
fireactions --username admin --password secret pools list
```

## Commands

### Main Application Commands

#### `server`

Starts the Fireactions server.

```bash
fireactions server
```

#### `runner`

Starts the virtual machine runner. This command should be run inside the virtual machine and is automatically executed by the VM image.

```bash
fireactions runner
```

### Pool Management Commands

#### `pools list` (alias: `pools ls`)

List all configured pools with their current status.

```bash
fireactions pools list
```

#### `pools show <NAME>`

Display detailed information about a specific pool.

```bash
fireactions pools show default
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

#### `pools scale <NAME> --replicas=<N>`

Scale a pool to the specified number of replicas. The pool will scale up or down to match the desired number.

```bash
# Scale to 5 replicas
fireactions pools scale default --replicas=5

# Scale down to 0 (stop all VMs)
fireactions pools scale default --replicas=0
```

**Note**: You can scale down to 0 to stop all VMs in a pool.

### MicroVM Management Commands

#### `ps` (alias: `ls`)

List all running VMs across all pools.

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

### Additional Commands

#### `reload`

Reload the server configuration without downtime. This allows you to update pool configurations without restarting the server or interrupting running VMs.

```bash
fireactions reload
```

**Note**: The server must be running for this command to work.

## Examples

### Basic Workflow

```bash
# List all pools
fireactions pools list

# Scale up a pool
fireactions pools scale production --replicas=10

# Check VM status
fireactions ps

# SSH into a VM for debugging
fireactions login production-abc123

# Scale down when done
fireactions pools scale production --replicas=0
```

### Using Authentication

```bash
fireactions -e https://fireactions.example.com -u admin -p secret pools list
```
