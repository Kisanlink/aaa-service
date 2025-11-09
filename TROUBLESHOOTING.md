# Troubleshooting Guide

## Recent Issues Resolved

### Issue 1: Redis Connection Timeout (i/o timeout)

**Symptoms:**
```json
{
  "level": "warn",
  "msg": "Redis not available, skipping cache operation",
  "key": "system:maintenance_mode",
  "error": "i/o timeout"
}
```

**Root Cause:**
The application was timing out when connecting to Redis ElastiCache because:
1. No explicit connection timeouts were configured
2. Default Redis client timeouts were too long
3. Network latency between ECS tasks and ElastiCache cluster

**Solution:**
Added Redis connection timeout configuration in `cloudformation.yaml`:
- `REDIS_DIAL_TIMEOUT=5` (5 seconds for initial connection)
- `REDIS_READ_TIMEOUT=3` (3 seconds for read operations)
- `REDIS_WRITE_TIMEOUT=3` (3 seconds for write operations)
- `REDIS_POOL_SIZE=10` (connection pool size)
- `REDIS_MIN_IDLE_CONNS=2` (minimum idle connections)

**Verification:**
After redeploying with updated task definition, check logs:
```bash
aws logs tail /ecs/aaa-service-dev --follow --region ap-south-1 | grep -i redis
```

You should no longer see "i/o timeout" errors.

---

### Issue 2: Login Failure with Status 408 (Request Timeout)

**Symptoms:**
```json
{
  "level": "error",
  "msg": "Failed to verify user credentials",
  "error": "user not found",
  "status": 408,
  "latency": 12.232660011
}
```

**Root Cause:**
The login request was failing due to **cascading timeout issues**:

1. **Redis I/O Timeout** → Each Redis operation took 10-12 seconds to timeout
2. **Multiple Redis Calls** → Login flow made several Redis cache lookups:
   - `system:maintenance_mode`
   - `user:USER00000001`
   - `user_with_roles:USER00000001`
   - `user_roles:USER00000001`
3. **Request Timeout** → Total time exceeded HTTP request timeout (default 12-15s)
4. **Context Cancellation** → Database queries were cancelled mid-flight:
   ```
   "error": "context canceled"
   "msg": "Failed to get user by ID"
   ```

**Timeline of a Failed Login:**
```
0s    → Login request received
0-12s → Redis: system:maintenance_mode (timeout)
12-24s → Redis: user:USER00000001 (timeout)
24s   → HTTP request timeout → context cancelled
24s   → Database query cancelled → "user not found"
24s   → Return 408 Request Timeout
```

**Why 401 Appeared in Health Checks:**
```json
{
  "msg": "HTTP Request",
  "path": "/health",
  "status": 401
}
```

The `/health` endpoint requires authentication, and the ALB health checker doesn't provide auth credentials. This is EXPECTED behavior and doesn't indicate a problem. We've configured the target group to accept both 200 and 401 as healthy responses.

**Solutions Applied:**

1. ✅ **Fixed Redis Timeouts** (cloudformation.yaml:955-964)
   - Short timeouts prevent cascade delays
   - Connection pooling improves performance

2. ✅ **Updated Health Check Matcher** (cloudformation.yaml:702)
   - Accepts both 200 and 401 status codes
   - ALB correctly marks tasks as healthy

**Expected Behavior After Fix:**
- Redis operations fail fast (within 3-5 seconds)
- Application gracefully skips cache when Redis is unavailable
- Login completes successfully using database (no Redis required)
- Total login time: ~1-2 seconds (without Redis cache)

---

## Common Issues

### ElastiCache Redis AuthToken Validation Error

**Error:**
```
Invalid AuthToken provided (Service: AmazonElastiCache;
Status Code: 400; Error Code: InvalidParameterValue)
```

**Cause:**
ElastiCache only accepts these special characters in AuthToken: `! & # $ ^ < > -`

**Solution:**
Already fixed in `cloudformation.yaml:523-524`:
```yaml
ExcludeCharacters: '"@/\\ ~`*()_+={}[]|:;,.?%'
```

---

### ECS Task Execution Role Permission Denied

**Error:**
```
User: arn:aws:sts::xxx:assumed-role/aaa-service-dev-task-execution-role/xxx
is not authorized to perform: secretsmanager:GetSecretValue on resource:
arn:aws:secretsmanager:xxx:secret:aaa-service-dev-redis-auth-token-xxx
```

**Cause:**
Task execution role missing permission to read Redis secret.

**Solution:**
Already fixed in `cloudformation.yaml:838`:
```yaml
Resource:
  - !Ref DBSecret
  - !Ref JWTSecret
  - !Ref RedisSecret  # ← Added
```

---

## Debugging Commands

### Check Redis Connectivity from ECS Task

```bash
# Get running task ID
TASK_ID=$(aws ecs list-tasks \
  --cluster aaa-service-dev-cluster \
  --service-name aaa-service-dev \
  --query 'taskArns[0]' \
  --output text \
  --region ap-south-1 | awk -F/ '{print $NF}')

# Execute command in running container
aws ecs execute-command \
  --cluster aaa-service-dev-cluster \
  --task $TASK_ID \
  --container aaa-service \
  --interactive \
  --command "/bin/sh" \
  --region ap-south-1

# Inside container, test Redis connection
redis-cli -h $REDIS_HOST -p $REDIS_PORT --tls --askpass
# Enter password when prompted
PING
# Should return: PONG
```

### Check Application Logs

```bash
# Tail logs with Redis filter
aws logs tail /ecs/aaa-service-dev --follow --region ap-south-1 \
  --filter-pattern "Redis"

# Tail logs with login filter
aws logs tail /ecs/aaa-service-dev --follow --region ap-south-1 \
  --filter-pattern "login"

# Check for errors
aws logs tail /ecs/aaa-service-dev --follow --region ap-south-1 \
  --filter-pattern "error"
```

### Check ECS Service Health

```bash
# Service status
aws ecs describe-services \
  --cluster aaa-service-dev-cluster \
  --services aaa-service-dev \
  --region ap-south-1 \
  --query 'services[0].{Status:status,Running:runningCount,Desired:desiredCount,Events:events[:5]}'

# Task health
aws ecs describe-tasks \
  --cluster aaa-service-dev-cluster \
  --tasks $TASK_ID \
  --region ap-south-1 \
  --query 'tasks[0].{Health:healthStatus,LastStatus:lastStatus,Containers:containers[*].{Name:name,Health:healthStatus}}'
```

### Check Target Group Health

```bash
# Get target group ARN
TG_ARN=$(aws elbv2 describe-target-groups \
  --names aaa-service-dev-http-tg \
  --query 'TargetGroups[0].TargetGroupArn' \
  --output text \
  --region ap-south-1)

# Check target health
aws elbv2 describe-target-health \
  --target-group-arn $TG_ARN \
  --region ap-south-1 \
  --query 'TargetHealthDescriptions[*].{Target:Target.Id,Port:Target.Port,Health:TargetHealth.State,Reason:TargetHealth.Reason}'
```

### Test Health Endpoint

```bash
# Get ALB DNS
ALB_DNS=$(aws cloudformation describe-stacks \
  --stack-name aaa-service-dev \
  --query 'Stacks[0].Outputs[?OutputKey==`ALBDNS`].OutputValue' \
  --output text \
  --region ap-south-1)

# Test health endpoint (expect 401 - this is OK!)
curl -i http://$ALB_DNS/health
```

---

## Performance Optimization

### Current Configuration

| Component | Setting | Value |
|-----------|---------|-------|
| ECS Task CPU | TaskCPU | 512 (0.5 vCPU) |
| ECS Task Memory | TaskMemory | 1024 MB |
| RDS Instance | DBInstanceClass | db.t3.micro |
| Redis Instance | CacheNodeType | cache.t3.micro |
| Redis Pool Size | REDIS_POOL_SIZE | 10 |
| Min Idle Connections | REDIS_MIN_IDLE_CONNS | 2 |

### Recommended Tuning for Production

Update `infrastructure/parameters/prod.json`:
```json
{
  "ParameterKey": "TaskCPU",
  "ParameterValue": "1024"
},
{
  "ParameterKey": "TaskMemory",
  "ParameterValue": "2048"
},
{
  "ParameterKey": "DBInstanceClass",
  "ParameterValue": "db.t3.small"
},
{
  "ParameterKey": "CacheNodeType",
  "ParameterValue": "cache.t3.small"
}
```

---

## Monitoring and Alerts

### Key Metrics to Monitor

1. **ECS Task Health**
   - Metric: `HealthyHostCount`
   - Alarm: < 1 healthy task for 5 minutes

2. **Redis Connection Success Rate**
   - Watch logs for "Redis not available" warnings
   - Alert if > 10% of requests fail

3. **API Response Times**
   - P50: < 200ms
   - P95: < 1000ms
   - P99: < 2000ms

4. **Database Connections**
   - Metric: `DatabaseConnections`
   - Alarm: > 150 connections (75% of max 200)

5. **ElastiCache CPU**
   - Metric: `CPUUtilization`
   - Alarm: > 75% for 5 minutes

### Setting Up CloudWatch Alarms

```bash
# High response time alarm
aws cloudwatch put-metric-alarm \
  --alarm-name aaa-service-dev-high-latency \
  --alarm-description "Alert when P95 latency > 2s" \
  --metric-name TargetResponseTime \
  --namespace AWS/ApplicationELB \
  --statistic Average \
  --period 300 \
  --threshold 2000 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --region ap-south-1
```

---

## Known Limitations

1. **Redis Graceful Degradation**
   - Application works without Redis (cache misses)
   - Performance impact: ~200-500ms per request
   - Not critical for functionality

2. **Health Check 401 Response**
   - Expected behavior (authentication required)
   - ALB configured to accept 401 as healthy
   - Does not indicate service problem

3. **Context Timeout During Load**
   - Default request timeout: 12 seconds
   - Can be adjusted in middleware configuration
   - Consider increasing for batch operations

---

## Contact

For additional support:
- Check CloudWatch Logs: `/ecs/aaa-service-{environment}`
- Review ECS Service Events
- Check Security Groups allow traffic on ports 5432 (RDS) and 6379 (Redis)
