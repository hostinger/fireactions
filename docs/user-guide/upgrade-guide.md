# Upgrading

This guide covers the process of upgrading Fireactions to a newer version.

## Upgrade Steps

### 1. Stop the Fireactions service

Stop the Fireactions service to prevent new runners from starting:

```bash
sudo systemctl stop fireactions
```

Verify the service is stopped:

```bash
sudo systemctl status fireactions
```

### 2. Download new Fireactions binary

Download the new release from the [GitHub releases page](https://github.com/hostinger/fireactions/releases):

```bash
# Example for version X.Y.Z
wget https://github.com/hostinger/fireactions/releases/download/vX.Y.Z/fireactions_X.Y.Z_linux_amd64.tar.gz
tar -xzf fireactions_X.Y.Z_linux_amd64.tar.gz
```

### 3. Replace the binary

Replace the old binary with the new one:

```bash
sudo mv fireactions /usr/local/bin/fireactions
sudo chmod +x /usr/local/bin/fireactions
```

### 4. Verify the binary

Confirm the new version is installed:

```bash
fireactions version
```

### 5. Validate configuration

Check your configuration for compatibility with the new version:

```bash
fireactions validate /etc/fireactions/config.yaml
```

If validation fails, review the error messages and update your configuration according to the release notes. Breaking changes are typically documented in the release notes with migration instructions.

### 6. Start the Fireactions service

Start Fireactions with the new version:

```bash
sudo systemctl start fireactions
```

## Post-Upgrade Verification

After starting the service, verify the upgrade was successful:

### Check Service Status

```bash
sudo systemctl status fireactions
```

The service should be `active (running)`.

### Review Logs

Check the logs for any errors or warnings:

```bash
sudo journalctl -u fireactions -n 100 -f
```

Look for specific error messages or warnings that might indicate issues.

### Verify Pool Status

List the pools to ensure they are running correctly:

```bash
fireactions pools ls
```

### Monitor Metrics

If metrics are enabled, check the metrics endpoint:

```bash
curl http://127.0.0.1:8081/metrics
```

### Test with a Workflow

Trigger a test GitHub workflow to verify runners are being created and jobs execute successfully.
