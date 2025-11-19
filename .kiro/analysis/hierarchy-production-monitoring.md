# Hierarchy Implementation - Production Monitoring Guide

## Critical Metrics to Monitor

### 1. Role Inheritance Performance
```yaml
metrics:
  - name: role_inheritance_calculation_duration
    type: histogram
    unit: milliseconds
    alert_threshold: p99 > 500ms
    description: Time to calculate effective roles including inheritance

  - name: role_inheritance_cache_hit_rate
    type: gauge
    unit: percentage
    alert_threshold: < 80%
    description: Cache effectiveness for inherited roles

  - name: inherited_roles_per_user
    type: histogram
    alert_threshold: p95 > 50
    description: Number of inherited roles per user (complexity indicator)
```

### 2. Hierarchy Depth Monitoring
```yaml
metrics:
  - name: organization_hierarchy_depth
    type: gauge
    alert_threshold: > 8
    description: Current max depth of organization hierarchies

  - name: group_hierarchy_depth
    type: gauge
    alert_threshold: > 6
    description: Current max depth of group hierarchies

  - name: hierarchy_depth_limit_violations
    type: counter
    alert_threshold: > 5 per hour
    description: Attempts to exceed hierarchy depth limits
```

### 3. Cross-Organization Violations
```yaml
metrics:
  - name: cross_org_validation_failures
    type: counter
    alert_threshold: > 1 per day
    description: Attempts to create cross-organization relationships
    severity: HIGH
    action: Potential security breach attempt
```

### 4. Optimistic Locking
```yaml
metrics:
  - name: optimistic_lock_conflicts
    type: counter
    alert_threshold: > 100 per hour
    description: Concurrent update conflicts detected

  - name: optimistic_lock_retry_success_rate
    type: gauge
    alert_threshold: < 95%
    description: Success rate of retried updates after conflicts
```

### 5. Database Query Performance
```yaml
metrics:
  - name: hierarchy_cte_query_duration
    type: histogram
    unit: milliseconds
    alert_threshold: p99 > 100ms
    description: CTE query performance for hierarchy traversal

  - name: get_hierarchy_depth_duration
    type: histogram
    alert_threshold: p95 > 50ms
    description: Time to calculate hierarchy depth
```

## Alert Configurations

### Critical Alerts (Page immediately)
```yaml
- alert: CircularReferenceDetected
  condition: circular_reference_attempts > 0
  severity: CRITICAL
  message: "Circular reference in hierarchy detected"
  runbook: Check for data corruption, investigate source

- alert: CrossOrgSecurityBreach
  condition: cross_org_validation_failures > 1
  severity: CRITICAL
  message: "Multiple cross-org attempts from same user"
  runbook: Potential privilege escalation attempt

- alert: HierarchyDepthAnomaly
  condition: max_hierarchy_depth > configured_limit
  severity: CRITICAL
  message: "Hierarchy depth exceeds configured limits"
  runbook: Database constraint may be bypassed
```

### Warning Alerts (Notify on-call)
```yaml
- alert: HighOptimisticLockConflicts
  condition: optimistic_lock_conflicts_rate > 50/min
  severity: WARNING
  message: "High rate of concurrent update conflicts"
  runbook: May indicate bot activity or system issue

- alert: RoleInheritanceSlow
  condition: role_calculation_p99 > 1s
  severity: WARNING
  message: "Role inheritance calculation degraded"
  runbook: Check cache health, consider increasing TTL

- alert: CacheMissRateHigh
  condition: cache_hit_rate < 70%
  severity: WARNING
  message: "Role inheritance cache ineffective"
  runbook: Review cache eviction, possible memory pressure
```

## Dashboard Panels

### Executive Dashboard
1. **Hierarchy Health Score** (0-100)
   - Weighted composite of depth, performance, and error rates

2. **Active Organizations/Groups**
   - Total count with hierarchy depth distribution

3. **Role Inheritance Effectiveness**
   - Average inherited roles per user
   - Cache hit rate
   - Calculation latency p50/p95/p99

### Operations Dashboard
1. **Real-time Performance**
   - CTE query latency histogram
   - Cache operation latency
   - Optimistic lock conflict rate

2. **Hierarchy Visualization**
   - Max depth by organization
   - Deepest paths identified
   - Orphaned nodes count

3. **Error Tracking**
   - Validation failures by type
   - Cross-org attempts
   - Circular reference detections

### Security Dashboard
1. **Access Pattern Analysis**
   - Cross-org access attempts
   - Unusual inheritance patterns
   - Rapid hierarchy modifications

2. **Audit Trail**
   - Hierarchy structure changes
   - Role assignment changes
   - Failed validation attempts

## Logging Requirements

### Structured Logging Fields
```json
{
  "event": "hierarchy_operation",
  "operation": "calculate_depth|traverse_parents|get_descendants",
  "entity_type": "organization|group",
  "entity_id": "uuid",
  "depth": 5,
  "duration_ms": 45,
  "cache_hit": false,
  "error": null,
  "user_id": "uuid",
  "correlation_id": "request-uuid"
}
```

### Audit Events to Track
```go
const (
    AuditEventHierarchyDepthExceeded    = "hierarchy.depth.exceeded"
    AuditEventCircularReferenceBlocked  = "hierarchy.circular.blocked"
    AuditEventCrossOrgBlocked          = "hierarchy.cross_org.blocked"
    AuditEventOptimisticLockConflict   = "update.optimistic_lock.conflict"
    AuditEventRoleInheritanceCalculated = "role.inheritance.calculated"
)
```

## Performance Baselines

### Expected Performance Characteristics
| Operation | P50 | P95 | P99 | Max |
|-----------|-----|-----|-----|-----|
| GetParentHierarchy | 10ms | 25ms | 50ms | 100ms |
| GetDescendants | 15ms | 35ms | 75ms | 150ms |
| CalculateEffectiveRoles | 25ms | 100ms | 250ms | 500ms |
| GetHierarchyDepth | 5ms | 15ms | 30ms | 50ms |
| UpdateWithVersion | 20ms | 50ms | 100ms | 200ms |

### Capacity Limits
- Max organizations: 10,000
- Max groups per org: 1,000
- Max hierarchy depth: 10 (orgs) / 8 (groups)
- Max roles per user (including inherited): 100
- Cache size: 10,000 entries per node
- Cache TTL: 5 minutes (roles), 30 minutes (profiles)

## Anomaly Detection Rules

### Behavioral Anomalies
1. **Sudden Depth Increase**
   - Alert if any hierarchy grows by >3 levels in 1 hour

2. **Mass Hierarchy Changes**
   - Alert if >10% of hierarchies modified in 10 minutes

3. **Role Explosion**
   - Alert if any user suddenly has >20 new inherited roles

4. **Cache Thrashing**
   - Alert if same key evicted/loaded >10 times in 1 minute

## Health Check Endpoints

### Hierarchy Health Check
```http
GET /health/hierarchy
Response:
{
  "status": "healthy|degraded|unhealthy",
  "max_org_depth": 7,
  "max_group_depth": 6,
  "circular_references": 0,
  "orphaned_nodes": 0,
  "cache_hit_rate": 0.92,
  "avg_calculation_time_ms": 35
}
```

### Role Inheritance Health Check
```http
GET /health/role-inheritance
Response:
{
  "status": "healthy",
  "cache_entries": 4523,
  "cache_hit_rate": 0.89,
  "avg_roles_per_user": 4.2,
  "avg_inherited_roles": 2.8,
  "calculation_errors_1h": 0
}
```

## Runbook References

### Common Issues and Resolutions

1. **High Cache Miss Rate**
   - Check Redis memory usage
   - Review cache eviction policy
   - Consider increasing TTL during low-change periods

2. **Slow Hierarchy Queries**
   - Verify indexes are being used (EXPLAIN ANALYZE)
   - Check for missing indexes on parent_id
   - Consider partitioning large tables

3. **Optimistic Lock Storms**
   - Implement exponential backoff in clients
   - Check for automation/bots making rapid updates
   - Consider request rate limiting

4. **Memory Pressure from Deep Hierarchies**
   - Review and enforce depth limits
   - Implement query result size limits
   - Consider pagination for large result sets

## Capacity Planning

### Growth Indicators to Monitor
- Organization creation rate
- Group creation rate
- Average hierarchy depth trend
- Role assignment rate
- Cache size growth rate

### Scaling Triggers
- Cache hit rate < 75% consistently
- P99 latency > SLA for 15 minutes
- Memory usage > 80%
- Database connections > 80% of pool

## Compliance and Audit

### Required Audit Retention
- Hierarchy structure changes: 7 years
- Role assignments: 7 years
- Access failures: 90 days
- Performance metrics: 30 days

### Compliance Checks
- Weekly: Verify no circular references
- Monthly: Validate depth limits enforced
- Quarterly: Audit cross-org access attempts
- Annually: Full hierarchy integrity check
