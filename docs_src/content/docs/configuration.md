# Configuration

The harbor-sync binary

## Environment Variables

| ENV | DEFAULT | DESCRIPTION |
|---|---|---|
| `HARBOR_API_ENDPOINT` | - | specify the harbor URL |
| `HARBOR_USERNAME` | - | set the username used for authenticating with harbor |
| `HARBOR_PASSWORD` | - | password for harbor authentication |
| `LEADER_ELECT` | true | enable/disable leader election |
| `NAMESPACE` | kube-system | namespace in which harbor-sync runs (used for leader-election) |
| `HARBOR_POLL_INTERVAL` | 5m | poll interval to update harbor projects & robot accounts |
| `FORCE_SYNC_INTERVAL` | 10m | set this to force reconciliation after a certain time |
| `ROTATION_INTERVAL` | 60m | set this to rotate the credentials after the specified time |



## Command Line Interface

The harbor-sync binary has a subcommand that starts sync process: `controller`.

### Standalone mode

The controller may run in `standalone` mode: This removes the necessity to run inside the Kubernetes cluster. In this mode `harbor-sync` reads a config file (see `kind: HarborSync` and `kind: HarborRobotAccount`) and reconciles the robot accounts in Harbor. Webhooks will be called to propagate the credentials into other subsystems. The `mappings` field will have not effect - this is specific to Kubernetes. In standalone mode state (i.e. the credentials for the robot accounts) is stored on disk.

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
      --leader-elect                    enable leader election (default true)
      --metrics-addr string             The address the metric endpoint binds to. (default ":8080")
      --namespace string                namespace in which harbor-sync runs (used for leader-election) (default "kube-system")
      --rotation-interval duration      set this to rotate the credentials after the specified time (default 1h0m0s)
      --skip-tls-verification           Skip TLS certificate verification

Global Flags:
      --loglevel string   set the loglevel (default "info")

```
