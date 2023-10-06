# ![Fireactions Server](./chart-icon.png "Fireactions")

## Fireactions Server Helm Chart

## Introduction

This chart bootstraps a Fireactions server component as a Deployment or StatefulSet on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

### Prerequisites

- Kubernetes 1.17+ (recommended 1.20+)
- Helm 3.6+ (recommended 3.7+)

### Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| config.externalUrl | string | `"https://actions-on-firecrackers.example.com"` |  |
| config.github.appId | string | `""` |  |
| config.github.appPrivateKey | string | `""` |  |
| config.github.webhookSecret | string | `""` |  |
| config.listenAddr | string | `"0.0.0.0:8080"` |  |
| config.logLevel | string | `"debug"` |  |
| config.pools | object | `{}` |  |
| config.provisioners | object | `{}` |  |
| config.storage | object | `{}` |  |
| extraObjects | list | `[]` |  |
| fullnameOverride | string | `""` |  |
| hostAliases | list | `[]` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/konradasb/actions-on-firecrackers/server"` |  |
| image.tag | string | `"latest"` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"ImplementationSpecific"` |  |
| ingress.tls | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| postgresql.enabled | bool | `false` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| service.port | int | `80` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |
