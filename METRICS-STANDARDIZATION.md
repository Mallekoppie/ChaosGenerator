# Target Metrics Standardization

## Overview

As of this update, target service metrics have been standardized to use a common naming convention with `target_type` labels instead of protocol-specific metric names. This allows for:

- **Unified dashboards** that work across all target types (HTTP, gRPC, and future protocols)
- **Simplified metric queries** with consistent naming patterns
- **Better extensibility** for adding new target types without dashboard changes
- **Consistent monitoring** across different protocols

## Changes Summary

### Metric Names

All target metrics have been renamed to remove protocol identifiers from the metric name and add them as labels instead. Agent metrics have also been updated to include `target_protocol` label to track which protocol the agent is testing against.

#### Before (Protocol-Specific)
```
chaos_target_http_request_duration_seconds
chaos_target_http_requests_total
chaos_target_http_processing_time_seconds
chaos_target_http_connection_time_seconds
chaos_target_http_errors_total
chaos_target_http_success_total

chaos_target_grpc_request_duration_seconds
chaos_target_grpc_requests_total
chaos_target_grpc_processing_time_seconds
chaos_target_grpc_errors_total
chaos_target_grpc_success_total
chaos_target_grpc_response_size_bytes
chaos_target_grpc_request_size_bytes
```

#### After (Standardized)
```
chaos_target_request_duration_seconds{target_type="http|grpc"}
chaos_target_requests_total{target_type="http|grpc"}
chaos_target_processing_time_seconds{target_type="http|grpc"}
chaos_target_connection_time_seconds{target_type="http"}  # HTTP only
chaos_target_errors_total{target_type="http|grpc"}
chaos_target_success_total{target_type="http|grpc"}
chaos_target_response_size_bytes{target_type="http|grpc"}
chaos_target_request_size_bytes{target_type="http|grpc"}
```

### Label Structure

All target metrics now include `target_type` as the first label:

| Metric | Labels |
|--------|--------|
| `chaos_target_request_duration_seconds` | `target_type`, `use_case`, `method`, `status_code` |
| `chaos_target_requests_total` | `target_type`, `use_case`, `method`, `status_code` |
| `chaos_target_processing_time_seconds` | `target_type`, `use_case` |
| `chaos_target_connection_time_seconds` | `target_type`, `use_case` |
| `chaos_target_errors_total` | `target_type`, `use_case`, `error_type` |
| `chaos_target_success_total` | `target_type`, `use_case` |
| `chaos_target_response_size_bytes` | `target_type`, `use_case` |
| `chaos_target_request_size_bytes` | `target_type`, `use_case` |

All agent metrics now include `target_protocol` as the first label:

| Metric | Labels |
|--------|--------|
| `chaos_agent_request_duration_seconds` | `target_protocol`, `use_case`, `method`, `status_code`, `connection_pooled` |
| `chaos_agent_requests_total` | `target_protocol`, `use_case`, `method`, `status_code`, `connection_pooled` |
| `chaos_agent_errors_total` | `target_protocol`, `use_case`, `error_type`, `connection_pooled` |
| `chaos_agent_connection_time_seconds` | `target_protocol`, `use_case`, `connection_pooled` |
| `chaos_agent_response_time_seconds` | `target_protocol`, `use_case`, `connection_pooled` |
| `chaos_agent_processing_time_seconds` | `target_protocol`, `use_case`, `connection_pooled` |
| `chaos_agent_success_total` | `target_protocol`, `use_case`, `connection_pooled` |

### Query Examples

#### Query Both HTTP and gRPC Together
```promql
# Combined request rate across all target types
sum(rate(chaos_target_requests_total[5m])) by (target_type, use_case)

# p95 latency comparison
histogram_quantile(0.95, sum by (le, target_type, use_case) 
  (rate(chaos_target_request_duration_seconds_bucket[5m])))

# Agent perspective - all protocols
sum(rate(chaos_agent_requests_total[5m])) by (target_protocol, use_case)

# p95 end-to-end latency from agent by protocol
histogram_quantile(0.95, sum by (le, target_protocol, use_case) 
  (rate(chaos_agent_request_duration_seconds_bucket[5m])))
```

#### Query Specific Target Type
```promql
# HTTP-only metrics
rate(chaos_target_requests_total{target_type="http"}[5m])

# gRPC-only metrics
rate(chaos_target_requests_total{target_type="grpc"}[5m])

# Agent testing HTTP targets
rate(chaos_agent_requests_total{target_protocol="https"}[5m])

# Agent testing gRPC targets
rate(chaos_agent_requests_total{target_protocol="grpc"}[5m])
```

#### Cross-Protocol Comparison
```promql
# Success rate by target type
sum(rate(chaos_target_success_total[5m])) by (target_type) 
  / 
sum(rate(chaos_target_requests_total[5m])) by (target_type)

# Agent success rate by protocol
sum(rate(chaos_agent_success_total[5m])) by (target_protocol) 
  / 
sum(rate(chaos_agent_requests_total[5m])) by (target_protocol)
```

## Implementation Details

### Code Changes

1. **internal/target-http/metrics.go**
   - Renamed all metrics from `chaos_target_http_*` to `chaos_target_*`
   - Added `target_type` as first label to all metric definitions
   - Added `ResponseSize` and `RequestSize` metrics for consistency with gRPC

2. **internal/target-grpc/metrics.go**
   - Renamed all metrics from `chaos_target_grpc_*` to `chaos_target_*`
   - Added `target_type` as first label to all metric definitions

3. **internal/target-http/handlers.go**
   - Updated all `WithLabelValues()` calls to include `"http"` as first parameter
   - Updated all 11 use case handlers (UC1-UC11)

4. **internal/target-grpc/service.go**
   - Updated all `WithLabelValues()` calls to include `"grpc"` as first parameter
   - Updated all 11 use case implementations (UC1-UC11)

5. **internal/agent/metrics.go**
   - Added `target_protocol` as first label to all agent metric definitions
   - Updated metric help text to be protocol-agnostic

6. **internal/agent/executor.go**
   - Added `targetProtocol` field to `HTTPExecutor` struct
   - Updated all `WithLabelValues()` calls to include `targetProtocol` as first parameter
   - Updated all metric recording calls in all use case handlers

7. **internal/agent/grpc_executor.go**
   - Added `targetProtocol` field to `GRPCExecutor` struct
   - Updated all `WithLabelValues()` calls to include `targetProtocol` as first parameter
   - Updated all metric recording calls in all use case handlers

8. **monitoring/grafana/provisioning/dashboards/chaos-comparison.json**
   - Updated all PromQL queries to use new metric names
   - Added `target_type="http"` filter where needed, or removed filter to show all protocols
   - Added `target_protocol` to agent metric queries
   - Updated legend formats to include `target_type` and `target_protocol` labels

## Benefits

### 1. Unified Dashboards
Dashboards can now display metrics from both HTTP and gRPC targets side-by-side or combined:
```promql
# Single query for all protocols
sum(rate(chaos_target_requests_total[5m])) by (target_type)
```

### 2. Easier Target Addition
Adding new target types (WebSocket, MQTT, etc.) only requires:
- Create new target service with standardized metrics
- Add appropriate `target_type` label value
- Existing dashboards automatically include new target type

### 3. Consistent Monitoring
All target types report the same core metrics:
- Request duration
- Request count
- Processing time
- Errors and successes
- Request/response sizes

### 4. Better Cardinality Control
Protocol type moved from metric name to label reduces the number of unique metric names while maintaining queryability.

## Migration Notes

### For Dashboard Users
- Existing saved queries using old metric names will need to be updated
- Update any custom alerts or recording rules to use new metric names
- The provided Grafana dashboards have been updated automatically

### For Developers
When implementing new target types:
1. Use the standardized metric names from `chaos_target_*` pattern
2. Always include `target_type` as the first label
3. Use the value corresponding to your protocol (e.g., "websocket", "mqtt")
4. Follow the label structure defined in this document

### Example New Target Implementation
```go
// In internal/target-websocket/metrics.go
var (
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "chaos_target_request_duration_seconds",
            Help: "Target service request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"target_type", "use_case", "method", "status_code"},
    )
    // ... other metrics
)

// In handlers
RequestDuration.WithLabelValues("websocket", useCase, method, statusCode).Observe(duration)
```

## Backward Compatibility

⚠️ **Breaking Change**: Old metric names (`chaos_target_http_*`, `chaos_target_grpc_*`) are no longer emitted.

If you have:
- Custom dashboards
- Alerting rules
- Recording rules
- External metric consumers

You must update them to use the new standardized metric names with `target_type` labels.

## Testing

To verify metrics are working correctly:

1. **Start the services:**
   ```bash
   make run-full-stack
   ```

2. **Check Prometheus targets:**
   - Open http://localhost:9091/targets
   - Verify both target-http and target-grpc are UP

3. **Query metrics:**
   ```bash
   # Check HTTP metrics
   curl http://localhost:9092/metrics | grep chaos_target_request_duration
   
   # Check gRPC metrics  
   curl http://localhost:9095/metrics | grep chaos_target_request_duration
   ```

4. **Verify labels:**
   Both should show `chaos_target_request_duration_seconds` with `target_type="http"` or `target_type="grpc"` respectively.

## Future Enhancements

With this standardization in place, future improvements can include:

1. **Unified Target Dashboard**: Single dashboard showing all target types
2. **Protocol Comparison Views**: Automatically compare performance across protocols
3. **Dynamic Target Discovery**: Dashboards automatically adapt to new target types
4. **Cross-Protocol SLOs**: Define service level objectives across all protocols

## Questions?

For questions or issues related to this change, refer to:
- [CONFIGURATION.md](CONFIGURATION.md) for metric configuration
- [readme.md](readme.md) for general documentation
- Prometheus query examples above for common queries
