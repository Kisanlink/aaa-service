# Aadhaar Integration Implementation Plan

## Overview

This document outlines the phased implementation plan for integrating Aadhaar verification into aaa-service. The plan is broken down into 5 phases over 4-5 weeks with clear tasks, acceptance criteria, and dependencies.

## Timeline Summary

```
Week 1: Phase 1 - Database & Models
Week 2: Phase 2 - Sandbox Client & Service Layer
Week 3: Phase 3 - Handlers & REST API
Week 4: Phase 4 - gRPC & Integration
Week 5: Phase 5 - Testing, Documentation & Deployment
```

Total Estimated Effort: **180-200 engineering hours**

---

## Phase 1: Database Schema & Data Models (Week 1)

**Duration:** 3-4 days
**Effort:** 30-35 hours
**Owner:** @agent-sde-backend-engineer

### Tasks

#### Task 1.1: Create Database Migrations
**Effort:** 6-8 hours

- [ ] Create migration file: `YYYYMMDDHHMM_create_aadhaar_verifications_table.go`
  - [ ] Define table structure with all columns
  - [ ] Add indexes (user_id, transaction_id, reference_id, status)
  - [ ] Add foreign key constraint to users table
  - [ ] Add rollback function

- [ ] Create migration file: `YYYYMMDDHHMM_create_otp_attempts_table.go`
  - [ ] Define table structure
  - [ ] Add indexes (aadhaar_verification_id, status)
  - [ ] Add foreign key constraint
  - [ ] Add rollback function

- [ ] Create migration file: `YYYYMMDDHHMM_update_user_profiles_kyc_fields.go`
  - [ ] Add aadhaar_verified column
  - [ ] Add aadhaar_verified_at column
  - [ ] Add kyc_status column
  - [ ] Add indexes
  - [ ] Add rollback function

- [ ] Test migrations on local database
- [ ] Verify rollback procedures work
- [ ] Document migration steps in migration-plan.md

**Acceptance Criteria:**
- All migrations execute successfully
- Rollbacks work without data loss
- Indexes created properly
- Foreign keys enforced

**Files:**
- `/migrations/YYYYMMDDHHMM_create_aadhaar_verifications_table.go`
- `/migrations/YYYYMMDDHHMM_create_otp_attempts_table.go`
- `/migrations/YYYYMMDDHHMM_update_user_profiles_kyc_fields.go`

#### Task 1.2: Create Data Models
**Effort:** 8-10 hours

- [ ] Create `internal/entities/models/aadhaar_verification.go`
  - [ ] Define AadhaarVerification struct
  - [ ] Define AadhaarAddress struct (JSONB support)
  - [ ] Implement JSONB marshaling/unmarshaling
  - [ ] Add GORM tags
  - [ ] Add validation tags
  - [ ] Add relationships (User, OTPAttempts)

- [ ] Create `internal/entities/models/otp_attempt.go`
  - [ ] Define OTPAttempt struct
  - [ ] Add GORM tags
  - [ ] Add relationship to AadhaarVerification

- [ ] Update `internal/entities/models/user_profile.go`
  - [ ] Add aadhaar_verified field
  - [ ] Add aadhaar_verified_at field
  - [ ] Add kyc_status field

**Acceptance Criteria:**
- Models compile without errors
- JSONB marshaling works correctly
- GORM tags properly configured
- Relationships defined

**Files:**
- `/internal/entities/models/aadhaar_verification.go`
- `/internal/entities/models/otp_attempt.go`
- `/internal/entities/models/user_profile.go` (updated)

#### Task 1.3: Create Request/Response Models
**Effort:** 6-8 hours

- [ ] Create `internal/entities/requests/kyc/generate_otp_request.go`
  - [ ] Define GenerateOTPRequest struct
  - [ ] Add validation tags (aadhaar_number: 12 digits, consent: "Y")

- [ ] Create `internal/entities/requests/kyc/verify_otp_request.go`
  - [ ] Define VerifyOTPRequest struct
  - [ ] Add validation tags (reference_id: required, otp: 6 digits)

- [ ] Create `internal/entities/responses/kyc/generate_otp_response.go`
  - [ ] Define GenerateOTPResponse struct
  - [ ] Include reference_id, transaction_id, expires_at

- [ ] Create `internal/entities/responses/kyc/verify_otp_response.go`
  - [ ] Define VerifyOTPResponse struct
  - [ ] Include aadhaar_data, profile_id, address_id

- [ ] Create `internal/entities/responses/kyc/kyc_status_response.go`
  - [ ] Define KYCStatusResponse struct
  - [ ] Include verification status fields

**Acceptance Criteria:**
- All request models have proper validation
- Response models match API specification
- Models compile without errors

**Files:**
- `/internal/entities/requests/kyc/generate_otp_request.go`
- `/internal/entities/requests/kyc/verify_otp_request.go`
- `/internal/entities/responses/kyc/generate_otp_response.go`
- `/internal/entities/responses/kyc/verify_otp_response.go`
- `/internal/entities/responses/kyc/kyc_status_response.go`

#### Task 1.4: Create Repository Layer
**Effort:** 10-12 hours

- [ ] Create `internal/repositories/kyc/aadhaar_verification_repository.go`
  - [ ] Define AadhaarVerificationRepository interface
  - [ ] Implement Create(ctx, *models.AadhaarVerification) error
  - [ ] Implement GetByID(ctx, id) (*models.AadhaarVerification, error)
  - [ ] Implement GetByUserID(ctx, userID) (*models.AadhaarVerification, error)
  - [ ] Implement GetByReferenceID(ctx, refID) (*models.AadhaarVerification, error)
  - [ ] Implement UpdateStatus(ctx, id, status) error
  - [ ] Implement IncrementAttempts(ctx, id) error
  - [ ] Implement CreateOTPAttempt(ctx, *models.OTPAttempt) error

- [ ] Write unit tests for repository
  - [ ] Test Create with valid data
  - [ ] Test GetByUserID with existing/non-existing user
  - [ ] Test GetByReferenceID
  - [ ] Test UpdateStatus
  - [ ] Test IncrementAttempts

**Acceptance Criteria:**
- All CRUD operations implemented
- Repository uses kisanlink-db patterns
- All tests passing
- Error handling for not found cases

**Files:**
- `/internal/repositories/kyc/aadhaar_verification_repository.go`
- `/internal/repositories/kyc/aadhaar_verification_repository_test.go`

---

## Phase 2: Sandbox Client & Service Layer (Week 2)

**Duration:** 5-6 days
**Effort:** 50-60 hours
**Owner:** @agent-sde-backend-engineer

### Tasks

#### Task 2.1: Implement Sandbox API Client
**Effort:** 12-15 hours

- [ ] Create `internal/services/kyc/sandbox_client.go`
  - [ ] Define SandboxClient struct
  - [ ] Implement NewSandboxClient(baseURL, apiKey, apiSecret)
  - [ ] Implement GenerateOTP(ctx, aadhaarNumber, consent) (*SandboxOTPResponse, error)
  - [ ] Implement VerifyOTP(ctx, referenceID, otp) (*SandboxVerifyResponse, error)
  - [ ] Add request/response logging
  - [ ] Add error handling for API errors
  - [ ] Add retry logic (exponential backoff, max 3 retries)
  - [ ] Add timeout handling (10 seconds)

- [ ] Create `internal/services/kyc/sandbox_models.go`
  - [ ] Define SandboxOTPResponse struct
  - [ ] Define SandboxVerifyResponse struct
  - [ ] Define KYCData struct

- [ ] Write unit tests with mock HTTP server
  - [ ] Test successful OTP generation
  - [ ] Test OTP generation with API error
  - [ ] Test successful OTP verification
  - [ ] Test OTP verification with invalid OTP
  - [ ] Test timeout handling
  - [ ] Test retry logic

**Acceptance Criteria:**
- Client successfully calls Sandbox API
- All error cases handled
- Retry logic works correctly
- Tests pass with mock server

**Files:**
- `/internal/services/kyc/sandbox_client.go`
- `/internal/services/kyc/sandbox_models.go`
- `/internal/services/kyc/sandbox_client_test.go`

#### Task 2.2: Implement S3 Photo Upload Service
**Effort:** 8-10 hours

- [ ] Create `internal/storage/s3_client.go`
  - [ ] Define S3Client struct
  - [ ] Implement NewS3Client(bucketName, region, accessKey, secretKey)
  - [ ] Implement UploadPhoto(ctx, userID, photoBase64, fileName) (string, error)
  - [ ] Add photo validation (format, size)
  - [ ] Add photo compression if size > 2MB
  - [ ] Enable SSE-S3 encryption
  - [ ] Return S3 URL

- [ ] Write unit tests
  - [ ] Test successful upload
  - [ ] Test invalid format
  - [ ] Test oversized file
  - [ ] Test compression
  - [ ] Test S3 error handling

**Acceptance Criteria:**
- Photos successfully upload to S3
- Encryption enabled
- Compression works for large files
- Tests pass

**Files:**
- `/internal/storage/s3_client.go`
- `/internal/storage/s3_client_test.go`

#### Task 2.3: Implement KYC Service Layer
**Effort:** 20-25 hours

- [ ] Create `internal/services/kyc/service.go`
  - [ ] Define Service struct with dependencies
  - [ ] Implement NewService constructor
  - [ ] Define Config struct for OTP settings

- [ ] Create `internal/services/kyc/generate_otp.go`
  - [ ] Implement GenerateOTP(ctx, req) (resp, error)
  - [ ] Validate request
  - [ ] Check rate limits
  - [ ] Call SandboxClient.GenerateOTP()
  - [ ] Create AadhaarVerification record (status=PENDING)
  - [ ] Return response with reference_id

- [ ] Create `internal/services/kyc/verify_otp.go`
  - [ ] Implement VerifyOTP(ctx, req) (resp, error)
  - [ ] Validate request
  - [ ] Check attempt limits
  - [ ] Call SandboxClient.VerifyOTP()
  - [ ] Upload photo to S3
  - [ ] Update user profile
  - [ ] Create/update address
  - [ ] Update verification status to VERIFIED
  - [ ] Record OTP attempt
  - [ ] Return response

- [ ] Create `internal/services/kyc/kyc_status.go`
  - [ ] Implement GetKYCStatus(ctx, userID) (resp, error)
  - [ ] Query aadhaar_verifications table
  - [ ] Return status response

- [ ] Create `internal/services/kyc/rate_limiter.go`
  - [ ] Implement CheckRateLimit(ctx, aadhaarNumber) error
  - [ ] Check attempts in last 60 seconds
  - [ ] Return error if limit exceeded

- [ ] Write comprehensive unit tests
  - [ ] Test GenerateOTP success
  - [ ] Test GenerateOTP with rate limit exceeded
  - [ ] Test VerifyOTP success
  - [ ] Test VerifyOTP with invalid OTP
  - [ ] Test VerifyOTP with max attempts exceeded
  - [ ] Test GetKYCStatus

**Acceptance Criteria:**
- All service methods implemented
- Rate limiting works
- Integration with UserService, AddressService works
- All unit tests passing
- Code coverage > 90%

**Files:**
- `/internal/services/kyc/service.go`
- `/internal/services/kyc/generate_otp.go`
- `/internal/services/kyc/verify_otp.go`
- `/internal/services/kyc/kyc_status.go`
- `/internal/services/kyc/rate_limiter.go`
- `/internal/services/kyc/service_test.go`

#### Task 2.4: Create Service Interfaces
**Effort:** 4-5 hours

- [ ] Update `internal/interfaces/interfaces.go`
  - [ ] Define KYCService interface
  - [ ] Define AadhaarVerificationRepository interface
  - [ ] Define SandboxClient interface (for mocking)
  - [ ] Define S3Client interface (for mocking)

**Acceptance Criteria:**
- All interfaces defined
- Interfaces match implementations

**Files:**
- `/internal/interfaces/interfaces.go` (updated)

---

## Phase 3: REST API Handlers & Routes (Week 3)

**Duration:** 4-5 days
**Effort:** 40-45 hours
**Owner:** @agent-sde-backend-engineer

### Tasks

#### Task 3.1: Create REST Handlers
**Effort:** 15-18 hours

- [ ] Create `internal/handlers/kyc/aadhaar_handler.go`
  - [ ] Define Handler struct with KYCService dependency
  - [ ] Implement NewHandler constructor
  - [ ] Implement GenerateOTP handler
    - [ ] Parse request body
    - [ ] Validate request
    - [ ] Call service.GenerateOTP()
    - [ ] Return JSON response
  - [ ] Implement VerifyOTP handler
    - [ ] Parse request body
    - [ ] Extract user_id from context/token
    - [ ] Validate request
    - [ ] Call service.VerifyOTP()
    - [ ] Return JSON response
  - [ ] Implement GetKYCStatus handler
    - [ ] Extract user_id from path parameter
    - [ ] Validate authorization (user can only view own status or admin)
    - [ ] Call service.GetKYCStatus()
    - [ ] Return JSON response

- [ ] Add error handling
  - [ ] Map service errors to HTTP status codes
  - [ ] Return structured error responses
  - [ ] Log errors with context

**Acceptance Criteria:**
- All handlers implemented
- Request validation works
- Error responses follow API spec
- Authorization checks in place

**Files:**
- `/internal/handlers/kyc/aadhaar_handler.go`

#### Task 3.2: Create Routes
**Effort:** 6-8 hours

- [ ] Create `internal/handlers/kyc/routes.go`
  - [ ] Define RegisterKYCRoutes(router, handler, authMiddleware)
  - [ ] Add POST /api/v1/kyc/aadhaar/otp
  - [ ] Add POST /api/v1/kyc/aadhaar/otp/verify
  - [ ] Add GET /api/v1/kyc/status/:user_id
  - [ ] Apply authentication middleware
  - [ ] Apply rate limiting middleware (if needed)

- [ ] Update `cmd/server/main.go`
  - [ ] Initialize KYCService
  - [ ] Initialize KYCHandler
  - [ ] Register KYC routes

**Acceptance Criteria:**
- Routes registered correctly
- Middleware applied
- Service instantiated with all dependencies

**Files:**
- `/internal/handlers/kyc/routes.go`
- `/cmd/server/main.go` (updated)

#### Task 3.3: Integration Testing
**Effort:** 10-12 hours

- [ ] Create `internal/handlers/kyc/aadhaar_handler_test.go`
  - [ ] Test GenerateOTP endpoint
    - [ ] Test with valid request
    - [ ] Test with invalid aadhaar_number
    - [ ] Test with missing consent
    - [ ] Test with rate limit exceeded
    - [ ] Test with Sandbox API error
  - [ ] Test VerifyOTP endpoint
    - [ ] Test with valid OTP
    - [ ] Test with invalid OTP
    - [ ] Test with expired OTP
    - [ ] Test with max attempts exceeded
    - [ ] Test with missing user_id
  - [ ] Test GetKYCStatus endpoint
    - [ ] Test with valid user_id
    - [ ] Test with non-existing user
    - [ ] Test authorization (user vs admin)

- [ ] Create end-to-end test: `test/integration/aadhaar_flow_test.go`
  - [ ] Test complete OTP flow (generate → verify)
  - [ ] Verify user profile updated
  - [ ] Verify address created
  - [ ] Verify photo uploaded to S3

**Acceptance Criteria:**
- All handler tests passing
- Integration test covers complete flow
- Mock services used for testing

**Files:**
- `/internal/handlers/kyc/aadhaar_handler_test.go`
- `/test/integration/aadhaar_flow_test.go`

#### Task 3.4: Add Middleware Enhancements
**Effort:** 4-5 hours

- [ ] Add rate limiting middleware for KYC endpoints (if not already present)
- [ ] Add request validation middleware
- [ ] Add audit logging middleware for KYC operations

**Acceptance Criteria:**
- Rate limiting prevents abuse
- All requests logged

**Files:**
- `/internal/middleware/kyc_middleware.go` (if needed)

---

## Phase 4: gRPC Service & Proto Definitions (Week 4)

**Duration:** 4-5 days
**Effort:** 35-40 hours
**Owner:** @agent-sde-backend-engineer

### Tasks

#### Task 4.1: Create Proto Definitions
**Effort:** 8-10 hours

- [ ] Create `pkg/proto/kyc.proto`
  - [ ] Define KYCService
  - [ ] Define GenerateAadhaarOTP RPC
  - [ ] Define VerifyAadhaarOTP RPC
  - [ ] Define GetKYCStatus RPC
  - [ ] Define UploadAadhaarPhoto RPC (streaming)
  - [ ] Define all request/response messages
  - [ ] Add proper field tags and comments

- [ ] Generate Go code from proto
  - [ ] Run `protoc` or `make proto-gen`
  - [ ] Verify generated files

**Acceptance Criteria:**
- Proto file compiles without errors
- Generated Go code works

**Files:**
- `/pkg/proto/kyc.proto`
- `/pkg/proto/kyc.pb.go` (generated)
- `/pkg/proto/kyc_grpc.pb.go` (generated)

#### Task 4.2: Implement gRPC Service
**Effort:** 15-18 hours

- [ ] Create `internal/grpc_server/kyc_service.go`
  - [ ] Define KYCServiceServer struct
  - [ ] Implement GenerateAadhaarOTP RPC
    - [ ] Convert proto request to service request
    - [ ] Call service.GenerateOTP()
    - [ ] Convert service response to proto response
  - [ ] Implement VerifyAadhaarOTP RPC
    - [ ] Convert proto request to service request
    - [ ] Call service.VerifyOTP()
    - [ ] Convert service response to proto response
  - [ ] Implement GetKYCStatus RPC
    - [ ] Call service.GetKYCStatus()
    - [ ] Convert to proto response
  - [ ] Implement UploadAadhaarPhoto RPC (streaming)
    - [ ] Receive photo metadata
    - [ ] Stream chunks (1MB each)
    - [ ] Buffer in memory (max 5MB)
    - [ ] Upload to S3
    - [ ] Return photo URL

- [ ] Add error mapping (gRPC codes)
  - [ ] Map service errors to gRPC status codes
  - [ ] Return proper error details

**Acceptance Criteria:**
- All RPCs implemented
- Streaming works correctly
- Error handling proper
- gRPC codes match errors

**Files:**
- `/internal/grpc_server/kyc_service.go`

#### Task 4.3: Register gRPC Service
**Effort:** 3-4 hours

- [ ] Update `cmd/server/main.go`
  - [ ] Register KYCService with gRPC server
  - [ ] Add service to reflection (for debugging)

- [ ] Test gRPC service with grpcurl or client
  - [ ] Test GenerateAadhaarOTP
  - [ ] Test VerifyAadhaarOTP
  - [ ] Test GetKYCStatus

**Acceptance Criteria:**
- gRPC service registered
- Service callable via grpcurl

**Files:**
- `/cmd/server/main.go` (updated)

#### Task 4.4: gRPC Integration Tests
**Effort:** 8-10 hours

- [ ] Create `internal/grpc_server/kyc_service_test.go`
  - [ ] Test GenerateAadhaarOTP RPC
  - [ ] Test VerifyAadhaarOTP RPC
  - [ ] Test GetKYCStatus RPC
  - [ ] Test UploadAadhaarPhoto streaming

- [ ] Create gRPC client for testing
  - [ ] Test end-to-end flow via gRPC

**Acceptance Criteria:**
- All gRPC tests passing
- Streaming works correctly

**Files:**
- `/internal/grpc_server/kyc_service_test.go`

---

## Phase 5: Testing, Documentation & Deployment (Week 5)

**Duration:** 5-6 days
**Effort:** 45-50 hours
**Owner:** @agent-sde-backend-engineer + @agent-business-logic-tester

### Tasks

#### Task 5.1: Comprehensive Testing
**Effort:** 15-18 hours

- [ ] Security Testing (@agent-business-logic-tester)
  - [ ] Test OTP rate limiting
  - [ ] Test authorization checks (user can only verify own Aadhaar)
  - [ ] Test Aadhaar masking in logs
  - [ ] Test photo upload security (file type validation)
  - [ ] Test SQL injection prevention
  - [ ] Test XSS prevention in responses

- [ ] Business Logic Validation (@agent-business-logic-tester)
  - [ ] Test state transitions (PENDING → VERIFIED → FAILED)
  - [ ] Test OTP expiration (5 minutes)
  - [ ] Test max attempts enforcement (3 attempts)
  - [ ] Test cooldown period (60 seconds)
  - [ ] Test single-use OTP
  - [ ] Test address auto-population logic
  - [ ] Test name parsing (1, 2, 3+ names)

- [ ] Performance Testing
  - [ ] Load test with 100 concurrent OTP generations
  - [ ] Load test with 100 concurrent OTP verifications
  - [ ] Test database query performance
  - [ ] Test S3 upload performance

**Acceptance Criteria:**
- All security tests passing
- All business logic validated
- Performance targets met
- No vulnerabilities found

**Files:**
- `/test/security/kyc_security_test.go`
- `/test/performance/kyc_load_test.go`

#### Task 5.2: Documentation
**Effort:** 10-12 hours

- [ ] Update Swagger/OpenAPI specification
  - [ ] Document POST /api/v1/kyc/aadhaar/otp
  - [ ] Document POST /api/v1/kyc/aadhaar/otp/verify
  - [ ] Document GET /api/v1/kyc/status/:user_id
  - [ ] Add example requests/responses
  - [ ] Document error codes

- [ ] Create API usage guide
  - [ ] How to generate OTP
  - [ ] How to verify OTP
  - [ ] How to check status
  - [ ] Common error scenarios

- [ ] Update .kiro/specs/aadhaar-integration/
  - [ ] Mark completed tasks
  - [ ] Document any deviations from plan
  - [ ] Add lessons learned

- [ ] Create runbook for operations
  - [ ] How to monitor KYC service
  - [ ] How to troubleshoot failures
  - [ ] How to handle Sandbox API outages
  - [ ] How to rotate API keys

**Acceptance Criteria:**
- Swagger docs complete
- Usage guide clear and tested
- Runbook reviewed by ops team

**Files:**
- `/docs/api/swagger.yaml` (updated)
- `/docs/kyc-usage-guide.md`
- `/docs/runbooks/kyc-operations.md`

#### Task 5.3: Configuration & Environment Setup
**Effort:** 6-8 hours

- [ ] Add environment variables to .env.example
  ```
  AADHAAR_SANDBOX_URL=https://api.sandbox.co.in
  AADHAAR_SANDBOX_API_KEY=your-key
  AADHAAR_SANDBOX_API_SECRET=your-secret
  AWS_S3_BUCKET=aaa-aadhaar-photos
  AWS_S3_REGION=ap-south-1
  AWS_ACCESS_KEY_ID=your-key
  AWS_SECRET_ACCESS_KEY=your-secret
  OTP_EXPIRATION_SECONDS=300
  OTP_MAX_ATTEMPTS=3
  OTP_COOLDOWN_SECONDS=60
  PHOTO_MAX_SIZE_MB=5
  ```

- [ ] Create config struct in `internal/config/kyc_config.go`
  - [ ] Load from environment
  - [ ] Validate required fields
  - [ ] Add defaults

- [ ] Document configuration in README.md

**Acceptance Criteria:**
- All env vars documented
- Config loads correctly
- Validation works

**Files:**
- `.env.example` (updated)
- `/internal/config/kyc_config.go`
- `README.md` (updated)

#### Task 5.4: Deployment Preparation
**Effort:** 8-10 hours

- [ ] Create database migration scripts for production
  - [ ] Verify migrations on staging database
  - [ ] Document rollback procedure

- [ ] Add feature flag support
  - [ ] Add AADHAAR_ENABLED flag (default: false)
  - [ ] Disable KYC endpoints when flag is false
  - [ ] Document flag in configuration

- [ ] Create deployment checklist
  - [ ] Pre-deployment checks
  - [ ] Migration steps
  - [ ] Post-deployment verification
  - [ ] Rollback steps

- [ ] Set up monitoring & alerting
  - [ ] Add Prometheus metrics endpoints
  - [ ] Configure alerts for error rates
  - [ ] Configure alerts for Sandbox API failures
  - [ ] Add health check for KYC service

**Acceptance Criteria:**
- Migrations tested on staging
- Feature flag works
- Monitoring configured
- Deployment checklist reviewed

**Files:**
- `/.kiro/specs/aadhaar-integration/deployment-checklist.md`
- `/internal/monitoring/kyc_metrics.go`

#### Task 5.5: Staging Deployment & UAT
**Effort:** 6-8 hours

- [ ] Deploy to staging environment
  - [ ] Run database migrations
  - [ ] Deploy service with feature flag OFF
  - [ ] Verify health checks
  - [ ] Enable feature flag

- [ ] Conduct User Acceptance Testing (UAT)
  - [ ] Test OTP generation with test Aadhaar numbers
  - [ ] Test OTP verification
  - [ ] Test error scenarios
  - [ ] Test with mobile app/web client

- [ ] Monitor staging for 24 hours
  - [ ] Check error logs
  - [ ] Monitor performance metrics
  - [ ] Verify Sandbox API integration

**Acceptance Criteria:**
- Service deployed successfully
- UAT completed with no critical issues
- No errors in 24-hour monitoring period

---

## Dependencies & Prerequisites

### External Dependencies
- [ ] Sandbox.co.in API credentials obtained
- [ ] AWS S3 bucket created and configured
- [ ] IAM credentials for S3 access
- [ ] Redis cache available (for rate limiting)

### Internal Dependencies
- [ ] UserService available
- [ ] AddressService available
- [ ] AuditService available
- [ ] AuthService available
- [ ] Database migration tooling ready

### Development Environment
- [ ] Go 1.24 installed
- [ ] PostgreSQL 15 available
- [ ] Redis 7 available
- [ ] AWS CLI configured
- [ ] protoc compiler installed

---

## Risk Mitigation

### Risk 1: Sandbox API Downtime
**Mitigation:**
- Implement retry logic with exponential backoff
- Add circuit breaker pattern
- Monitor API availability
- Have fallback manual verification process

### Risk 2: S3 Upload Failures
**Mitigation:**
- Implement retry logic
- Store photo temporarily in database if S3 fails
- Background job to retry failed uploads
- Monitor S3 error rates

### Risk 3: Database Migration Issues
**Mitigation:**
- Test migrations thoroughly on staging
- Have rollback scripts ready
- Backup production database before migration
- Conduct migration during low-traffic period

### Risk 4: Performance Degradation
**Mitigation:**
- Load test before production deployment
- Optimize database queries
- Add caching where appropriate
- Monitor performance metrics

---

## Rollout Strategy

### Phase 1: Database Migration (Day 1)
- Run migrations during maintenance window
- Verify data integrity
- Test rollback procedure

### Phase 2: Service Deployment (Day 1)
- Deploy with feature flag OFF
- Verify health checks
- Test internal endpoints

### Phase 3: Gradual Rollout (Days 2-7)
- Day 2: Enable for 1% of users (internal testing)
- Day 3: Enable for 10% of users
- Day 4: Enable for 25% of users
- Day 5: Enable for 50% of users
- Day 6: Enable for 75% of users
- Day 7: Enable for 100% of users

### Monitoring During Rollout
- Monitor error rates at each rollout stage
- Check performance metrics
- Review user feedback
- Be ready to rollback if issues arise

---

## Success Criteria

### Technical Success
- [ ] All tests passing (unit, integration, e2e)
- [ ] Code coverage > 90%
- [ ] No critical bugs in production
- [ ] Performance targets met
- [ ] Zero security vulnerabilities

### Business Success
- [ ] Users can successfully complete Aadhaar verification
- [ ] OTP success rate > 95%
- [ ] Average verification time < 5 minutes
- [ ] Customer support tickets < 5% of verifications

### Operational Success
- [ ] Service uptime > 99.9%
- [ ] Mean time to recovery < 5 minutes
- [ ] All runbooks documented
- [ ] Team trained on new service

---

## Post-Implementation Tasks

### Week 6: Monitoring & Optimization
- [ ] Review error logs and fix edge cases
- [ ] Optimize slow database queries
- [ ] Fine-tune rate limiting parameters
- [ ] Gather user feedback

### Week 7: Decommission Old Service
- [ ] Migrate remaining users from standalone service
- [ ] Redirect old endpoints to new service
- [ ] Shut down standalone aadhaar-verification service
- [ ] Archive old code

---

## Conclusion

This implementation plan provides a detailed, phased approach to integrating Aadhaar verification into aaa-service. By following this plan, the team can deliver a production-ready KYC system in 4-5 weeks with high quality, comprehensive testing, and proper documentation.
