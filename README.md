# Harbor Sync
You shouldn't care about renewing robot accounts or copying credentials from harbor to kubernetes.
This controller manages harbor robot accounts and synchronizes them with your cluster. They will be re-created if they expire and synced into the respective namespaces.
In addition, this controller can integrate other APIs using webhooks.


## Recon loop

This is pretty straight-forward:

* find harbor projects that match the configured regular expression
  * reconcile robot accounts: (re-)create them if they do not exist, are disabled, expired or we do not manage the token
* find namespaces using a `mapping` config
  * for each namespace: create a secret with type `dockerconfigjson` with the specified name.

The reconciliation loop is triggered from essentially three sources:
* Control Plane: whenever a SyncConfig is created/updated/deleted
* Harbor Polling: whenever the state in harbor changes (project or robota account is created, updated, deleted)
* time-based using the configured `force-sync-interval`: forces reconciliation in a fixed interval to cover cases like namespace creation or robot account expiration

## Configuration

| ENV | DEFAULT | DESCRIPTION |
|---|---|---|
| `HARBOR_API_ENDPOINT` | - | specify the harbor URL |
| `HARBOR_USERNAME` | - | set the username used for authenticating with harbor |
| `HARBOR_PASSWORD` | - | password for harbor authentication |


```
Usage:
  -harbor-poll-interval duration
        poll interval to update harbor projects & robot accounts (default 5m)
  -force-sync-interval
        force reconciliation interval (default 10m)
  -kubeconfig string
        Paths to a kubeconfig. Only required if out-of-cluster.
  -metrics-addr string
        The address the metric endpoint binds to. (default ":8080")
```

## Supported Use-cases

### 1:1 Mapping
Literally map harbor project to namespaces.

```yaml
spec:
  projectSelector:
  - type: Regex
    name: my-project
    robotAccountSuffix: k8s-sync-robot
    mapping:
    - type: Translate
      namespace: team-a
      secret: "my-project-pull-token"
    - type: Translate
      namespace: team-b
      secret: "my-project-pull-token"
```

This will create a robot account in `my-central-project` harbor project and sync the credentials into `team-a` and `team-b`'s namespace as secret `central-project-token`.

### 1:N using regular expressions

You can specify regular expressions to map a large number of projects to namespaces.

```yaml
spec:
  projectSelector:
  - type: Regex
    name: team-(.*)
    robotAccountSuffix: k8s-sync-robot
    mapping:
    - type: Translate
      namespace: team-$1    # references capturing group from projectSelector.name
      secret: team-$1-pull-token # same here
```

This maps harbor teams with the prefix `team-`. E.g. Harbor `team-frontend` maps to k8s namespace `team-frontend`. The secret's name will always be `my-pull-token`. Non-existent k8s namespaces will be ignored.

### 1:N Mapping with matching namespaces

You have one harbor project and want to deploy the pull secrets into several namespaces matching a regular expression:

```yaml
spec:
  projectSelector:
  - type: Regex
    name: platform-team
    robotAccountSuffix: k8s-sync-robot
    mapping:
    - type: Match  # treat namespace as regex
      namespace: team-.* # if ns matches this it will receive the secret
      secret: platform-pull-token
```

E.g. pull tokens for the `platform-team` project should be distributed into all namespaces matching `team-.*`.

## Webhooks

This is not yet *designed* or even implemented.
This controller should fire a webhook if a robot account expires and is being recreated. One could potentially integrate other secret storage systems such AWS Secrets Manager or HashiCorp Vault.

### Contributions

Pull requests are welcome!
* Read [CONTRIBUTING.md](./CONTRIBUTING.md) and check out [help wanted](https://github.com/moolen/harbor-sync/labels/help%20wanted) issues.
* Please submit github issues for feature requests, bugs or documentation problems
* Questions/comments and support can be posted as [github issue](https://github.com/moolen/harbor-sync/issues).
