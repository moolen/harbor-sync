# Configuration

The harbor-sync binary

## Environment Variables

| ENV | DEFAULT | DESCRIPTION |
|---|---|---|
| `HARBOR_API_ENDPOINT` | - | specify the harbor URL |
| `HARBOR_USERNAME` | - | set the username used for authenticating with harbor |
| `HARBOR_PASSWORD` | - | password for harbor authentication |


## Command Line Interface

The harbor-sync binary has two subcommands: `store` and `controller`. Store is for dumping the credentials on disk. Controller runs the controller.

### Standalone mode

The controller may run in `standalone` mode: This removes the necessity to run inside the Kubernetes cluster. In this mode `harbor-sync` reads a config file (see `kind: HarborSync`) and reconciles the robot accounts in Harbor. Webhooks will be called to propagate the credentials into other subsystems. The `mappings` field will have not effect - this is specific to Kubernetes.

```
Controller should run inside Kubernetes. It reconciles the desired state by managing the robot accounts in Harbor.

Usage:
  harbor-sync controller [flags]
  harbor-sync controller [command]

Available Commands:
  standalone  Runs the controller in standalone mode. Does not require Kubernetes. It manages robot accounts and sends webhooks.

Flags:
      --force-sync-interval duration    set this to force reconciliation after a certain time (default 10m0s)
      --harbor-api-endpoint string      URL to the Harbor API Endpoint
      --harbor-password string          Harbor password to use for authentication
      --harbor-poll-interval duration   poll interval to update harbor projects & robot accounts (default 5m0s)
      --harbor-username string          Harbor username to use for authentication
  -h, --help                            help for controller
      --metrics-addr string             The address the metric endpoint binds to. (default ":8080")
      --rotation-interval duration      set this to rotate the credentials after the specified time (default 1h0m0s)

Global Flags:
      --loglevel string   set the loglevel (default "debug")
      --store string      path in which the credentials will be stored (default "/data")

```
