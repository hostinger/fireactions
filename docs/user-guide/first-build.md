# Running Your First Build

After installing and configuring Fireactions, verify your setup by running a test workflow.

## Verify Runners Are Registered

Check your GitHub organization's Actions settings to confirm runners are registered:

1. Navigate to your GitHub organization settings
2. Go to **Actions** → **Runners**
3. Verify that runners are listed as **Idle** and ready to receive jobs

If runners aren't appearing, check the Fireactions logs:

```bash
sudo journalctl -u fireactions -f
```

## Create a Test Workflow

Create a simple workflow to test your Fireactions setup. In your repository, create `.github/workflows/test-fireactions.yml`:

```yaml
name: Test Fireactions

on:
  workflow_dispatch:
  push:
    branches:
      - main
  pull_request:

jobs:
  test:
    name: Test Runner
    runs-on: fireactions-example  # Replace with your pool label
    steps:
      - name: Check runner environment
        run: |
          echo "Runner is working!"
          uname -a
          docker --version
```

**Important:** Replace `fireactions-example` with the label from your [pool configuration](../reference/configuration.md).

## Run the Workflow

Trigger the workflow using one of these methods:

- **Manual trigger:** Go to Actions tab → Select workflow → Click "Run workflow"
- **Push to main:** Commit and push changes to the main branch
- **Pull request:** Open a pull request

## Verify Execution

Watch the workflow run in GitHub Actions:

1. Go to the **Actions** tab in your repository
2. Click on the workflow run
3. Verify the job completes successfully
4. Check that it ran on a Fireactions runner

## Expected Behavior

When the workflow runs:

1. Fireactions creates a new Firecracker microVM
2. The GitHub runner inside the VM picks up the job
3. Job executes in the isolated environment
4. VM is destroyed after job completion

You should see the workflow complete in ~20-30 seconds from trigger to finish.

## Troubleshooting

If the workflow doesn't run or fails:

- Verify pool labels match between configuration and workflow
- Check Fireactions logs for errors
- Ensure sufficient system resources (CPU, memory, disk)
- See the [Troubleshooting Guide](../help/troubleshooting.md) for common issues
