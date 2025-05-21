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

Pool management commands:
  resume      Resume a paused pool, enabling it to scale up again
  pause       Pause a pool, preventing it from scaling up
  scale       Scale a pool to specified number of replicas
  show        Retrieve a specific pool by name
  list        List all pools

MicroVM management commands:
  microvm     Manage MicroVMs within a pool

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

If the Fireactions server is configured with basic authentication, user must include the username and password using the `--username` and `--password` flags.

## Commands

### `runner`

Starts the virtual machine runner. This command should be run inside the virtual machine.

### `server`

Starts the server.

### `resume <NAME>`

Resume a paused pool, enabling it to scale up again.

### `pause <NAME>`

Pause a pool, preventing it from scaling up.

### `scale <NAME> [--replicas=<REPLICAS>]`

Scale a pool to specified number of replicas.

### `show <NAME>`

Retrieve a specific pool by name.

### `list`

List all pools.

### `reload`

Reload the server with the latest configuration (no downtime).

## `microvm`

### `list <POOL-NAME>`

List all MicroVMs in the specified pool.

### `login <VM-ID>`

Display a command to SSH into a MicroVM (along with its details) by specified VM ID.
