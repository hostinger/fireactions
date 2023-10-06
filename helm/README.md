# Fireactions

<img src="./fireactions-server/chart-icon.png" alt="logo" width="150"/>

## Fireactions Helm Charts

Official [Fireactions](https://fireactions.github.io/) Helm Charts for deploying it on [Kubernetes](https://kubernetes.io/).

## Before you begin

### Setup a Kubernetes Cluster

### Install Helm

Grab the latest [Helm release](https://github.com/helm/helm#install).

### Add Helm chart repo

Add the repo as follows:

```console
helm repo add fireactions https://hostinger.github.io/fireactions/
helm repo update
```

Fireactions Helm charts can be also found on [ArtifactHub](https://artifacthub.io/packages/search?repo=fireactions).

## Search and install charts

```console
helm search repo fireactions/
helm install my-release fireactions/<chart>
```

**_NOTE_**: For instructions on how to install a chart follow instructions in its `README.md`.

## Contributing
