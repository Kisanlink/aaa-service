# Kisanlink Old Infrastructure Cleanup & FPO Billing

**Goal:** Safely delete old 3-VPC infrastructure after shared VPC migration + calculate per-FPO costs for billing
**Region:** ap-south-1 (Mumbai) | **Account:** 859086474730
**Timing:** After 7 days of stable operation on new shared VPC

---

## Part 1: Cost Per FPO Breakdown (For Billing)

### Shared Platform Costs (Fixed, regardless of FPO count)

| Resource | Spec | Monthly Cost |
|----------|------|--------------|
| NAT Gateway | 1x (single AZ) | $32 |
| Shared ALB | 1x | $22 |
| VPC Endpoints | 4x (ECR, Secrets, Logs, S3) | $29 |
| ECS Fargate (AAA) | 0.5 vCPU / 1GB | $15 |
| RDS AAA | db.t4g.micro | $12 |
| ElastiCache Redis | cache.t4g.micro | $11 |
| ECS Fargate (Farmers) | 1 vCPU / 2GB | $30 |
| RDS Farmers | db.t4g.small (PostGIS) | $24 |
| CloudWatch Logs | ~5GB/month | $3 |
| **Shared Total** | | **$178/month** |

### Per-FPO Direct Costs (Each FPO adds this)

| Resource | Spec | Monthly Cost |
|----------|------|--------------|
| ECS Fargate (ERP) | 0.5 vCPU / 1GB | $15 |
| RDS (ERP DB) | db.t4g.micro, 20GB gp3 | $12 |
| **Per-FPO Direct Total** | | **$27/month** |

### Billing Model Options

**Option A: Direct Cost Only (simplest)**
Bill each FPO only $27/month. Platform absorbs the $178/month shared cost.

**Option B: Platform Fee + Direct (recommended)**

| FPO Count | Shared/FPO | Direct/FPO | Total/FPO | Platform Total |
|-----------|------------|------------|-----------|----------------|
| 1 | $178.00 | $27 | $205 | $205 |
| 2 | $89.00 | $27 | $116 | $232 |
| 3 | $59.33 | $27 | $86 | $259 |
| 5 | $35.60 | $27 | $63 | $313 |
| 10 | $17.80 | $27 | $45 | $448 |
| 20 | $8.90 | $27 | $36 | $718 |
| 50 | $3.56 | $27 | $31 | $1,528 |

**Option C: Tiered Flat Rate (predictable for FPOs)**

| Tier | Range | Price/FPO/month |
|------|-------|-----------------|
| Starter | 1-5 FPOs | $75 |
| Growth | 6-15 FPOs | $50 |
| Scale | 16-50 FPOs | $35 |

### Key Economics
- Break-even vs old 3-VPC ($369/mo): Shared setup with <= 7 FPOs
- Marginal cost per additional FPO: **$27/month** (just ECS + RDS)
- No additional ALB, NAT, or VPC endpoint costs per FPO

---

## Part 2: Old Resource Inventory

### Old CloudFormation Stacks

| Stack Name Pattern | Service | Creates |
|-------------------|---------|---------|
| `aaa-service-beta` | AAA | VPC, ALB, ECS, RDS, Redis, SGs |
| `aaa-service-staging` | AAA | VPC, ALB, ECS, RDS, Redis, SGs |
| `aaa-service-prod` | AAA | VPC, ALB, ECS, RDS, Redis, SGs |
| `farmers-module-production` | Farmers | VPC, ALB, ECS, RDS, SGs |
| Per-FPO ERP stacks | ERP | ECS, RDS, TG, Listener Rule |

> Deleting the CloudFormation stack handles all child resources automatically in correct dependency order. This is the safest approach.

### Old Secrets Manager Secrets

| Secret Pattern | Service |
|---------------|---------|
| `aaa-service-{env}-db-credentials` | AAA |
| `aaa-service-{env}-jwt-secret` | AAA |
| `aaa-service-{env}-redis-auth-token` | AAA |
| `farmers-module/{env}/db-password` | Farmers |
| `farmers-module/{env}/aaa-api-key` | Farmers |
| `farmers-module/{env}/jwt-secret` | Farmers |
| `farmers-module/{env}/secret-key` | Farmers |

> Secrets are NOT deleted with CloudFormation stacks if they have DeletionPolicy: Retain. Keep until migration confirmed successful, then delete manually.

---

## Part 3: Safe Deletion Plan

### Prerequisites (BEFORE starting)

- [ ] New stack `kisanlink-shared-beta-infra` is `CREATE_COMPLETE`
- [ ] All ECS services healthy (1/1 running) in new cluster
- [ ] All /health endpoints responding on new URLs
- [ ] DNS (*.beta.kisanlink.in) points to NEW shared ALB
- [ ] Data migration complete (old DB data imported to new RDS)
- [ ] 7 days of stable operation on new infrastructure
- [ ] CloudWatch logs confirm no errors on new services
- [ ] Old services receiving ZERO traffic (check old ALB metrics)

### Phase 1: Verify Zero Traffic (Day 7-8)

Check `RequestCount` metric for each old ALB over last 24 hours:

| Old ALB | Expected RequestCount (24h) |
|---------|---------------------------|
| `aaa-service-beta-alb` | 0 |
| `farmers-module-beta-alb` | 0 |
| Old ERP ALBs | 0 |

**Console:** EC2 > Load Balancers > Select OLD ALB > Monitoring tab

If count > 0: DNS may still point to old ALB. Investigate before proceeding.

### Phase 2: Scale Down Old Services (Day 8)

Scale to 0 tasks first (reversible!) before deleting anything.

| Action | Console Path |
|--------|-------------|
| Scale AAA to 0 | ECS > `aaa-service-beta-cluster` > `aaa-service-beta` > Update > Desired: 0 |
| Scale Farmers to 0 | ECS > `farmers-module-beta-cluster` > `farmers-module-beta-service` > Update > Desired: 0 |
| Scale ERP(s) to 0 | ECS > [erp-cluster] > [erp-service] > Update > Desired: 0 |

**Wait 24 hours.** If problems arise, scale back up as rollback.

### Phase 3: Create Final RDS Snapshots (Day 9)

**Console:** RDS > Databases > Select instance > Actions > Take snapshot

| Old RDS Instance | Snapshot Name |
|-----------------|---------------|
| `aaa-service-beta-db` | `aaa-service-beta-db-final-YYYYMMDD` |
| `farmers-module-beta-postgres` | `farmers-module-beta-db-final-YYYYMMDD` |
| Old ERP RDS instances | `erp-{fpo}-beta-db-final-YYYYMMDD` |

Wait for all snapshots to show status **available**.

### Phase 4: Delete Old CloudFormation Stacks (Day 9-10)

**Console:** CloudFormation > Stacks > Select stack > Delete

**Deletion order (delete child stacks before parent):**

**Step 4.1:** Delete old ERP stacks first (no cross-stack dependencies)
**Step 4.2:** Delete old AAA stack (`aaa-service-beta`) - expect 15-30 min
**Step 4.3:** Delete old Farmers stack (`farmers-module-production`)

### Phase 5: Manual Cleanup (Day 10+)

| Resource | Console Path | Action |
|----------|-------------|--------|
| Old Elastic IPs | VPC > Elastic IPs | Release if unassociated |
| Old CloudWatch Logs | CloudWatch > Log groups | Delete `/ecs/aaa-service-*`, `/ecs/farmers-module-*` |
| Old Secrets | Secrets Manager | Schedule deletion (30-day recovery window) |
| Old ECR images | ECR > Repositories | Clean up untagged images (keep repos) |
| Old task definitions | ECS > Task Definitions | Deregister old families |

### Phase 6: CI/CD Pipeline Updates

**Already completed in this repo:**
- `.github/workflows/deploy-ecs.yml` - Updated cluster/service names
- `.github/workflows/deploy-ecs-advanced.yml` - Updated cluster/service names

**Still needed in other repos:**
- `fpo-erp/.github/workflows/deploy-erp.yml` - Update to shared cluster
- `farmers-module/.github/workflows/` - Update cluster/service references

Key naming changes:
- Cluster: `aaa-service-{env}-cluster` -> `kisanlink-shared-{env}-cluster`
- AAA Service: `aaa-service-{env}` -> `aaa-service-shared-{env}`
- Farmers Service: `farmers-module-{env}-service` -> `farmers-module-shared-{env}`
- Task Definition: `aaa-service-{env}` -> `aaa-service-shared-{env}`
- Region: `us-east-1` -> `ap-south-1`

---

## Part 4: Rollback Plan

### During Phase 2 (scaled to 0):
1. ECS > Old cluster > Old service > Update > Set desired count to 1
2. Update Route 53 to point back to old ALB
3. Service restores in ~2 minutes

### During Phase 4 (stacks being deleted):
- Cannot stop in-progress deletion
- RDS snapshots from Phase 3 are available
- Re-deploy from templates in git if needed

### After Phase 4 (stacks deleted):
- Old infra gone, but snapshots preserved
- Restore from snapshots in new shared VPC if needed

---

## Part 5: Checklist Summary

### Pre-Deletion
- [ ] New shared VPC fully operational (7+ days)
- [ ] DNS points to new ALB
- [ ] All /health endpoints green on new infra
- [ ] Old ALBs showing 0 requests for 24h

### Deletion (Beta Environment)
- [ ] Scale old AAA ECS service to 0
- [ ] Scale old Farmers ECS service to 0
- [ ] Scale old ERP ECS service(s) to 0
- [ ] Wait 24 hours, monitor new infra
- [ ] Create final RDS snapshots (AAA, Farmers, ERP)
- [ ] Delete old ERP CloudFormation stack(s)
- [ ] Delete old AAA CloudFormation stack
- [ ] Delete old Farmers CloudFormation stack
- [ ] Release orphaned Elastic IPs
- [ ] Schedule old Secrets Manager secrets for deletion
- [ ] Clean up old CloudWatch log groups
- [ ] Deregister old ECS task definition families
- [ ] Update CI/CD workflows in all repos

### Post-Deletion Verification
- [ ] No orphaned VPCs in VPC console
- [ ] No orphaned ECS clusters
- [ ] No orphaned ALBs
- [ ] RDS snapshots available for recovery
- [ ] CI/CD pipelines deploy successfully to new shared infra
