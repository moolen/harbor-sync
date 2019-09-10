# Configuration

## Environment Variables

| ENV | DEFAULT | DESCRIPTION |
|---|---|---|
| `HARBOR_API_ENDPOINT` | - | specify the harbor URL |
| `HARBOR_USERNAME` | - | set the username used for authenticating with harbor |
| `HARBOR_PASSWORD` | - | password for harbor authentication |


## Commandline flags

```
Usage:
  -harbor-poll-interval duration
        poll interval to update harbor projects & robot accounts (default 5m)
  -force-sync-interval
        force reconciliation interval (default 10m)
  -rotation-interval duration
        set this to rotate the credentials after the specified time (default 1h0m0s)
  -store string
        path to the credentials cache (default "/data")
  -v value
        log level for V logs
  -kubeconfig string
        Paths to a kubeconfig. Only required if out-of-cluster.
  -metrics-addr string
        The address the metric endpoint binds to. (default ":8080")
```
