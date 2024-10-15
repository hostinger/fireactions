# Upgrading

## Upgrading the GitHub runner

GitHub runner is a component inside the Fireactions runner image. To upgrade it, the Fireactions runner image needs to be updated. The runner image is defined in the `runner.image` variable in the `pools` configuration.

Refer to the [changelog](https://github.com/actions/runner) for any breaking changes before the upgrade.

## Upgrading Fireactions

To upgrade Fireactions, override `fireactions_version` variable in your Ansible playbook:

```yaml
fireactions_version: 0.2.3
```

Then run the Ansible playbook:

```bash
ansible-playbook -i <inventory> --diff --tags fireactions <playbook>
```

Keep in mind, that this will restart the Fireactions process and cause a short downtime to GitHub runners. It's best to schedule the upgrade during off-peak hours.
