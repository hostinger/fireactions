# Running the first build

Once everything is configured and Fireactions is running, check the status of the registered GitHub runners in GitHub.

If everything is working correctly, the GitHub runners should be marked as Idle and ready to receive jobs.

## Creating a new GitHub workflow

To create a new GitHub workflow, you need to create a new file in the `.github/workflows` directory of your repository. The file should have a `.yml` extension and contain the following content:

```yaml
name: test

on:
  workflow_dispatch:
  pull_request:
      branches:
      - '*'
  push:
      branches:
      - main

jobs:
  test:
    name: test
    runs-on: fireactions-example # The label(s) of the Fireactions pool
    steps:
    - name: Example
      run: |
        echo "Hello, world!"
```

This workflow will run on every push to the `main` branch, every pull request, and every manual trigger. The job will run on the `fireactions-example` pool, which is the label of the pool that you have created in the previous steps, while [configuring Fireactions](../user-guide/configuration.md)

## Triggering the workflow

To trigger the workflow, you can push a new commit to the `main` branch, create a new pull request, or manually trigger the workflow from the GitHub Actions UI.

The workflow job will be picked up by the GitHub runner and executed on the Fireactions pool that you have configured.
