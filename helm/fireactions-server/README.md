## Fireactions Server Helm Chart

<img src="./chart-icon.png" alt="logo" width="150"/>

## Introduction

This chart bootstraps a Fireactions server component as deployment/daemonset on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

### Prerequisites

- Kubernetes 1.17+ (recommended 1.20+)
- Helm 3.6+ (recommended 3.7+)

### Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| config | object | `{"dataDir":"/var/lib/fireactions","defaultFlavor":"1vcpu-1gb","defaultGroup":"us-east-2","flavors":[],"github":{"appId":"","appPrivateKey":"","existingSecret":"","jobLabelPrefix":"fireactions-","webhookSecret":""},"groups":[{"name":"us-east-2"},{"enabled":false,"name":"us-west-1"}],"listenAddr":"0.0.0.0:8080","logLevel":"info","scheduler":""}` | Fireactions server config values. |
| config.dataDir | string | `"/var/lib/fireactions"` | Data directory for the server. This is where the server will store its state. |
| config.defaultFlavor | string | `"1vcpu-1gb"` | The default flavor to use for jobs if no flavor is specified. |
| config.defaultGroup | string | `"us-east-2"` | The default group to use for jobs if no group is specified in GitHub job label. The group must be defined in the 'groups' section. |
| config.flavors | list | `[]` | Flavors are used to define the resources available to a job. Atleast one flavor must be defined. The name of the flavor must be unique. The disk size is in GB, the memory size is in MB, and the CPU count is the number of vCPUs. |
| config.github | object | `{"appId":"","appPrivateKey":"","existingSecret":"","jobLabelPrefix":"fireactions-","webhookSecret":""}` | GitHub configuration options. |
| config.github.appId | string | `""` | The GitHub App ID and PEM-encoded private key. See: https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#generating-a-private-key |
| config.github.existingSecret | string | `""` | Existing secret name to use. The secret must contain `webhook-secret` (optional), `app-id` and `app-private-key` keys. |
| config.github.jobLabelPrefix | string | `"fireactions-"` | Job label prefix to search for in received GitHub events. |
| config.github.webhookSecret | string | `""` | The secret used to verify GitHub webhook payloads. |
| config.groups | list | `[{"name":"us-east-2"},{"enabled":false,"name":"us-west-1"}]` | Groups are used to separate clients into logical groups, e.g. by region, datacenter, etc. Atleast one group must be defined. Group name must be unique and should not contain any hyphens (-). |
| config.listenAddr | string | `"0.0.0.0:8080"` | Listen address for the HTTP server. This is where the GitHub webhook should be configured to send events. |
| config.logLevel | string | `"info"` | Log level must be one of: debug, info, warn, error, fatal, panic, trace. |
| config.scheduler | string | `""` | Scheduler configuration options (optional). |
| extraObjects | list | `[]` | Array of extra manifests to deploy |
| fullnameOverride | string | `""` |  |
| hostAliases | list | `[]` | hostAliases to add to pod's hosts file. |
| image.pullPolicy | string | `"IfNotPresent"` | IfNotPresent, Always or Never. |
| image.repository | string | `"ghcr.io/hostinger/fireactions"` | Container image repository for fireactions-server. |
| image.tag | string | `"0.0.1"` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | Configure a list of imagePullsecrets. |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"ImplementationSpecific"` |  |
| ingress.tls | list | `[]` |  |
| kind | string | `"StatefulSet"` | Use either Deployment or StatefulSet (default). ref: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/  |
| livenessProbe | object | `{}` |  |
| metrics.enabled | bool | `false` | Enable metrics to be exposed. |
| metrics.port | int | `8081` | The port on which the application is exposing the metrics. |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| persistence.accessModes[0] | string | `"ReadWriteOnce"` |  |
| persistence.enabled | bool | `true` | Can be enabled when running as StatefulSet to create a VolumeClaimTemplate. |
| persistence.existingClaim | string | `""` | Use an existing PVC instead of the VolumeClaimTemplate. |
| persistence.size | string | `"8Gi"` |  |
| persistence.storageClass | string | `""` |  |
| podAnnotations | object | `{}` | Annotations to be added to the pod |
| podSecurityContext | object | `{}` | Security Context policies for pods |
| readinessProbe | object | `{}` |  |
| replicaCount | int | `1` | Number of replicas to run. At the moment HA setup is not supported, so this must be 1. |
| resources | object | `{}` |  |
| securityContext | object | `{}` | Security Context policies for containers |
| service.port | int | `8080` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account. |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created. |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. |
| tolerations | list | `[]` |  |
