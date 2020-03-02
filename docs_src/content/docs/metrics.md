# Metrics

Harbor Sync Controller exposes prometheus metrics. You configure the listen address / port via `-metrics-addr`.
The following metrics are available:

| metric name | type | labels | description |
|---|---|---|---|
| `http_request_duration_seconds` | histogram | `code,method,path` | keeps track of the duration API requests towards harbor |
| `harbor_matching_projects` | gauge | `config,selector_type,selector_project_name` | total number of matching projects per HarborSyncConfig |
| `harbor_robot_account_expiry` | gauge | `project,robot` | the date after which the robot account expires, expressed as Unix Epoch Time |
| `harbor_sync_sent_webhooks` | gauge | `config,target,status_code` | The number of webhooks sent |
