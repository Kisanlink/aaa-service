# Aadhaar Integration Design Specification

## API Contracts

### REST API Endpoints

#### 1. Generate Aadhaar OTP
**Endpoint:** `POST /api/v1/kyc/aadhaar/otp`

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "aadhaar_number": "123456789012",
  "consent": "Y"
}
```

**Validation:**
- `aadhaar_number`: Required, exactly 12 digits
- `consent`: Required, must be "Y"

**Success Response (200 OK):**
```json
{
  "status_code": 200,
  "message": "OTP sent successfully",
  "data": {
    "reference_id": "123456789",
    "transaction_id": "TXN1234567890",
    "timestamp": 1700000000000,
    "expires_at": 1700000300000
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid aadhaar_number or consent
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Sandbox API failure

#### 2. Verify Aadhaar OTP
**Endpoint:** `POST /api/v1/kyc/aadhaar/otp/verify`

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
user_id: <user_id_from_token>
```

**Request Body:**
```json
{
  "reference_id": "123456789",
  "otp": "123456"
}
```

**Validation:**
- `reference_id`: Required, numeric string
- `otp`: Required, 6 digits

**Success Response (200 OK):**
```json
{
  "status_code": 200,
  "message": "OTP verification successful",
  "data": {
    "verification_status": "VERIFIED",
    "profile_id": "PROF00000123",
    "address_id": "ADDR00000456",
    "aadhaar_data": {
      "name": "John Doe",
      "gender": "M",
      "date_of_birth": "1990-01-01",
      "year_of_birth": 1990,
      "care_of": "S/O Father Name",
      "full_address": "House 123, Street ABC, City XYZ, State, 123456",
      "address": {
        "house": "123",
        "street": "ABC Street",
        "landmark": "Near Park",
        "district": "District Name",
        "state": "State Name",
        "pincode": 123456,
        "country": "India"
      },
      "photo_url": "https://s3.amazonaws.com/bucket/aadhaar/photos/user123.jpg",
      "share_code": "1234",
      "status": "success"
    }
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid OTP or reference_id
- `401 Unauthorized`: Invalid or expired token
- `403 Forbidden`: User can only verify their own Aadhaar
- `404 Not Found`: Reference ID not found
- `429 Too Many Requests`: Max OTP attempts exceeded
- `500 Internal Server Error`: Verification failed

#### 3. Get KYC Status
**Endpoint:** `GET /api/v1/kyc/status/{user_id}`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
{
  "status_code": 200,
  "data": {
    "user_id": "USER00000123",
    "kyc_status": "VERIFIED",
    "aadhaar_verified": true,
    "aadhaar_verified_at": "2024-01-15T10:30:00Z",
    "verification_attempts": 1,
    "last_verification_attempt": "2024-01-15T10:30:00Z"
  }
}
```

### gRPC API (Proto Definition)

**File:** `pkg/proto/kyc.proto`

```protobuf
syntax = "proto3";

package pb;
option go_package = "github.com/Kisanlink/aaa-service/v2/pkg/proto";

import "google/protobuf/timestamp.proto";

// KYC Service for identity verification
service KYCService {
  // Generate OTP for Aadhaar verification
  rpc GenerateAadhaarOTP(GenerateAadhaarOTPRequest) returns (GenerateAadhaarOTPResponse);

  // Verify Aadhaar OTP and update user profile
  rpc VerifyAadhaarOTP(VerifyAadhaarOTPRequest) returns (VerifyAadhaarOTPResponse);

  // Get KYC verification status
  rpc GetKYCStatus(GetKYCStatusRequest) returns (GetKYCStatusResponse);

  // Upload Aadhaar photo (streaming)
  rpc UploadAadhaarPhoto(stream UploadPhotoRequest) returns (UploadPhotoResponse);
}

// Generate OTP Request
message GenerateAadhaarOTPRequest {
  string aadhaar_number = 1; // 12-digit Aadhaar number
  string consent = 2;        // "Y" for yes
}

// Generate OTP Response
message GenerateAadhaarOTPResponse {
  int32 status_code = 1;
  string message = 2;
  string reference_id = 3;
  int64 timestamp = 4;
  string transaction_id = 5;
  int64 expires_at = 6;
}

// Verify OTP Request
message VerifyAadhaarOTPRequest {
  string reference_id = 1;
  string otp = 2;
  string user_id = 3; // User to update after verification
}

// Verify OTP Response
message VerifyAadhaarOTPResponse {
  int32 status_code = 1;
  string message = 2;
  AadhaarData aadhaar_data = 3;
  string profile_id = 4;
  string address_id = 5;
}

// Aadhaar Data
message AadhaarData {
  string name = 1;
  string gender = 2;
  string date_of_birth = 3;
  int32 year_of_birth = 4;
  string care_of = 5;
  string full_address = 6;
  AadhaarAddress address = 7;
  string photo_url = 8;
  string share_code = 9;
  string status = 10;
}

// Aadhaar Address
message AadhaarAddress {
  string house = 1;
  string street = 2;
  string landmark = 3;
  string district = 4;
  string state = 5;
  int32 pincode = 6;
  string country = 7;
}

// Get KYC Status Request
message GetKYCStatusRequest {
  string user_id = 1;
}

// Get KYC Status Response
message GetKYCStatusResponse {
  int32 status_code = 1;
  string user_id = 2;
  string kyc_status = 3;
  bool aadhaar_verified = 4;
  google.protobuf.Timestamp aadhaar_verified_at = 5;
  int32 verification_attempts = 6;
  google.protobuf.Timestamp last_verification_attempt = 7;
}

// Upload Photo Request (streaming)
message UploadPhotoRequest {
  oneof data {
    PhotoMetadata metadata = 1; // First message
    bytes chunk = 2;             // Subsequent chunks (1MB each)
  }
}

// Photo Metadata
message PhotoMetadata {
  string user_id = 1;
  string file_name = 2;
  int64 file_size = 3;
  string content_type = 4;
}

// Upload Photo Response
message UploadPhotoResponse {
  int32 status_code = 1;
  string message = 2;
  string photo_url = 3;
}
```

## Database Schema

### Table: aadhaar_verifications

```sql
CREATE TABLE aadhaar_verifications (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  aadhaar_number VARCHAR(12),
  transaction_id VARCHAR(255) UNIQUE,
  reference_id VARCHAR(255) UNIQUE,
  otp_requested_at TIMESTAMP,
  otp_verified_at TIMESTAMP,
  verification_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, VERIFIED, FAILED
  kyc_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, VERIFIED, REJECTED
  photo_url TEXT,
  name VARCHAR(255),
  date_of_birth DATE,
  gender VARCHAR(20),
  full_address TEXT,
  address_json JSONB, -- Store full address structure
  attempts INT DEFAULT 0,
  last_attempt_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  created_by VARCHAR(255),
  updated_by VARCHAR(255),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_aadhaar_verifications_user_id ON aadhaar_verifications(user_id);
CREATE INDEX idx_aadhaar_verifications_transaction_id ON aadhaar_verifications(transaction_id);
CREATE INDEX idx_aadhaar_verifications_reference_id ON aadhaar_verifications(reference_id);
CREATE INDEX idx_aadhaar_verifications_status ON aadhaar_verifications(verification_status, kyc_status);
```

### Table: otp_attempts

```sql
CREATE TABLE otp_attempts (
  id VARCHAR(255) PRIMARY KEY,
  aadhaar_verification_id VARCHAR(255) NOT NULL,
  attempt_number INT NOT NULL,
  otp_value VARCHAR(6), -- Hashed, not plain text
  ip_address VARCHAR(45),
  user_agent TEXT,
  status VARCHAR(50), -- SUCCESS, FAILED, EXPIRED
  failed_reason VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (aadhaar_verification_id) REFERENCES aadhaar_verifications(id) ON DELETE CASCADE
);

CREATE INDEX idx_otp_attempts_verification_id ON otp_attempts(aadhaar_verification_id);
CREATE INDEX idx_otp_attempts_status ON otp_attempts(status);
```

### Table: user_profiles (enhancements)

```sql
-- Add new columns to existing user_profiles table
ALTER TABLE user_profiles
ADD COLUMN aadhaar_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN aadhaar_verified_at TIMESTAMP,
ADD COLUMN kyc_status VARCHAR(50) DEFAULT 'PENDING';

CREATE INDEX idx_user_profiles_kyc_status ON user_profiles(kyc_status);
CREATE INDEX idx_user_profiles_aadhaar_verified ON user_profiles(aadhaar_verified);
```

## Data Models

### Go Structs

**File:** `internal/entities/models/aadhaar_verification.go`

```go
package models

import (
    "time"
    "database/sql/driver"
    "encoding/json"
)

type AadhaarVerification struct {
    ID                   string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
    UserID               string         `gorm:"type:varchar(255);not null;index" json:"user_id"`
    AadhaarNumber        string         `gorm:"type:varchar(12)" json:"aadhaar_number,omitempty"`
    TransactionID        string         `gorm:"type:varchar(255);unique" json:"transaction_id"`
    ReferenceID          string         `gorm:"type:varchar(255);unique" json:"reference_id"`
    OTPRequestedAt       *time.Time     `json:"otp_requested_at,omitempty"`
    OTPVerifiedAt        *time.Time     `json:"otp_verified_at,omitempty"`
    VerificationStatus   string         `gorm:"type:varchar(50);default:'PENDING'" json:"verification_status"`
    KYCStatus            string         `gorm:"type:varchar(50);default:'PENDING'" json:"kyc_status"`
    PhotoURL             string         `gorm:"type:text" json:"photo_url,omitempty"`
    Name                 string         `gorm:"type:varchar(255)" json:"name,omitempty"`
    DateOfBirth          *time.Time     `json:"date_of_birth,omitempty"`
    Gender               string         `gorm:"type:varchar(20)" json:"gender,omitempty"`
    FullAddress          string         `gorm:"type:text" json:"full_address,omitempty"`
    AddressJSON          AadhaarAddress `gorm:"type:jsonb" json:"address,omitempty"`
    Attempts             int            `gorm:"default:0" json:"attempts"`
    LastAttemptAt        *time.Time     `json:"last_attempt_at,omitempty"`
    CreatedAt            time.Time      `json:"created_at"`
    UpdatedAt            time.Time      `json:"updated_at"`
    DeletedAt            *time.Time     `gorm:"index" json:"deleted_at,omitempty"`
    CreatedBy            string         `gorm:"type:varchar(255)" json:"created_by,omitempty"`
    UpdatedBy            string         `gorm:"type:varchar(255)" json:"updated_by,omitempty"`

    // Relationships
    User        *User        `gorm:"foreignKey:UserID" json:"-"`
    OTPAttempts []OTPAttempt `gorm:"foreignKey:AadhaarVerificationID" json:"otp_attempts,omitempty"`
}

type AadhaarAddress struct {
    House       string `json:"house"`
    Street      string `json:"street"`
    Landmark    string `json:"landmark"`
    District    string `json:"district"`
    State       string `json:"state"`
    Pincode     int    `json:"pincode"`
    Country     string `json:"country"`
}

// Implement JSONB marshaling
func (a AadhaarAddress) Value() (driver.Value, error) {
    return json.Marshal(a)
}

func (a *AadhaarAddress) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return nil
    }
    return json.Unmarshal(bytes, a)
}

func (AadhaarVerification) TableName() string {
    return "aadhaar_verifications"
}
```

**File:** `internal/entities/models/otp_attempt.go`

```go
package models

import "time"

type OTPAttempt struct {
    ID                      string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
    AadhaarVerificationID   string    `gorm:"type:varchar(255);not null;index" json:"aadhaar_verification_id"`
    AttemptNumber           int       `gorm:"not null" json:"attempt_number"`
    OTPValue                string    `gorm:"type:varchar(255)" json:"-"` // Hashed
    IPAddress               string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
    UserAgent               string    `gorm:"type:text" json:"user_agent,omitempty"`
    Status                  string    `gorm:"type:varchar(50)" json:"status"` // SUCCESS, FAILED, EXPIRED
    FailedReason            string    `gorm:"type:varchar(255)" json:"failed_reason,omitempty"`
    CreatedAt               time.Time `json:"created_at"`

    // Relationships
    AadhaarVerification *AadhaarVerification `gorm:"foreignKey:AadhaarVerificationID" json:"-"`
}

func (OTPAttempt) TableName() string {
    return "otp_attempts"
}
```

## Service Layer Design

### KYC Service Interface

**File:** `internal/services/kyc/service.go`

```go
package kyc

import (
    "context"
    "github.com/Kisanlink/aaa-service/v2/internal/entities/models"
    "github.com/Kisanlink/aaa-service/v2/internal/entities/requests/kyc"
    "github.com/Kisanlink/aaa-service/v2/internal/entities/responses/kyc"
)

type Service struct {
    aadhaarRepo      AadhaarVerificationRepository
    userService      UserService
    addressService   AddressService
    sandboxClient    *SandboxClient
    s3Client         *S3Client
    auditService     AuditService
    logger           Logger
    config           *Config
}

type Config struct {
    OTPExpirationSeconds int
    OTPMaxAttempts       int
    OTPCooldownSeconds   int
    PhotoMaxSizeMB       int
}

func NewService(
    aadhaarRepo AadhaarVerificationRepository,
    userService UserService,
    addressService AddressService,
    sandboxClient *SandboxClient,
    s3Client *S3Client,
    auditService AuditService,
    logger Logger,
    config *Config,
) *Service {
    return &Service{
        aadhaarRepo:    aadhaarRepo,
        userService:    userService,
        addressService: addressService,
        sandboxClient:  sandboxClient,
        s3Client:       s3Client,
        auditService:   auditService,
        logger:         logger,
        config:         config,
    }
}

// GenerateOTP generates OTP for Aadhaar verification
func (s *Service) GenerateOTP(ctx context.Context, req *kyc.GenerateOTPRequest) (*kyc.GenerateOTPResponse, error)

// VerifyOTP verifies OTP and updates user profile
func (s *Service) VerifyOTP(ctx context.Context, req *kyc.VerifyOTPRequest) (*kyc.VerifyOTPResponse, error)

// GetKYCStatus retrieves KYC verification status for a user
func (s *Service) GetKYCStatus(ctx context.Context, userID string) (*kyc.KYCStatusResponse, error)

// CheckRateLimit checks if user has exceeded rate limits
func (s *Service) CheckRateLimit(ctx context.Context, aadhaarNumber string) error

// RecordAttempt records an OTP verification attempt
func (s *Service) RecordAttempt(ctx context.Context, verificationID string, status string, reason string) error
```

### Sandbox Client Design

**File:** `internal/services/kyc/sandbox_client.go`

```go
package kyc

import (
    "context"
    "bytes"
    "encoding/json"
    "net/http"
    "time"
)

type SandboxClient struct {
    baseURL   string
    apiKey    string
    apiSecret string
    client    *http.Client
}

func NewSandboxClient(baseURL, apiKey, apiSecret string) *SandboxClient {
    return &SandboxClient{
        baseURL:   baseURL,
        apiKey:    apiKey,
        apiSecret: apiSecret,
        client:    &http.Client{Timeout: 10 * time.Second},
    }
}

func (s *SandboxClient) GenerateOTP(ctx context.Context, aadhaarNumber, consent string) (*SandboxOTPResponse, error)

func (s *SandboxClient) VerifyOTP(ctx context.Context, referenceID, otp string) (*SandboxVerifyResponse, error)

func (s *SandboxClient) setHeaders(req *http.Request) {
    req.Header.Set("accept", "application/json")
    req.Header.Set("x-api-version", "2.0")
    req.Header.Set("content-type", "application/json")
    req.Header.Set("x-api-key", s.apiKey)
}
```

## State Machine

### Verification Status Flow

```
┌─────────┐
│ PENDING │ ← Initial state when OTP requested
└────┬────┘
     │
     ├─── OTP Verified ──────────────────────┐
     │                                       │
     │                                       ▼
     │                               ┌──────────┐
     │                               │ VERIFIED │
     │                               └──────────┘
     │
     ├─── OTP Expired ────────────────────┐
     │                                    │
     ├─── Max Attempts Exceeded ──────────┤
     │                                    │
     └─── API Error ──────────────────────┤
                                          │
                                          ▼
                                    ┌────────┐
                                    │ FAILED │
                                    └────┬───┘
                                         │
                                         │ Retry after cooldown
                                         ▼
                                    ┌─────────┐
                                    │ PENDING │
                                    └─────────┘
```

### State Transition Rules

1. **PENDING → VERIFIED**
   - Condition: Valid OTP verified successfully
   - Actions:
     - Update user profile
     - Create/update address
     - Upload photo to S3
     - Set aadhaar_verified = true
     - Record otp_verified_at timestamp

2. **PENDING → FAILED**
   - Conditions:
     - OTP expired (> 5 minutes)
     - Max attempts exceeded (3 attempts)
     - Sandbox API error
   - Actions:
     - Record failure reason
     - Increment attempt counter
     - Set cooldown period

3. **FAILED → PENDING**
   - Condition: User retries after cooldown period (60 seconds)
   - Actions:
     - Reset attempt counter
     - Generate new reference_id
     - Update last_attempt_at

## Error Handling

### Error Codes

```go
const (
    ErrInvalidAadhaar        = "INVALID_AADHAAR"
    ErrInvalidOTP            = "INVALID_OTP"
    ErrOTPExpired            = "OTP_EXPIRED"
    ErrMaxAttemptsExceeded   = "MAX_ATTEMPTS_EXCEEDED"
    ErrRateLimitExceeded     = "RATE_LIMIT_EXCEEDED"
    ErrSandboxAPIError       = "SANDBOX_API_ERROR"
    ErrPhotoUploadFailed     = "PHOTO_UPLOAD_FAILED"
    ErrUserNotFound          = "USER_NOT_FOUND"
    ErrVerificationNotFound  = "VERIFICATION_NOT_FOUND"
)
```

### Error Response Format

```json
{
  "status_code": 400,
  "error": {
    "code": "INVALID_OTP",
    "message": "The OTP provided is invalid or has expired",
    "details": {
      "attempts_remaining": 2,
      "cooldown_seconds": 60
    }
  }
}
```

## Photo Upload Flow

### Chunked Streaming (gRPC)

```
1. Client sends PhotoMetadata (user_id, file_name, file_size, content_type)
2. Server validates metadata
3. Client streams photo data in 1MB chunks
4. Server buffers chunks in memory (max 5MB)
5. Server validates file format (JPG/PNG)
6. Server compresses if size > 2MB
7. Server uploads to S3 with encryption
8. Server returns photo_url
```

### S3 Upload Strategy

- **Bucket**: `aaa-aadhaar-photos`
- **Key Pattern**: `aadhaar/{user_id}/{timestamp}_{filename}`
- **Encryption**: SSE-S3 (server-side encryption)
- **Access**: Private (signed URLs for retrieval)
- **Lifecycle**: Archive after 7 years (compliance)

## Performance Targets

| Operation | Target | Max |
|-----------|--------|-----|
| OTP Generation | < 1.5s | 3s |
| OTP Verification | < 2s | 5s |
| Photo Upload (5MB) | < 3s | 10s |
| Database Query | < 50ms | 100ms |
| Cache Lookup | < 10ms | 50ms |

## Security Considerations

### Data Masking

Aadhaar numbers must be masked in all logs and responses:
- Display format: `XXXX-XXXX-1234` (show only last 4 digits)
- Never log full Aadhaar number

### OTP Security

- OTP never stored in plain text
- Reference ID used instead of OTP for tracking
- OTP valid for 5 minutes only
- Single-use OTP (cannot reuse)

### Photo Security

- Photos encrypted at rest (S3 SSE)
- Access via signed URLs only
- Expire signed URLs after 1 hour
- Scan for malware before storage

## Conclusion

This design specification provides detailed API contracts, database schema, data models, service interfaces, state machines, and error handling for the Aadhaar integration. All components follow established patterns in the aaa-service codebase.
