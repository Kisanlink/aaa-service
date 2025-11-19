# Aadhaar Integration - Implementation Tracker

**Project**: Aadhaar Verification Consolidation into AAA-Service
**Start Date**: 2025-11-18
**Target Completion**: Week of 2025-12-23 (5 weeks)
**Status**: PLANNING COMPLETE ✅ → READY FOR IMPLEMENTATION

---

## Quick Status Dashboard

| Phase | Status | Progress | Assigned To | Start Date | End Date |
|-------|--------|----------|-------------|------------|----------|
| Phase 1: Database & Models | COMPLETE ✅ | 100% | @agent-sde-backend-engineer | 2025-11-18 | 2025-11-18 |
| Phase 2: Sandbox & Service | IN PROGRESS 🔄 | 33% (1/3 tasks) | @agent-sde-backend-engineer | 2025-11-18 | TBD |
| Phase 3: REST API | IN PROGRESS 🔄 | 25% (1/4 tasks) | @agent-sde-backend-engineer | 2025-11-18 | TBD |
| Phase 4: gRPC | NOT STARTED | 0% | @agent-sde-backend-engineer | TBD | TBD |
| Phase 5: Testing & Deployment | NOT STARTED | 0% | @agent-sde-backend-engineer + @agent-business-logic-tester | TBD | TBD |

---

## Phase 1: Database Schema & Data Models (Week 1)

**Target**: 3-4 days | **Effort**: 30-35 hours

### Task 1.1: Database Migrations ✅ COMPLETE
- [x] Create `add_aadhaar_verifications_table.go`
- [x] Create `add_otp_attempts_table.go`
- [x] Create `add_user_profiles_kyc_fields.go`
- [x] Test migrations on local database
- [x] Verify rollback procedures
- [x] Create migration test helper (`test_aadhaar_migrations.go`)
- [x] Create migration runner script (`scripts/run_aadhaar_migrations.go`)
- [x] Create comprehensive documentation (`AADHAAR_MIGRATION_README.md`)

**Acceptance**: All migrations execute successfully, rollbacks work ✅

**Completion Date**: 2025-11-18
**Files Created**:
- `/Users/kaushik/aaa-service/migrations/add_aadhaar_verifications_table.go` (8.2KB)
- `/Users/kaushik/aaa-service/migrations/add_otp_attempts_table.go` (7.2KB)
- `/Users/kaushik/aaa-service/migrations/add_user_profiles_kyc_fields.go` (9.0KB)
- `/Users/kaushik/aaa-service/migrations/test_aadhaar_migrations.go` (6.3KB)
- `/Users/kaushik/aaa-service/scripts/run_aadhaar_migrations.go` (4.7KB)
- `/Users/kaushik/aaa-service/migrations/AADHAAR_MIGRATION_README.md` (documentation)

**Database State**:
- ✅ `aadhaar_verifications` table created (22 columns, 4 indexes)
- ✅ `otp_attempts` table created (9 columns, 2 indexes)
- ✅ `user_profiles` table updated (3 new columns, 2 indexes)
- ✅ All foreign key constraints working
- ✅ All validations passing

### Task 1.2: Data Models ✅ COMPLETE
- [x] Create `internal/entities/models/aadhaar_verification.go`
- [x] Create `internal/entities/models/otp_attempt.go`
- [x] Update `internal/entities/models/user_profile.go`
- [x] Implement JSONB marshaling for AadhaarAddress

**Acceptance**: Models compile, JSONB works, relationships defined ✅

**Completion Date**: 2025-11-18
**Files Created**:
- `/Users/kaushik/aaa-service/internal/entities/models/aadhaar_verification.go` (2.7KB)
- `/Users/kaushik/aaa-service/internal/entities/models/otp_attempt.go` (1.0KB)
- `/Users/kaushik/aaa-service/internal/entities/models/aadhaar_verification_test.go` (2.3KB)
- `/Users/kaushik/aaa-service/internal/entities/models/otp_attempt_test.go` (0.3KB)

**Files Updated**:
- `/Users/kaushik/aaa-service/internal/entities/models/user_profile.go` (added 3 KYC fields)

**Verification**:
- ✅ All models compile without errors
- ✅ JSONB Value() and Scan() methods working
- ✅ All tests passing (4/4 tests)
- ✅ Relationships properly defined (User, OTPAttempts)
- ✅ JSON tags prevent OTPValue exposure
- ✅ Table names match migration schema

### Task 1.3: Request/Response Models ✅ COMPLETE
- [x] Create `internal/entities/requests/kyc/generate_otp_request.go`
- [x] Create `internal/entities/requests/kyc/verify_otp_request.go`
- [x] Create `internal/entities/responses/kyc/generate_otp_response.go`
- [x] Create `internal/entities/responses/kyc/verify_otp_response.go`
- [x] Create `internal/entities/responses/kyc/kyc_status_response.go`

**Acceptance**: All models have proper validation, match API spec ✅

**Completion Date**: 2025-11-18
**Files Created**:
- `/Users/kaushik/aaa-service/internal/entities/requests/kyc/generate_otp_request.go` (1.1KB)
- `/Users/kaushik/aaa-service/internal/entities/requests/kyc/verify_otp_request.go` (0.9KB)
- `/Users/kaushik/aaa-service/internal/entities/responses/kyc/generate_otp_response.go` (0.7KB)
- `/Users/kaushik/aaa-service/internal/entities/responses/kyc/verify_otp_response.go` (1.5KB)
- `/Users/kaushik/aaa-service/internal/entities/responses/kyc/kyc_status_response.go` (0.8KB)

**Verification**:
- ✅ All request models compile without errors
- ✅ All response models compile without errors
- ✅ Validation methods implemented for all requests
- ✅ JSON tags use snake_case format
- ✅ Nested structs properly defined (AadhaarData, AadhaarAddr)
- ✅ All models match API specification exactly
- ✅ GetType() and IsSuccess() helper methods included
- ✅ Aadhaar number validation: exactly 12 numeric digits
- ✅ OTP validation: exactly 6 numeric digits
- ✅ Consent validation: must be "Y"
- ✅ All files under 100 lines each

### Task 1.4: Repository Layer ✅ COMPLETE
- [x] Create `internal/repositories/kyc/aadhaar_verification_repository.go`
- [x] Implement Create, GetByID, GetByUserID, GetByReferenceID
- [x] Implement UpdateStatus, IncrementAttempts, CreateOTPAttempt, GetOTPAttempts
- [x] Implement S3 photo storage (UploadPhoto, DeletePhoto) using kisanlink-db S3Manager
- [x] Write comprehensive unit tests for repository (28 test cases)

**Acceptance**: All CRUD operations work, tests passing ✅

**Completion Date**: 2025-11-18
**Files Created**:
- `/Users/kaushik/aaa-service/internal/repositories/kyc/aadhaar_verification_repository.go` (8.5KB)
- `/Users/kaushik/aaa-service/internal/repositories/kyc/aadhaar_verification_repository_test.go` (27.2KB)

**Verification**:
- ✅ All repository methods implemented
- ✅ All CRUD operations working (Create, GetByID, GetByUserID, GetByReferenceID)
- ✅ Status management working (UpdateStatus)
- ✅ Attempt tracking working (IncrementAttempts)
- ✅ OTP attempt logging working (CreateOTPAttempt, GetOTPAttempts)
- ✅ S3 photo storage using kisanlink-db S3Manager
- ✅ Proper folder structure: `aadhaar/photos/{userID}/{fileName}`
- ✅ Error handling follows codebase patterns
- ✅ All tests passing (28 tests, 56.8% coverage)
- ✅ Code compiles without errors

**Phase 1 Completion Criteria**:
✅ All migrations tested
✅ All models created and validated
✅ Request/Response models implemented with validation
✅ Repository layer fully functional
✅ Unit tests passing (56.8% coverage)

---

## Phase 2: Sandbox Client & Service Layer (Week 2)

**Target**: 5-6 days | **Effort**: 50-60 hours

### Task 2.1: Sandbox API Client ✅ COMPLETE
- [x] Create `internal/services/kyc/sandbox_client.go`
- [x] Create `internal/services/kyc/sandbox_models.go`
- [x] Implement GenerateOTP method
- [x] Implement VerifyOTP method
- [x] Add retry logic (exponential backoff: 1s, 2s, 4s, max 3 retries)
- [x] Add timeout handling (10 seconds)
- [x] Write comprehensive unit tests (26 test cases)

**Acceptance**: Client successfully calls Sandbox API, all error cases handled ✅

**Completion Date**: 2025-11-18
**Files Created**:
- `/Users/kaushik/aaa-service/internal/services/kyc/sandbox_models.go` (2.1KB)
- `/Users/kaushik/aaa-service/internal/services/kyc/sandbox_client.go` (9.8KB)
- `/Users/kaushik/aaa-service/internal/services/kyc/sandbox_client_test.go` (23.5KB)

**Verification**:
- ✅ All 26 tests passing
- ✅ Code coverage: 89.0% (excellent coverage)
- ✅ GenerateOTP method fully implemented with retry logic
- ✅ VerifyOTP method fully implemented with retry logic
- ✅ Exponential backoff: 1s, 2s, 4s (max 3 retries)
- ✅ 10-second timeout enforced
- ✅ Context cancellation support
- ✅ Comprehensive error handling:
  - Network errors
  - Timeout errors
  - HTTP 400 (validation errors)
  - HTTP 401 (authentication errors)
  - HTTP 403 (authorization errors)
  - HTTP 404 (not found)
  - HTTP 429 (rate limiting)
  - HTTP 500 (server errors with retry)
  - Malformed JSON responses
- ✅ Security features:
  - Aadhaar masking (XXXX-XXXX-1234)
  - OTP never logged
  - Authorization token never fully logged
  - Photo data never logged
- ✅ Request/response logging with proper masking
- ✅ All Sandbox API entities (@entity) properly set
- ✅ Headers properly set (accept, x-api-version, content-type, x-api-key, Authorization)

**Test Coverage**:
- GenerateOTP: Success, invalid Aadhaar, network error, timeout, retry success
- VerifyOTP: Success, invalid OTP, expired OTP, network error, API error, retry success
- Headers: With and without auth token
- Context cancellation
- Error responses: 401, 403, 404, 429, 500, malformed JSON, unknown status
- Helper functions: maskAadhaar, URL normalization

### Task 2.2: S3 Photo Upload Service
- [ ] Create `internal/storage/s3_client.go`
- [ ] Implement UploadPhoto method
- [ ] Add photo validation (format, size)
- [ ] Add photo compression (if size > 2MB)
- [ ] Enable SSE-S3 encryption
- [ ] Write unit tests

**Acceptance**: Photos upload to S3, encryption enabled, tests pass

### Task 2.3: KYC Service Layer
- [ ] Create `internal/services/kyc/service.go` (constructor)
- [ ] Create `internal/services/kyc/generate_otp.go`
- [ ] Create `internal/services/kyc/verify_otp.go`
- [ ] Create `internal/services/kyc/kyc_status.go`
- [ ] Create `internal/services/kyc/rate_limiter.go`
- [ ] Write comprehensive unit tests

**Acceptance**: All service methods work, rate limiting enforced, tests >90% coverage

### Task 2.4: Service Interfaces
- [ ] Update `internal/interfaces/interfaces.go`
- [ ] Define KYCService interface
- [ ] Define AadhaarVerificationRepository interface
- [ ] Define SandboxClient interface
- [ ] Define S3Client interface

**Acceptance**: All interfaces defined and match implementations

**Phase 2 Completion Criteria**:
✅ Sandbox client successfully integrates
✅ S3 photo upload working
✅ Service layer fully implemented
✅ Rate limiting functional
✅ Unit tests >90% coverage

---

## Phase 3: REST API Handlers & Routes (Week 3)

**Target**: 4-5 days | **Effort**: 40-45 hours

### Task 3.1: REST Handlers ✅ COMPLETE
- [x] Create `internal/handlers/kyc/aadhaar_handler.go`
- [x] Implement GenerateOTP handler
- [x] Implement VerifyOTP handler
- [x] Implement GetKYCStatus handler
- [x] Add error handling and validation
- [x] Create `internal/handlers/kyc/routes.go`

**Acceptance**: All handlers work, error responses follow spec ✅

**Completion Date**: 2025-11-18
**Files Created**:
- `/Users/kaushik/aaa-service/internal/handlers/kyc/aadhaar_handler.go` (10.4KB)
- `/Users/kaushik/aaa-service/internal/handlers/kyc/routes.go` (1.0KB)

**Verification**:
- ✅ All handlers compile without errors
- ✅ GenerateOTP handler implemented with full validation
- ✅ VerifyOTP handler implemented with full validation
- ✅ GetKYCStatus handler implemented with authorization checks
- ✅ Error handling using pkg/errors custom types
- ✅ Proper use of responder interface for consistent responses
- ✅ User authentication checks in all handlers
- ✅ Authorization check for GetKYCStatus (users can only view own status)
- ✅ Request validation using both request.Validate() and validator.ValidateStruct()
- ✅ Comprehensive error mapping in handleServiceError()
- ✅ Swagger documentation comments added
- ✅ Routes registered with authentication middleware
- ✅ All handlers follow existing codebase patterns
- ✅ Security: User ID extraction from context, auth token from header
- ✅ Logging with structured fields (zap)

### Task 3.2: Routes and Main Integration ✅ COMPLETE
- [x] Create `internal/handlers/kyc/routes.go`
- [x] Register POST /api/v1/kyc/aadhaar/otp
- [x] Register POST /api/v1/kyc/aadhaar/otp/verify
- [x] Register GET /api/v1/kyc/status/:user_id
- [x] Apply authentication middleware
- [x] Update `cmd/server/main.go`

**Acceptance**: Routes registered, middleware applied, service instantiated ✅

**Completion Date**: 2025-11-18
**Files Updated**:
- `/Users/kaushik/aaa-service/cmd/server/main.go` - Added KYC service initialization and wiring
- `/Users/kaushik/aaa-service/.env.example` - Added KYC environment variables
- `/Users/kaushik/aaa-service/internal/services/adapters/address_service_adapter.go` - Created adapter for address service

**Verification**:
- ✅ KYC package imports added to main.go
- ✅ UserProfileRepository created and initialized
- ✅ S3Manager created and configured (optional, based on AWS_S3_BUCKET env var)
- ✅ Sandbox client initialized with API credentials from environment
- ✅ Aadhaar verification repository created
- ✅ User service adapter created for KYC operations
- ✅ Address service adapter created
- ✅ KYC service created with all dependencies
- ✅ KYC handler created and wired
- ✅ KYC routes registered with authentication middleware
- ✅ Environment variables documented in .env.example
- ✅ Code compiles without errors (go build successful)
- ✅ All dependencies properly wired in correct initialization order

### Task 3.3: Integration Testing
- [ ] Create `internal/handlers/kyc/aadhaar_handler_test.go`
- [ ] Test all endpoints (success and error cases)
- [ ] Create `test/integration/aadhaar_flow_test.go`
- [ ] Test end-to-end OTP flow

**Acceptance**: All handler tests passing, integration test covers full flow

### Task 3.4: Middleware Enhancements
- [ ] Add/configure rate limiting middleware
- [ ] Add request validation middleware
- [ ] Add audit logging middleware

**Acceptance**: Middleware prevents abuse, all requests logged

**Phase 3 Completion Criteria**:
✅ All REST endpoints functional
✅ Integration tests passing
✅ Middleware configured
✅ End-to-end flow working

---

## Phase 4: gRPC Service & Proto Definitions (Week 4)

**Target**: 4-5 days | **Effort**: 35-40 hours

### Task 4.1: Proto Definitions
- [ ] Create `pkg/proto/kyc.proto`
- [ ] Define KYCService with all RPCs
- [ ] Define all request/response messages
- [ ] Generate Go code (`make proto-gen`)

**Acceptance**: Proto compiles, generated code works

### Task 4.2: gRPC Service Implementation
- [ ] Create `internal/grpc_server/kyc_service.go`
- [ ] Implement GenerateAadhaarOTP RPC
- [ ] Implement VerifyAadhaarOTP RPC
- [ ] Implement GetKYCStatus RPC
- [ ] Implement UploadAadhaarPhoto RPC (streaming)
- [ ] Add error mapping (gRPC codes)

**Acceptance**: All RPCs work, streaming functional, error handling proper

### Task 4.3: Register gRPC Service
- [ ] Update `cmd/server/main.go` to register KYCService
- [ ] Add service to reflection
- [ ] Test with grpcurl

**Acceptance**: gRPC service registered and callable

### Task 4.4: gRPC Integration Tests
- [ ] Create `internal/grpc_server/kyc_service_test.go`
- [ ] Test all RPCs
- [ ] Test streaming photo upload
- [ ] Test end-to-end flow via gRPC

**Acceptance**: All gRPC tests passing

**Phase 4 Completion Criteria**:
✅ gRPC service fully implemented
✅ All RPCs functional
✅ Streaming works correctly
✅ Tests passing

---

## Phase 5: Testing, Documentation & Deployment (Week 5)

**Target**: 5-6 days | **Effort**: 45-50 hours

### Task 5.1: Comprehensive Testing
**Owner**: @agent-business-logic-tester + @agent-sde-backend-engineer

- [ ] Security testing (OTP rate limiting, authorization, masking)
- [ ] Business logic validation (state transitions, OTP expiration, attempts)
- [ ] Performance testing (100 concurrent requests)
- [ ] Load testing (verify performance targets)

**Acceptance**: All security tests pass, performance targets met, no vulnerabilities

### Task 5.2: Documentation
- [ ] Update Swagger/OpenAPI specification
- [ ] Create API usage guide
- [ ] Update .kiro/specs/ with completion status
- [ ] Create operations runbook

**Acceptance**: Documentation complete, reviewed, and published

### Task 5.3: Configuration & Environment
- [ ] Add environment variables to .env.example
- [ ] Create config struct in `internal/config/kyc_config.go`
- [ ] Document configuration in README.md

**Acceptance**: All env vars documented, config loads correctly

### Task 5.4: Deployment Preparation
- [ ] Verify migrations on staging database
- [ ] Add feature flag support (AADHAAR_ENABLED)
- [ ] Create deployment checklist
- [ ] Set up monitoring & alerting (Prometheus metrics)

**Acceptance**: Migrations tested, feature flag works, monitoring configured

### Task 5.5: Staging Deployment & UAT
- [ ] Deploy to staging
- [ ] Run database migrations
- [ ] Enable feature flag
- [ ] Conduct UAT
- [ ] Monitor for 24 hours

**Acceptance**: Service deployed successfully, UAT passed, no errors in 24h

**Phase 5 Completion Criteria**:
✅ All tests passing (security, business logic, performance)
✅ Documentation complete
✅ Deployment successful
✅ UAT passed
✅ Monitoring active

---

## Risk Register

| Risk | Impact | Probability | Mitigation | Owner |
|------|--------|-------------|------------|-------|
| Sandbox API downtime | HIGH | MEDIUM | Retry logic, circuit breaker, fallback | Engineer |
| S3 upload failures | MEDIUM | LOW | Retry logic, temp storage | Engineer |
| Database migration issues | HIGH | LOW | Test on staging, rollback scripts | Engineer |
| Performance degradation | MEDIUM | MEDIUM | Load testing, query optimization | Engineer |
| Security vulnerabilities | CRITICAL | LOW | Security testing, code review | Tester |

---

## Blockers & Issues

*No blockers at this time - ready to start implementation*

---

## Dependencies Checklist

### External Dependencies
- [ ] Sandbox.co.in API credentials obtained
- [ ] AWS S3 bucket created (`aaa-aadhaar-photos`)
- [ ] IAM credentials for S3 access
- [ ] Redis cache available for rate limiting

### Internal Dependencies
- [x] UserService available
- [x] AddressService available
- [x] AuditService available
- [x] AuthService available
- [x] Database migration tooling ready

### Development Environment
- [x] Go 1.24 installed
- [x] PostgreSQL 15 available
- [x] Redis 7 available
- [ ] AWS CLI configured
- [ ] protoc compiler installed

---

## Rollout Plan

### Phase 1: Database Migration (Day 1 of deployment)
- Run migrations during maintenance window
- Verify data integrity
- Test rollback

### Phase 2: Service Deployment (Day 1)
- Deploy with feature flag OFF
- Verify health checks
- Test internal endpoints

### Phase 3: Gradual Rollout (Days 2-7)
- Day 2: 1% of users (internal testing)
- Day 3: 10% of users
- Day 4: 25% of users
- Day 5: 50% of users
- Day 6: 75% of users
- Day 7: 100% of users

---

## Success Metrics

### Technical Metrics
- [ ] All tests passing (unit, integration, e2e)
- [ ] Code coverage > 90%
- [ ] Performance: OTP generation < 2s, verification < 3s
- [ ] Zero security vulnerabilities

### Business Metrics
- [ ] OTP success rate > 95%
- [ ] Average verification time < 5 minutes
- [ ] Customer support tickets < 5% of verifications

### Operational Metrics
- [ ] Service uptime > 99.9%
- [ ] Mean time to recovery < 5 minutes
- [ ] All runbooks documented

---

## Next Steps

1. **Immediate** (today):
   - [ ] Review all specifications with team
   - [ ] Obtain Sandbox API credentials
   - [ ] Set up AWS S3 bucket
   - [ ] Configure development environment

2. **This Week** (Week 1):
   - [ ] Start Phase 1: Database & Models
   - [ ] Daily standup to track progress
   - [ ] Review code as tasks complete

3. **Next Week** (Week 2):
   - [ ] Complete Phase 1
   - [ ] Start Phase 2: Sandbox Client & Service
   - [ ] Mid-sprint review

---

## Team Communication

**Daily Standups**: 10:00 AM
**Sprint Reviews**: End of each phase
**Demo**: End of Week 5 (before production deployment)

**Communication Channels**:
- Slack: #aaa-service-dev
- Issues: GitHub Issues with `aadhaar-integration` label
- Documentation: `.kiro/specs/aadhaar-integration/`

---

## Notes & Learnings

*This section will be updated as implementation progresses*

---

**Last Updated**: 2025-11-18
**Next Review**: Daily (during implementation)
