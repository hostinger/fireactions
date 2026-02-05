# Debugging workflows

Occasionally, it may be necessary to connect to a running Fireactions VM for debugging workflow steps or inspecting the filesystem.

[tmate](https://github.com/mxschmitt/action-tmate) action provides a way to SSH into a running runner VM and have full access.

Using it is easy - just decide which workflow step to intercept, comment out the subsequent steps, then insert the tmate action just before them:

```
    - name: Setup tmate session
      uses: mxschmitt/action-tmate@v3
```

!!! Warning
      For security purposes, it's advised to add SSH keys to your github [profile](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/adding-a-new-ssh-key-to-your-github-account) and limit access to yourself:
```
    - name: Setup tmate session
      uses: mxschmitt/action-tmate@v3
      with:
        limit-access-to-actor: true
```

Additionally, instead of having to add/remove, or uncomment the required config, you can make the tmate step to be conditional and use user input:

```
on:
  workflow_dispatch:
    inputs:
      debug_enabled:
        type: boolean
        description: 'Run the build with tmate debugging enabled (https://github.com/marketplace/actions/debugging-with-tmate)'
        required: false
        default: false
```

```
    steps:
      - name: Setup tmate session
        uses: mxschmitt/action-tmate@v3
        if: ${{ github.event_name == 'workflow_dispatch' && inputs.debug_enabled }}
```
