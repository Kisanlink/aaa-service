# ADR-001: Consolidate Aadhaar Verification into AAA-Service

## Status
**ACCEPTED** - 2025-11-18

## Context

The current architecture includes a standalone `aadhaar-verification` microservice that handles Aadhaar OTP generation and verification. This service communicates with the `aaa-service` via gRPC calls to update user profiles and addresses after successful verification.

### Current Architecture Issues

1. **Inter-Service Complexity**: gRPC calls between aadhaar-verification and aaa-service add latency and failure points
2. **Deployment Overhead**: Two separate services to maintain, deploy, and monitor
3. **Transaction Boundaries**: Difficult to maintain atomicity across service boundaries (OTP verification + profile update)
4. **Code Duplication**: Similar patterns duplicated across both services (error handling, logging, database access)
5. **Development Friction**: Changes require coordinating deployments across two services

### Business Context

- Aadhaar verification is a core KYC function, not a separate business domain
- The aadhaar-verification service is tightly coupled to aaa-service (all operations ultimately update aaa-service data)
- No other services consume aadhaar-verification (it's not truly a shared service)
- Team size doesn't justify microservice overhead (single team owns both services)

## Decision

**We will consolidate Aadhaar verification functionality into aaa-service as an internal KYC module.**

### What This Means

1. **Service Structure**: Aadhaar verification becomes a service layer within aaa-service at `internal/services/kyc/`
2. **API Strategy**: Support both REST and gRPC endpoints for maximum flexibility
3. **External Integration**: Direct integration with Sandbox.co.in API from aaa-service
4. **Photo Storage**: Direct S3 upload from aaa-service
5. **Transaction Management**: All operations (OTP verification, profile update, address creation) in single service context

## Consequences

### Positive

1. **Simplified Architecture**
   - One less service to deploy, monitor, and maintain
   - No inter-service network calls
   - Simpler dependency graph

2. **Better Performance**
   - Eliminate network overhead for internal operations
   - Single database transaction for profile updates
   - Reduced latency for end-to-end verification flow

3. **Improved Developer Experience**
   - Single codebase for related functionality
   - Easier to refactor and test
   - Simplified local development setup
   - Faster iteration cycles

4. **Cost Reduction**
   - One less service to run in production
   - Reduced infrastructure costs
   - Simpler CI/CD pipeline

5. **Transaction Integrity**
   - All KYC operations in single service context
   - Easier to maintain ACID properties
   - Simpler error handling and rollback

### Negative

1. **Service Size**
   - aaa-service becomes larger (more lines of code)
   - Slightly longer build times
   - More comprehensive test suite required

2. **Deployment Coupling**
   - Changes to KYC require full aaa-service deployment
   - Cannot scale KYC independently (though not needed based on current load)

3. **Team Organization**
   - If KYC team becomes separate in future, will need to re-evaluate
   - Harder to assign ownership if teams specialize

### Neutral

1. **API Compatibility**
   - Maintain same API contracts (REST + gRPC)
   - Clients won't be affected
   - Migration transparent to consumers

2. **External Dependencies**
   - Still depends on Sandbox.co.in API
   - Still depends on AWS S3
   - No change in external integration complexity

## Alternatives Considered

### Alternative 1: Keep Separate Microservice
**Rejected** because:
- Adds unnecessary complexity for current scale
- No clear benefit given tight coupling with aaa-service
- No other consumers of aadhaar-verification service
- Team size doesn't justify separate service

### Alternative 2: Use Event-Driven Architecture
**Rejected** because:
- Adds complexity without clear benefit
- KYC operations are synchronous by nature (user waits for result)
- Would still need tight integration with aaa-service
- Eventual consistency not acceptable for KYC operations

### Alternative 3: Merge as Separate Binary (Same Repo)
**Considered** but rejected because:
- Still have deployment overhead
- Still have inter-service communication
- Doesn't solve transaction boundary issues
- Minimal benefit over full consolidation

## Implementation Strategy

### Phase 1: Build KYC Module in aaa-service
- Create `internal/services/kyc/` module
- Implement Sandbox API client
- Implement S3 photo upload
- Create REST and gRPC endpoints

### Phase 2: Parallel Run
- Deploy new KYC endpoints in aaa-service
- Keep old aadhaar-verification service running
- Route 10% of traffic to new endpoints
- Monitor for errors and performance

### Phase 3: Full Migration
- Route 100% of traffic to aaa-service
- Update client applications
- Shut down aadhaar-verification service

### Phase 4: Cleanup
- Archive aadhaar-verification codebase
- Remove references to old service
- Update documentation

## Rollback Plan

If consolidation causes issues:

1. **Immediate Rollback** (< 5 minutes)
   - Set feature flag `AADHAAR_ENABLED=false`
   - Route traffic back to old service

2. **Full Rollback** (< 30 minutes)
   - Revert aaa-service deployment
   - Re-enable old aadhaar-verification service
   - Update routing rules

## Monitoring & Success Criteria

### Metrics to Track
- OTP generation success rate
- OTP verification success rate
- End-to-end verification latency
- Error rates
- Sandbox API call duration
- S3 upload success rate

### Success Criteria
- OTP success rate ≥ 95%
- Average verification time < 5 minutes
- Error rate < 2%
- No increase in support tickets
- Cost reduction of at least $X per month

## Decision Drivers

### Technical Factors
1. **Tight Coupling**: Aadhaar verification is tightly coupled to user management
2. **Transaction Boundaries**: Need atomic operations across verification and profile update
3. **Performance**: Eliminate network overhead
4. **Simplicity**: Single codebase easier to maintain

### Business Factors
1. **Team Size**: Single team owns both services
2. **Scale**: Current load doesn't require separate service
3. **Cost**: Infrastructure cost reduction
4. **Time to Market**: Faster development and deployment

### Organizational Factors
1. **DevOps Overhead**: One less service to manage
2. **Cognitive Load**: Simpler mental model for developers
3. **Documentation**: Easier to document integrated system

## Assumptions

1. **Load**: Current and projected load doesn't require independent scaling of KYC
2. **Team Structure**: Single team continues to own all AAA functionality
3. **External Dependencies**: Sandbox API and S3 remain reliable
4. **API Stability**: KYC API contracts won't change frequently

## Risks & Mitigation

### Risk 1: Service Size Grows Too Large
**Mitigation**:
- Monitor service complexity
- Keep modules well-organized
- Consider extraction if service becomes unwieldy

### Risk 2: Deployment Coupling Issues
**Mitigation**:
- Use feature flags for gradual rollout
- Maintain comprehensive test coverage
- Implement blue-green deployment

### Risk 3: Performance Degradation
**Mitigation**:
- Load test before production deployment
- Monitor performance metrics closely
- Optimize slow queries proactively

## Related Documents

- [Architecture Specification](./architecture.md)
- [Design Specification](./design.md)
- [Requirements](./requirements.md)
- [Implementation Plan](./implementation-plan.md)

## Notes

- This decision aligns with the "modular monolith" pattern
- We can always extract KYC as a separate service later if needed
- Focus on clear module boundaries within aaa-service
- Use Go interfaces to maintain testability

## Review & Approval

**Proposed by**: Development Team
**Reviewed by**: @agent-sde3-backend-architect
**Approved by**: Technical Lead
**Date**: 2025-11-18

## Revision History

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2025-11-18 | 1.0 | Initial decision | Development Team |
