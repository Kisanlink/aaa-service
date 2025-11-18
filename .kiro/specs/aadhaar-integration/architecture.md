# Aadhaar Integration Architecture

## Overview

This document describes the architecture for integrating Aadhaar verification functionality directly into the aaa-service, consolidating from a standalone microservice approach.

## Architectural Decisions

### Decision 1: API Strategy - REST + gRPC ✅
**Decision:** Support BOTH REST and gRPC endpoints

**Rationale:**
- REST for web/mobile clients: `POST /api/v1/kyc/aadhaar/otp`, `POST /api/v1/kyc/aadhaar/otp/verify`
- gRPC for internal service-to-service calls: `KYCService.GenerateAadhaarOTP()`, `KYCService.VerifyAadhaarOTP()`
- Maintains consistency with existing aaa-service API patterns
- Provides flexibility for different client types

**Implementation:**
- REST handlers in `internal/handlers/kyc/`
- gRPC service implementation in `internal/grpc_server/`
- Shared service layer in `internal/services/kyc/`

### Decision 2: Photo Storage - Best Practices for Binary Streaming ✅
**Decision:** Use chunked streaming with gRPC best practices

**Strategy:**
- Photo upload via multipart form-data in REST
- Chunked streaming in gRPC (1MB chunks)
- Store final photo in AWS S3
- Return S3 URL in response

**Implementation Details:**
```protobuf
message UploadPhotoRequest {
  oneof data {
    PhotoMetadata metadata = 1;  // First message
    bytes chunk = 2;              // Subsequent chunks
  }
}

message PhotoMetadata {
  string user_id = 1;
  string file_name = 2;
  int64 file_size = 3;
  string content_type = 4;
}
```

**Best Practices:**
- Maximum photo size: 5MB
- Supported formats: JPG, PNG
- Chunk size: 1MB (configurable)
- Timeout: 30 seconds per upload
- Automatic compression if size > 2MB

### Decision 3: Address Auto-Update ✅
**Decision:** Auto-update user address with Aadhaar data

**Rationale:**
- Reduces friction in user onboarding
- Aadhaar address is government-verified
- User can update later if needed

**Implementation:**
- After OTP verification, automatically create/update address
- Use AddressService to populate address fields
- Mark address source as "aadhaar_okyc" in metadata
- Set as primary address if no existing primary

**Workflow:**
```
Aadhaar Verification Success
  → Extract address data from Aadhaar response
  → Call AddressService.CreateOrUpdate()
  → Set is_primary = true if no existing primary
  → Store in metadata: {"source": "aadhaar_okyc", "full_address": "..."}
```

### Decision 4: Simple 3-State Machine ✅
**Decision:** Use simple 3-state verification workflow

**States:**
1. **PENDING** - User initiated Aadhaar verification, OTP requested
2. **VERIFIED** - OTP successfully verified, profile updated
3. **FAILED** - OTP verification failed or expired

**State Transitions:**
```
PENDING → VERIFIED  (on successful OTP verification)
PENDING → FAILED    (on OTP expiration or max attempts exceeded)
FAILED → PENDING    (user can retry after cooldown)
```

**Implementation:**
- Store state in `aadhaar_verifications.verification_status` column
- Track attempts in `otp_attempts` table
- Max attempts: 3 (configurable via `OTP_MAX_ATTEMPTS`)
- Cooldown period: 60 seconds between retries

### Decision 5: Sandbox API Integration - Simple Implementation ✅
**Decision:** Straightforward client implementation with API key from environment

**Approach:**
- No additional complexity layers
- Direct HTTP client to Sandbox.co.in API
- API credentials from environment variables
- Standard error handling and retry logic

**Implementation:**
```go
type SandboxClient struct {
    baseURL   string
    apiKey    string
    apiSecret string
    client    *http.Client
}

// Initialize from env
sandboxClient := NewSandboxClient(
    os.Getenv("AADHAAR_SANDBOX_URL"),
    os.Getenv("AADHAAR_SANDBOX_API_KEY"),
    os.Getenv("AADHAAR_SANDBOX_API_SECRET"),
)
```

**Error Handling:**
- Network errors: Retry up to 3 times with exponential backoff
- API errors: Log and return structured error to client
- Timeout: 10 seconds per request

## System Architecture

### Current Architecture (Before)
```
┌─────────────┐         ┌──────────────────────┐         ┌─────────────┐
│   Client    │────────>│ aadhaar-verification │────────>│ Sandbox API │
│             │         │    (Go/Gin)          │         │             │
└─────────────┘         └──────────────────────┘         └─────────────┘
                                 │
                                 │ gRPC calls
                                 ▼
                        ┌──────────────┐
                        │ aaa-service  │
                        │   (gRPC)     │
                        └──────────────┘
                                 │
                                 ▼
                        ┌──────────────┐
                        │  PostgreSQL  │
                        └──────────────┘
```

### Proposed Architecture (After)
```
┌─────────────┐         ┌─────────────────────┐         ┌─────────────┐
│   Client    │────────>│    aaa-service      │────────>│ Sandbox API │
│ (REST/gRPC) │         │  (REST + gRPC)      │         │             │
└─────────────┘         └─────────────────────┘         └─────────────┘
                                 │                                │
                                 │                                │
                                 ▼                                ▼
                        ┌──────────────┐                 ┌──────────┐
                        │  PostgreSQL  │                 │  AWS S3  │
                        │              │                 │ (Photos) │
                        └──────────────┘                 └──────────┘
```

## Component Diagram

```
aaa-service
├── API Layer
│   ├── REST Handlers (Gin)
│   │   ├── POST /api/v1/kyc/aadhaar/otp
│   │   ├── POST /api/v1/kyc/aadhaar/otp/verify
│   │   └── GET /api/v1/kyc/status/{user_id}
│   └── gRPC Service
│       ├── KYCService.GenerateAadhaarOTP()
│       ├── KYCService.VerifyAadhaarOTP()
│       └── KYCService.GetKYCStatus()
│
├── Service Layer
│   └── KYCService
│       ├── OTP Generation Logic
│       ├── OTP Verification Logic
│       ├── State Management
│       └── Integration with UserService, AddressService
│
├── Repository Layer
│   └── AadhaarVerificationRepository
│       ├── CreateVerification()
│       ├── UpdateVerificationStatus()
│       ├── GetVerificationByUserID()
│       └── RecordOTPAttempt()
│
├── External Integration Layer
│   ├── SandboxClient (Aadhaar API)
│   └── S3Client (Photo Storage)
│
└── Data Layer
    ├── aadhaar_verifications table
    ├── otp_attempts table
    └── user_profiles table (enhanced)
```

## Data Flow

### OTP Generation Flow
```
1. Client → REST/gRPC Request (aadhaar_number, consent)
2. KYCHandler → Validate request
3. KYCHandler → KYCService.GenerateOTP()
4. KYCService → Check rate limits
5. KYCService → SandboxClient.GenerateOTP()
6. SandboxClient → Sandbox API (HTTPS)
7. Sandbox API → Return OTP reference_id
8. KYCService → Create AadhaarVerification record (status=PENDING)
9. KYCService → Return response with reference_id
10. Client → Receive reference_id for verification
```

### OTP Verification Flow
```
1. Client → REST/gRPC Request (reference_id, otp, user_id)
2. KYCHandler → Validate request
3. KYCHandler → KYCService.VerifyOTP()
4. KYCService → Check attempt limits
5. KYCService → SandboxClient.VerifyOTP()
6. SandboxClient → Sandbox API (HTTPS)
7. Sandbox API → Return KYC data (name, dob, address, photo)
8. KYCService → Upload photo to S3
9. S3Client → Store photo, return URL
10. KYCService → Update user profile (name, dob, gender)
11. KYCService → Create/update address
12. KYCService → Update verification status = VERIFIED
13. KYCService → Return success response
14. Client → Receive success with profile/address IDs
```

## Security Architecture

### Authentication & Authorization
- All endpoints require valid JWT token
- User can only verify their own Aadhaar
- Admin role can view verification status for any user

### Data Protection
- Aadhaar number stored as VARCHAR(12) (encrypted at application level if needed)
- Photos encrypted at rest in S3 (SSE-S3)
- OTP never stored (only reference_id from Sandbox API)
- Sensitive data masked in logs

### Rate Limiting
- OTP generation: 3 attempts per 60 seconds per Aadhaar number
- OTP verification: 3 attempts per reference_id
- Cooldown period: 60 seconds after 3 failed attempts

### Audit Logging
- All OTP generation requests logged
- All verification attempts logged (success/failure)
- Photo uploads logged
- User profile updates logged

## Scalability Considerations

### Performance Targets
- OTP generation: < 2 seconds (including Sandbox API call)
- OTP verification: < 3 seconds (including Sandbox API + S3 upload)
- Database queries: < 100ms
- Photo upload: < 5 seconds for 5MB file

### Caching Strategy
- Cache verification status in Redis (TTL: 5 minutes)
- Cache user profile after update (TTL: 10 minutes)
- No caching of OTP data (security concern)

### Database Optimization
- Index on `aadhaar_verifications.user_id`
- Index on `aadhaar_verifications.transaction_id`
- Index on `otp_attempts.aadhaar_verification_id`
- Partition `otp_attempts` table by created_at (monthly)

## Integration Points

### Internal Services
- **UserService**: Update user validation status
- **AddressService**: Create/update user address
- **AuditService**: Log all KYC operations

### External Services
- **Sandbox.co.in API**: Aadhaar OTP generation/verification
- **AWS S3**: Photo storage
- **Redis**: Caching and rate limiting

## Deployment Strategy

### Phase 1: Database Setup
- Run migrations for new tables
- Add indexes
- Verify rollback procedures

### Phase 2: Service Deployment
- Deploy with feature flag `AADHAAR_ENABLED=false`
- Verify health checks
- Test internal endpoints

### Phase 3: Gradual Rollout
- Enable for 10% of users
- Monitor error rates and performance
- Gradually increase to 100%

### Phase 4: Decommission Old Service
- Migrate remaining users
- Shut down standalone aadhaar-verification service
- Archive old code

## Rollback Plan

### Immediate Rollback (< 5 minutes)
- Set feature flag `AADHAAR_ENABLED=false`
- Route traffic to old service (if still running)

### Full Rollback (< 30 minutes)
- Revert database migrations
- Redeploy previous version
- Update client configurations

## Monitoring & Alerting

### Key Metrics
- `kyc_otp_generation_total` (counter)
- `kyc_otp_generation_errors_total` (counter)
- `kyc_otp_verification_total` (counter)
- `kyc_otp_verification_errors_total` (counter)
- `kyc_photo_upload_duration` (histogram)
- `kyc_sandbox_api_duration` (histogram)

### Alerts
- Error rate > 5% for 5 minutes
- Sandbox API latency > 5 seconds
- Photo upload failures > 10% for 5 minutes
- Database query latency > 200ms

## Technology Stack

### Backend
- **Language**: Go 1.24
- **Framework**: Gin (REST), gRPC
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Storage**: AWS S3

### External APIs
- **Aadhaar Verification**: Sandbox.co.in API v2.0

### Infrastructure
- **Deployment**: Docker containers
- **Orchestration**: Kubernetes (if applicable)
- **CI/CD**: GitHub Actions

## Conclusion

This architecture consolidates Aadhaar verification into aaa-service, eliminating the need for a separate microservice while maintaining all functionality. The design follows established patterns in the codebase and provides clear integration points with existing services.
