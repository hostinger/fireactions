# GitHub App

Fireactions authenticates with GitHub as a [GitHub App](https://docs.github.com/en/apps/creating-github-apps). The App is
the only credential Fireactions needs: it creates just-in-time (JIT) runner registrations before starting a Firecracker
VM, and deletes the runner from GitHub after the VM exits.

The App is configured in the `github` section of the configuration file:

```yaml
github:
  app_id: 12345
  app_private_key: |
    -----BEGIN RSA PRIVATE KEY-----
    ...
    -----END RSA PRIVATE KEY-----
```

Which permissions the App needs depends on the [runner scope](concepts.md#runner-scope) of your pools.

## Permissions

### Organization scoped pools

For pools that set `runner.organization`, grant the following **organization permission**:

| Permission          | Access           |
|---------------------|------------------|
| Self-hosted runners | Read and write   |

This is the permission that covers `POST /orgs/{org}/actions/runners/generate-jitconfig` and
`DELETE /orgs/{org}/actions/runners/{runner_id}`. The organization `Administration` permission does **not** grant access
to these endpoints, so granting it instead won't work.

No repository permissions are needed: runners are registered with the organization itself, not with any of its
repositories.

### Repository scoped pools

For pools that set `runner.repository`, grant the following **repository permissions**:

| Permission     | Access                            |
|----------------|-----------------------------------|
| Administration | Read and write                    |
| Metadata       | Read-only (granted automatically) |

`Administration` covers `POST /repos/{owner}/{repo}/actions/runners/generate-jitconfig` and
`DELETE /repos/{owner}/{repo}/actions/runners/{runner_id}`. GitHub grants `Metadata: Read-only` automatically as soon as
any repository permission is selected.

Repository scoped pools are the only option for personal (user) accounts, which can't have organization runners. They
work for organization owned repositories too.

!!! note
    A single App can serve both kinds of pools. Grant the organization `Self-hosted runners` permission and the
    repository `Administration` permission if you run organization scoped and repository scoped pools side by side.

### Not required

Fireactions does not need any of the following, and you can leave them unset:

- **Webhooks and webhook events.** Fireactions keeps each pool at its configured number of replicas on its own and does
  not consume `workflow_job` or any other event, so the App needs no webhook URL and no event subscriptions.
- **The `Actions` permission.** Fireactions never reads or writes workflow runs, jobs or artifacts.
- **Account (user) permissions.**

## Creating and installing the App

The App can be owned by an organization or by a personal account. Ownership only decides who administers the App; what
matters for Fireactions is the account the App is *installed* on.

1. Create the App:
    - **Organization owned:** `https://github.com/organizations/<ORG>/settings/apps/new`
    - **Personal account owned:** `https://github.com/settings/apps/new`
2. Set a name and a homepage URL. Uncheck **Active** under **Webhook** — Fireactions doesn't use webhooks.
3. Under **Permissions**, grant the permissions for your runner scope from the tables above.
4. Create the App, then note the **App ID** and generate a **private key**. The downloaded `.pem` file is what goes into
   `github.app_private_key`.
5. Install the App:
    - **On an organization** (for organization scoped pools): **Install App** → choose the organization.
    - **On a personal account** (for repository scoped pools): **Install App** → choose your account, then
      **Only select repositories** and pick the repositories your pools register runners with.

!!! warning
    Changing an App's permissions after installation requires accepting the new permissions on the installation before
    they take effect. If Fireactions starts failing with `403` errors after a permission change, check the installation
    settings page for a pending request.

## Installation ID

Fireactions needs the ID of the App installation on the account owning the runners. It looks the ID up automatically
using the App's own JWT — `GET /orgs/{org}/installation` for organization scoped pools and
`GET /repos/{owner}/{repo}/installation` for repository scoped pools. Neither call needs an installation permission, so
in most setups nothing else is required.

Set `runner.installation_id` to skip the lookup:

```yaml
runner:
  repository: octocat/hello-world
  installation_id: 12345678
```

You can find the ID in the URL of the installation's settings page:

- Organization: **Settings** → **GitHub Apps** → **Configure**, the URL ends in
  `/organizations/<ORG>/settings/installations/<INSTALLATION_ID>`
- Personal account: **Settings** → **Applications** → **Configure**, the URL ends in
  `/settings/installations/<INSTALLATION_ID>`

Or via the API. This endpoint only accepts a JWT signed with the App's private key — a normal `gh auth login`
token won't work, so mint the JWT first and pass it explicitly:

```bash
APP_ID=<APP_ID>
KEY=<PATH_TO_PRIVATE_KEY.pem>

header=$(printf '{"alg":"RS256","typ":"JWT"}' | openssl base64 -A | tr '+/' '-_' | tr -d '=')
now=$(date +%s)
payload=$(printf '{"iat":%d,"exp":%d,"iss":"%s"}' "$((now - 60))" "$((now + 540))" "$APP_ID" \
  | openssl base64 -A | tr '+/' '-_' | tr -d '=')
sig=$(printf '%s.%s' "$header" "$payload" \
  | openssl dgst -sha256 -sign "$KEY" -binary | openssl base64 -A | tr '+/' '-_' | tr -d '=')
JWT="$header.$payload.$sig"

curl -sS -H "Authorization: Bearer $JWT" -H "Accept: application/vnd.github+json" \
  https://api.github.com/repos/<OWNER>/<REPOSITORY>/installation | jq .id
```

The same JWT works with `gh` if you'd rather use it, since `-H` overrides the token it would otherwise send:

```bash
gh api /repos/<OWNER>/<REPOSITORY>/installation -H "Authorization: Bearer $JWT" --jq .id
```

## API calls Fireactions makes

| Endpoint                                                    | Authenticated as | Permission                                |
|-------------------------------------------------------------|------------------|-------------------------------------------|
| `GET /orgs/{org}/installation`                              | App (JWT)        | None                                      |
| `GET /repos/{owner}/{repo}/installation`                    | App (JWT)        | None                                      |
| `POST /orgs/{org}/actions/runners/generate-jitconfig`       | Installation     | Organization `Self-hosted runners: write` |
| `DELETE /orgs/{org}/actions/runners/{runner_id}`            | Installation     | Organization `Self-hosted runners: write` |
| `POST /repos/{owner}/{repo}/actions/runners/generate-jitconfig` | Installation | Repository `Administration: write`        |
| `DELETE /repos/{owner}/{repo}/actions/runners/{runner_id}`  | Installation     | Repository `Administration: write`        |

The organization endpoints are only used by pools with `runner.organization` set, the repository endpoints only by pools
with `runner.repository` set.
