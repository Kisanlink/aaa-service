# Aadhaar Integration Requirements

## Functional Requirements

### FR-1: Aadhaar OTP Generation
**Priority:** HIGH
**User Story:** As a user, I want to generate an OTP for Aadhaar verification so that I can verify my identity.

**Acceptance Criteria:**
- [ ] System SHALL accept 12-digit Aadhaar number
- [ ] System SHALL require explicit user consent ("Y")
- [ ] System SHALL call Sandbox.co.in API to generate OTP
- [ ] System SHALL return reference_id for OTP verification
- [ ] System SHALL create aadhaar_verification record with status PENDING
- [ ] System SHALL enforce rate limiting (3 attempts per 60 seconds)
- [ ] System SHALL return appropriate error messages for failures

**Business Rules:**
- OTP generation requires valid JWT token
- User can only verify their own Aadhaar
- Aadhaar number must be unique per user
- OTP expires after 5 minutes

### FR-2: Aadhaar OTP Verification
**Priority:** HIGH
**User Story:** As a user, I want to verify the OTP sent to my Aadhaar-linked mobile so that I can complete KYC verification.

**Acceptance Criteria:**
- [ ] System SHALL accept reference_id and OTP
- [ ] System SHALL validate OTP with Sandbox API
- [ ] System SHALL retrieve KYC data (name, DOB, gender, address, photo)
- [ ] System SHALL upload photo to AWS S3
- [ ] System SHALL update user profile with KYC data
- [ ] System SHALL create/update user address
- [ ] System SHALL update verification status to VERIFIED
- [ ] System SHALL return profile_id and address_id in response
- [ ] System SHALL enforce max 3 OTP attempts per reference_id

**Business Rules:**
- OTP must be verified within 5 minutes of generation
- Failed OTP attempts are logged
- After 3 failed attempts, user must wait 60 seconds
- User status changes to "active" after successful verification

### FR-3: KYC Status Retrieval
**Priority:** MEDIUM
**User Story:** As a user/admin, I want to check KYC verification status so that I can track verification progress.

**Acceptance Criteria:**
- [ ] System SHALL return current KYC status (PENDING, VERIFIED, FAILED)
- [ ] System SHALL return aadhaar_verified flag
- [ ] System SHALL return aadhaar_verified_at timestamp
- [ ] System SHALL return verification attempt count
- [ ] System SHALL return last_verification_attempt timestamp
- [ ] Admin users can check status for any user
- [ ] Regular users can only check their own status

### FR-4: Photo Upload and Storage
**Priority:** HIGH
**User Story:** As the system, I need to store Aadhaar photos securely so that user identity can be visually verified.

**Acceptance Criteria:**
- [ ] System SHALL accept base64-encoded photo from Sandbox API
- [ ] System SHALL validate photo format (JPG, PNG only)
- [ ] System SHALL validate photo size (max 5MB)
- [ ] System SHALL compress photos larger than 2MB
- [ ] System SHALL upload photo to AWS S3
- [ ] System SHALL encrypt photos at rest (SSE-S3)
- [ ] System SHALL return publicly accessible S3 URL
- [ ] System SHALL update user profile with photo_url

**Business Rules:**
- Photos stored in S3 bucket: `aaa-aadhaar-photos`
- Key pattern: `aadhaar/{user_id}/{timestamp}_{filename}`
- Photos retained for 7 years (compliance requirement)
- Access via signed URLs only

### FR-5: Address Auto-Population
**Priority:** MEDIUM
**User Story:** As a user, I want my address automatically populated from Aadhaar data so that I don't have to enter it manually.

**Acceptance Criteria:**
- [ ] System SHALL extract address from Aadhaar verification response
- [ ] System SHALL parse address components (house, street, district, state, pincode)
- [ ] System SHALL create new address record if none exists
- [ ] System SHALL update existing address if already present
- [ ] System SHALL mark address as primary if no existing primary
- [ ] System SHALL store original Aadhaar address in metadata
- [ ] System SHALL allow user to edit address after verification

**Business Rules:**
- Address source marked as "aadhaar_okyc" in metadata
- Address type set to "HOME"
- User can have multiple addresses

### FR-6: User Profile Update
**Priority:** HIGH
**User Story:** As the system, I need to update user profile with verified Aadhaar data so that user information is accurate.

**Acceptance Criteria:**
- [ ] System SHALL update user full_name with Aadhaar name
- [ ] System SHALL update profile first_name, middle_name, last_name (parsed)
- [ ] System SHALL update profile date_of_birth
- [ ] System SHALL update profile gender
- [ ] System SHALL update profile avatar_url with photo
- [ ] System SHALL set aadhaar_verified = true
- [ ] System SHALL set aadhaar_verified_at timestamp
- [ ] System SHALL update kyc_status = VERIFIED
- [ ] System SHALL set user is_validated = true
- [ ] System SHALL set user status = "active"

**Business Rules:**
- Name parsing: First, Middle (if any), Last
- Single name: First name only
- Two names: First and Last
- Three+ names: First, Middle (combined), Last

### FR-7: Rate Limiting
**Priority:** HIGH
**User Story:** As a system administrator, I want to prevent OTP generation abuse so that the system remains secure.

**Acceptance Criteria:**
- [ ] System SHALL limit OTP generation to 3 attempts per 60 seconds per Aadhaar number
- [ ] System SHALL limit OTP verification to 3 attempts per reference_id
- [ ] System SHALL enforce 60-second cooldown after max attempts
- [ ] System SHALL return 429 status code when rate limit exceeded
- [ ] System SHALL include cooldown duration in error response
- [ ] System SHALL track attempts in database

**Business Rules:**
- Rate limits apply per Aadhaar number, not per user
- Cooldown resets after 60 seconds
- Successful verification resets attempt counter

### FR-8: Audit Logging
**Priority:** HIGH
**User Story:** As a compliance officer, I need detailed audit logs of KYC operations so that I can ensure regulatory compliance.

**Acceptance Criteria:**
- [ ] System SHALL log all OTP generation requests
- [ ] System SHALL log all OTP verification attempts (success/failure)
- [ ] System SHALL log photo uploads
- [ ] System SHALL log profile updates
- [ ] System SHALL log address updates
- [ ] System SHALL include user_id, IP address, timestamp in logs
- [ ] System SHALL log error details for failures
- [ ] Audit logs SHALL be immutable

**Business Rules:**
- Logs retained for 7 years (compliance)
- Sensitive data (Aadhaar, OTP) masked in logs
- Logs stored in audit_logs table

## Non-Functional Requirements

### NFR-1: Performance
**Priority:** HIGH

**Requirements:**
- OTP generation SHALL complete in < 2 seconds (including Sandbox API call)
- OTP verification SHALL complete in < 3 seconds (including Sandbox API + S3 upload)
- Database queries SHALL complete in < 100ms
- Photo upload (5MB) SHALL complete in < 5 seconds
- API response time (95th percentile) SHALL be < 500ms

**Measurement:**
- Monitor API latency using Prometheus metrics
- Set alerts for latency > 1 second

### NFR-2: Scalability
**Priority:** MEDIUM

**Requirements:**
- System SHALL handle 100 concurrent OTP generations
- System SHALL handle 100 concurrent OTP verifications
- System SHALL handle 50 concurrent photo uploads
- Database SHALL support 1 million verification records
- S3 storage SHALL support 1 million photos

**Capacity Planning:**
- Expected load: 10,000 verifications per day
- Peak load: 100 verifications per minute
- Storage growth: 5GB per month

### NFR-3: Security
**Priority:** CRITICAL

**Requirements:**
- All API endpoints SHALL require valid JWT authentication
- Aadhaar numbers SHALL be masked in logs and responses (show only last 4 digits)
- Photos SHALL be encrypted at rest in S3 (SSE-S3)
- OTP SHALL never be stored in plain text
- Communication with Sandbox API SHALL use HTTPS
- API keys SHALL be stored in environment variables, not code
- Rate limiting SHALL prevent brute force attacks
- Failed verification attempts SHALL be logged

**Compliance:**
- Aadhaar Act 2016 compliance
- Data Protection and Privacy Act (DPDPA) compliance
- PCI DSS compliance for photo storage

### NFR-4: Reliability
**Priority:** HIGH

**Requirements:**
- System uptime SHALL be 99.9% (43 minutes downtime per month)
- Database backup SHALL be performed daily
- S3 photos SHALL have 99.999999999% durability
- Sandbox API failures SHALL not crash the service
- Failed operations SHALL be retryable

**Error Handling:**
- Sandbox API errors: Retry up to 3 times with exponential backoff
- S3 upload errors: Retry up to 3 times
- Database errors: Log and return 500 error
- Network timeouts: 10 seconds for Sandbox API, 30 seconds for S3

### NFR-5: Maintainability
**Priority:** MEDIUM

**Requirements:**
- Code SHALL follow Go best practices
- Code coverage SHALL be > 90% for service logic
- All functions SHALL have clear documentation
- Error messages SHALL be descriptive
- Configuration SHALL be externalized (environment variables)
- Database migrations SHALL be versioned and reversible

**Code Standards:**
- Maximum 300 lines per file
- snake_case for file names
- PascalCase for exported names
- camelCase for private names
- No hardcoded credentials

### NFR-6: Observability
**Priority:** HIGH

**Requirements:**
- System SHALL expose Prometheus metrics
- System SHALL log all operations with structured logging
- System SHALL report health status via /health endpoint
- System SHALL track success/failure rates
- System SHALL monitor Sandbox API latency

**Metrics:**
- `kyc_otp_generation_total` (counter)
- `kyc_otp_generation_errors_total` (counter)
- `kyc_otp_verification_total` (counter)
- `kyc_otp_verification_errors_total` (counter)
- `kyc_photo_upload_duration` (histogram)
- `kyc_sandbox_api_duration` (histogram)

**Alerts:**
- Error rate > 5% for 5 minutes
- Sandbox API latency > 5 seconds
- Photo upload failures > 10%

### NFR-7: Testability
**Priority:** HIGH

**Requirements:**
- All services SHALL be mockable via interfaces
- Unit tests SHALL cover edge cases
- Integration tests SHALL use test database
- Sandbox API SHALL be mockable for testing
- End-to-end tests SHALL cover full OTP flow

**Test Coverage:**
- Unit tests: > 90% coverage
- Integration tests: All API endpoints
- Security tests: All authorization checks
- Performance tests: Load testing with 100 concurrent users

## Data Requirements

### DR-1: Data Retention
**Priority:** HIGH

**Requirements:**
- Aadhaar verification records SHALL be retained for 7 years
- OTP attempt records SHALL be retained for 1 year
- Photos SHALL be retained for 7 years
- Deleted records SHALL be soft-deleted (deleted_at timestamp)

### DR-2: Data Privacy
**Priority:** CRITICAL

**Requirements:**
- Aadhaar numbers SHALL be masked in all logs and API responses
- Full Aadhaar SHALL only be visible to admin users
- Photos SHALL not be publicly accessible (signed URLs only)
- User consent SHALL be recorded for Aadhaar verification
- Users SHALL be able to request data deletion (GDPR right to erasure)

### DR-3: Data Accuracy
**Priority:** HIGH

**Requirements:**
- Aadhaar data SHALL come directly from government-verified source (Sandbox API)
- No manual editing of Aadhaar-verified data
- Profile updates SHALL be atomic (all-or-nothing)
- Address updates SHALL preserve original Aadhaar address in metadata

## Integration Requirements

### IR-1: Sandbox.co.in API Integration
**Priority:** CRITICAL

**Requirements:**
- System SHALL integrate with Sandbox.co.in API v2.0
- System SHALL use API key and secret for authentication
- System SHALL handle API rate limits gracefully
- System SHALL retry failed API calls (max 3 retries)
- System SHALL log all API interactions
- System SHALL handle API downtime gracefully

**API Endpoints:**
- POST `/kyc/aadhaar/okyc/otp` - Generate OTP
- POST `/kyc/aadhaar/okyc/otp/verify` - Verify OTP

### IR-2: AWS S3 Integration
**Priority:** HIGH

**Requirements:**
- System SHALL upload photos to S3 bucket
- System SHALL use IAM credentials for S3 access
- System SHALL enable server-side encryption (SSE-S3)
- System SHALL generate signed URLs for photo access
- System SHALL handle S3 upload failures gracefully

**S3 Configuration:**
- Bucket: `aaa-aadhaar-photos`
- Region: `ap-south-1`
- Lifecycle: Archive to Glacier after 7 years

### IR-3: Internal Services Integration
**Priority:** HIGH

**Requirements:**
- System SHALL integrate with UserService for profile updates
- System SHALL integrate with AddressService for address creation
- System SHALL integrate with AuditService for logging
- System SHALL use existing authentication middleware
- System SHALL use existing authorization checks

## Compliance Requirements

### CR-1: Aadhaar Act 2016
**Priority:** CRITICAL

**Requirements:**
- System SHALL obtain explicit user consent before Aadhaar verification
- System SHALL not store Aadhaar number without consent
- System SHALL use Aadhaar only for identity verification
- System SHALL not share Aadhaar data with third parties
- System SHALL comply with UIDAI guidelines

### CR-2: Data Protection and Privacy Act (DPDPA)
**Priority:** CRITICAL

**Requirements:**
- System SHALL implement data minimization (collect only necessary data)
- System SHALL provide data access to users (right to access)
- System SHALL allow data deletion (right to erasure)
- System SHALL secure data at rest and in transit
- System SHALL notify users of data breaches

### CR-3: PCI DSS (Photo Storage)
**Priority:** HIGH

**Requirements:**
- Photos SHALL be encrypted at rest
- Access to photos SHALL be logged
- Photos SHALL not be stored on insecure systems
- Photo access SHALL require authentication

## Dependencies

### External Dependencies
- Sandbox.co.in API availability
- AWS S3 service availability
- PostgreSQL database
- Redis cache

### Internal Dependencies
- UserService (profile updates)
- AddressService (address creation)
- AuditService (logging)
- AuthService (authentication)

## Constraints

### Technical Constraints
- Sandbox API rate limits: 100 requests per minute
- S3 upload limit: 5GB per file (we limit to 5MB)
- Database connection pool: 100 connections
- Maximum concurrent photo uploads: 50

### Business Constraints
- OTP valid for 5 minutes only
- Maximum 3 OTP attempts per reference_id
- Cooldown period: 60 seconds after failed attempts
- Photo retention: 7 years minimum

## Success Criteria

### SC-1: Functional Success
- [ ] Users can successfully generate OTP
- [ ] Users can successfully verify OTP
- [ ] Profile updated with Aadhaar data after verification
- [ ] Address created/updated after verification
- [ ] Photo uploaded to S3 successfully
- [ ] KYC status updated to VERIFIED

### SC-2: Performance Success
- [ ] OTP generation < 2 seconds (95th percentile)
- [ ] OTP verification < 3 seconds (95th percentile)
- [ ] Photo upload < 5 seconds (5MB file)
- [ ] Database queries < 100ms
- [ ] Zero downtime deployment

### SC-3: Security Success
- [ ] No Aadhaar numbers leaked in logs
- [ ] All photos encrypted at rest
- [ ] Rate limiting prevents abuse
- [ ] Authorization checks pass security audit
- [ ] Zero security vulnerabilities in penetration testing

### SC-4: Quality Success
- [ ] Code coverage > 90%
- [ ] All tests passing
- [ ] No critical bugs in production
- [ ] API documentation complete
- [ ] Monitoring and alerting active

## Conclusion

This requirements document defines all functional, non-functional, data, integration, and compliance requirements for Aadhaar verification integration in aaa-service. All requirements follow the established patterns and standards of the aaa-service codebase.
