# Getting Started

## What is Harbor Sync Controller?
Harbor Sync Controller synchronizes Harbor with your Kubernetes cluster. It simplifies the management of robot accounts by automating the process of renewal and distribution of access tokens.

![Harbor Sync Controller]({{< baseurl >}}/harbor-sync-overview.png)

# Installation

## Prerequisites
### Harbor
You need a [Harbor](https://goharbor.io/#getting-started) deployment and a user with elevated privileges to read all projects and robot accounts: Create a dedicated user with `ProjectAdmin` permissions. Refer to the [official docs](https://github.com/goharbor/harbor/blob/master/docs/user_guide.md) about how to set up user authentication and role binding.

Also, check your harbor version. It must be `>= 1.8.0`. That's the version that [introduced robot accounts](https://github.com/goharbor/harbor/releases/tag/v1.8.0). Prior to that version it was not possible to create robot accounts.


### Execution Environment
The Controller usually runs in [Kubernetes](https://kubernetes.io) so you need a Kubernetes Cluster, too. Also, you can run this controller in a [standalone mode]({{< ref "configuration.md" >}}).

## Deployment

The following command creates a deployment and the necessary RBAC roles for the controller.

```bash
$ wget https://raw.githubusercontent.com/moolen/harbor-sync/master/install/kubernetes/quick-install.yaml

# change environment variables
$ vim quick-install.yaml

$ kubectl create -f quick-install.yaml
```

Also, take a look at the [kustomize setup](https://github.com/moolen/harbor-sync/tree/master/config) if you use kustomize. If you prefer to use helm for deployment feel free to contribute a helm chart.

## Important Notes

Harbor Sync Controller is **stateless**. Right now, harbor-sync stores the credentials for the robot accounts in a crd. This is necessary because there is no way to retrieve the token from the harbor API.

You can inspect the robot account credentials using: `kubectl get harborrobotaccount`.

## Next steps

You may want to check out the the [Usage Examples]({{< ref "usage.md" >}}) or [Configuration]({{< ref "configuration.md" >}}).
