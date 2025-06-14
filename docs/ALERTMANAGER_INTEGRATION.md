# AlertManager Integration

This document describes how to integrate Prometheus AlertManager with SLAR to automatically create alerts from AlertManager webhooks.

## Overview

The AlertManager integration allows SLAR to receive webhook notifications from Prometheus AlertManager and automatically create alerts in the SLAR system. This enables seamless integration between your monitoring infrastructure and on-call management.

## Features

- ✅ Receives AlertManager webhook notifications
- ✅ Converts AlertManager alerts to SLAR alerts
- ✅ Handles both firing and resolved alerts
- ✅ Extracts severity from alert labels
- ✅ Creates meaningful descriptions from annotations
- ✅ Prevents duplicate alerts using fingerprints
- ✅ Auto-assigns alerts to current on-call user

## API Endpoints

### Webhook Endpoint
```
POST /alertmanager/webhook
```

This endpoint receives webhook notifications from AlertManager.

### Info Endpoint
```
GET /alertmanager/info
```

Returns configuration information and examples.

## AlertManager Configuration

Add the following configuration to your `alertmanager.yml`:

```yaml
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'slar-webhook'

receivers:
- name: 'slar-webhook'
  webhook_configs:
  - url: 'http://your-slar-api:8080/alertmanager/webhook'
    send_resolved: true
    http_config:
      basic_auth:
        username: 'your-username'  # Optional
        password: 'your-password'  # Optional
```

## Alert Mapping

AlertManager alerts are mapped to SLAR alerts as follows:

| AlertManager Field | SLAR Field | Notes |
|-------------------|------------|-------|
| `fingerprint` | `id` | Used as unique identifier |
| `labels.alertname` | `title` | Alert title |
| `annotations.summary` or `annotations.description` | `description` | Alert description |
| `labels.severity` | `severity` | Maps to critical/warning/info |
| `startsAt` | `created_at` | When alert started |
| `status` | `status` | firing → new, resolved → closed |

## Severity Mapping

| AlertManager Severity | SLAR Severity |
|----------------------|---------------|
| `critical` | `critical` |
| `warning` | `warning` |
| `info` | `info` |
| Other values | `warning` (default) |

## Alert Lifecycle

### Firing Alerts
1. AlertManager sends webhook with `status: "firing"`
2. SLAR creates new alert with status `"new"`
3. Alert is auto-assigned to current on-call user
4. If alert already exists and was closed, it's reopened

### Resolved Alerts
1. AlertManager sends webhook with `status: "resolved"`
2. SLAR updates existing alert status to `"closed"`
3. If alert doesn't exist, creates it as closed

## Example Webhook Payload

```json
{
  "version": "4",
  "groupKey": "123456789",
  "truncatedAlerts": 0,
  "status": "firing",
  "receiver": "slar-webhook",
  "groupLabels": {
    "alertname": "HighCPUUsage"
  },
  "commonLabels": {
    "alertname": "HighCPUUsage",
    "instance": "node1",
    "severity": "critical"
  },
  "commonAnnotations": {
    "summary": "High CPU usage detected on node1"
  },
  "externalURL": "http://alertmanager.example.com",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HighCPUUsage",
        "instance": "node1",
        "severity": "critical"
      },
      "annotations": {
        "summary": "CPU usage is above 90% on node1"
      },
      "startsAt": "2025-06-14T16:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus.example.com/graph?g0.expr=...&g0.tab=1",
      "fingerprint": "abcdef1234567890"
    }
  ]
}
```

## Testing

You can test the webhook endpoint using curl:

```bash
curl -X POST http://localhost:8080/alertmanager/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "version": "4",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "TestAlert",
          "severity": "warning"
        },
        "annotations": {
          "summary": "This is a test alert"
        },
        "startsAt": "2025-06-14T16:00:00Z",
        "fingerprint": "test123"
      }
    ]
  }'
```

## Troubleshooting

### Common Issues

1. **Webhook not received**
   - Check AlertManager configuration
   - Verify SLAR API is accessible from AlertManager
   - Check network connectivity and firewall rules

2. **Alerts not created**
   - Check SLAR API logs for errors
   - Verify webhook payload format
   - Ensure database connectivity

3. **Duplicate alerts**
   - Check if fingerprint is provided in webhook
   - Verify alert ID generation logic

### Logs

Check SLAR API logs for webhook processing:

```bash
# If using Docker
docker logs slar-api

# If running directly
tail -f /var/log/slar/api.log
```

## Security Considerations

1. **Authentication**: Consider adding authentication to the webhook endpoint
2. **Network Security**: Use HTTPS and restrict access to the webhook endpoint
3. **Rate Limiting**: Implement rate limiting to prevent abuse
4. **Validation**: Webhook payload is validated before processing

## Future Enhancements

- [ ] Support for custom alert routing rules
- [ ] Integration with multiple AlertManager instances
- [ ] Alert enrichment from external sources
- [ ] Custom severity mapping configuration
- [ ] Webhook authentication and authorization 