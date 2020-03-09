# Metrics

Harbor Sync Controller exposes prometheus metrics. You configure the listen address / port via `-metrics-addr`.
The following metrics are available:

| metric name | type | labels | description |
|---|---|---|---|
| `http_request_duration_seconds` | histogram | `code,method,path` | keeps track of the duration API requests towards harbor |
| `harbor_matching_projects` | gauge | `config,selector_type,selector_project_name` | total number of matching projects per HarborSyncConfig |
| `harbor_robot_account_expiry` | gauge | `project,robot` | the date after which the robot account expires, expressed as Unix Epoch Time |
| `harbor_sync_sent_webhooks` | gauge | `config,target,status_code` | The number of webhooks sent |

## Alerts

Here are example alerts

```yaml
groups:
- name: harbor_rules
  rules:
  - alert: HarborAccountExpires
    expr: (harbor_robot_account_expiry - time()) / 86400 < 14
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "CRITICAL: harbor robot account '{{ $labels.robot }}' in project '{{ $labels.project }}' expires in less than 14d"
      description: "harbor robot account expires soon"
  - alert: HarborOutgoingWebhooksFailed
    expr: sum(increase(harbor_sync_sent_webhooks{status_code!="200"}[1h])) by (target) > 0
    labels:
      severity: critical
    annotations:
      summary: "CRITICAL: harbor outgoing webhook failed: '{{ $labels.target }}'"
      description: "harbor does not deliver the robot account information correctly"
```
