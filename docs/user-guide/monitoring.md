# Monitoring

Fireactions provides Prometheus metrics for monitoring.

The metrics can be enabled by setting the `metrics.enabled` configuration option to `true`. The metrics are exposed on the `/metrics` endpoint.

## Metrics

The following metrics are available, excluding the default Prometheus metrics:

| Metric Name                          | Description                                       | Labels                   |
|--------------------------------------|---------------------------------------------------|--------------------------|
| `fireactions_pool_current_runners_count` | Current number of runners in a pool           | `pool` (the pool name)   |
| `fireactions_pool_max_runners_count`     | Maximum number of runners in a pool           | `pool` (the pool name)   |
| `fireactions_pool_min_runners_count`     | Minimum number of runners in a pool           | `pool` (the pool name)   |
| `fireactions_pool_scale_requests`        | Number of scale requests for a pool           | `pool` (the pool name)   |
| `fireactions_pool_scale_failures`        | Number of scale failures for a pool           | `pool` (the pool name)   |
| `fireactions_pool_scale_successes`       | Number of scale successes for a pool          | `pool` (the pool name)   |
| `fireactions_pool_status`                | Status of a pool. 0 is paused, 1 is active    | `pool` (the pool name)   |
| `fireactions_pool_total`                 | Total number of pools                         | No labels                |
| `fireactions_server_up`                  | Whether the server is up. 0 is down, 1 is up  | No labels                |

## Grafana Dashboard

Example Grafana dashboard for vizualisation of Fireactions metrics:

![Grafana Dashboard](../images/grafana-dashboard.png)
