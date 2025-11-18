# Aadhaar Verification E2E Testing Guide

## Overview
This guide provides comprehensive instructions for testing the Aadhaar verification APIs end-to-end, including OTP generation, verification, and KYC status retrieval.

## Prerequisites

### Required Environment Variables
```bash
# Sandbox API Configuration
export AADHAAR_SANDBOX_URL=https://api.sandbox.co.in
export AADHAAR_SANDBOX_API_KEY=your-sandbox-api-key
export AADHAAR_SANDBOX_API_SECRET=your-sandbox-api-secret

# OTP Configuration
export OTP_EXPIRATION_SECONDS=300
export OTP_MAX_ATTEMPTS=3
export OTP_COOLDOWN_SECONDS=60
export PHOTO_MAX_SIZE_MB=5

# S3 Configuration
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export S3_BUCKET_NAME=your-bucket-name
export S3_REGION=ap-south-1

# Database
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=aaa_service
export DB_USER=postgres
export DB_PASSWORD=postgres
```

### Test User Setup
1. Create a test user in the database
2. Note the user ID for testing
3. Ensure the user doesn't have existing Aadhaar verification

## Test Scenarios

### Scenario 1: Happy Path - Complete Verification Flow

#### Step 1: Generate OTP
```bash
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "test-user-123",
    "aadhaar_number": "123456789012",
    "consent": "Y"
  }'
```

**Expected Response:**
```json
{
  "status_code": 200,
  "message": "OTP sent successfully to Aadhaar-linked mobile number",
  "reference_id": "REF123456789",
  "transaction_id": "TXN1234567890",
  "timestamp": 1234567890,
  "expires_at": 1234568190
}
```

**Validations:**
- ✓ Status code is 200
- ✓ Reference ID is returned
- ✓ Transaction ID is returned
- ✓ Expiry time is 5 minutes (300 seconds) from now
- ✓ OTP is sent to Aadhaar-linked mobile number (check Sandbox logs)

#### Step 2: Verify OTP
```bash
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "test-user-123",
    "reference_id": "REF123456789",
    "otp": "123456"
  }'
```

**Expected Response:**
```json
{
  "status_code": 200,
  "message": "Aadhaar verification successful",
  "aadhaar_data": {
    "name": "JOHN DOE",
    "gender": "M",
    "date_of_birth": "1990-01-01",
    "year_of_birth": 1990,
    "care_of": "S/O Father Name",
    "full_address": "House 123, Street Name, City, State - 123456",
    "address": {
      "house": "123",
      "street": "Street Name",
      "landmark": "Near Park",
      "district": "District Name",
      "state": "State Name",
      "pincode": 123456,
      "country": "India"
    },
    "photo_url": "https://s3.amazonaws.com/bucket/aadhaar/photos/test-user-123/aadhaar_1234567890.jpg",
    "share_code": "1234",
    "status": "success"
  },
  "profile_id": "profile-123",
  "address_id": "address-456"
}
```

**Validations:**
- ✓ Status code is 200
- ✓ Aadhaar data is returned with all fields
- ✓ Photo URL is accessible
- ✓ User profile is updated with Aadhaar data
- ✓ Address record is created
- ✓ Verification status is updated to VERIFIED
- ✓ KYC status is updated to APPROVED

#### Step 3: Check KYC Status
```bash
curl -X GET http://localhost:8080/api/v1/kyc/status/test-user-123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-User-ID: test-user-123"
```

**Expected Response:**
```json
{
  "status_code": 200,
  "message": "KYC status retrieved successfully",
  "kyc_status": {
    "user_id": "test-user-123",
    "verification_status": "VERIFIED",
    "kyc_status": "APPROVED",
    "name": "JOHN DOE",
    "gender": "M",
    "date_of_birth": "1990-01-01T00:00:00Z",
    "full_address": "House 123, Street Name, City, State - 123456",
    "photo_url": "https://s3.amazonaws.com/bucket/aadhaar/photos/test-user-123/aadhaar_1234567890.jpg",
    "verified_at": "2025-11-18T12:34:56Z"
  }
}
```

**Validations:**
- ✓ Status code is 200
- ✓ Verification status is VERIFIED
- ✓ KYC status is APPROVED
- ✓ All Aadhaar data is present

### Scenario 2: Invalid Aadhaar Number

```bash
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "test-user-123",
    "aadhaar_number": "invalid",
    "consent": "Y"
  }'
```

**Expected Response:**
```json
{
  "status_code": 400,
  "message": "Aadhaar number must be exactly 12 digits",
  "errors": [
    "invalid aadhaar number format"
  ]
}
```

**Validations:**
- ✓ Status code is 400
- ✓ Validation error is returned
- ✓ No OTP is generated

### Scenario 3: Missing Consent

```bash
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "test-user-123",
    "aadhaar_number": "123456789012",
    "consent": "N"
  }'
```

**Expected Response:**
```json
{
  "status_code": 400,
  "message": "User consent is required for Aadhaar verification",
  "errors": [
    "consent must be 'Y'"
  ]
}
```

**Validations:**
- ✓ Status code is 400
- ✓ Consent error is returned

### Scenario 4: Invalid OTP

```bash
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "test-user-123",
    "reference_id": "REF123456789",
    "otp": "wrong-otp"
  }'
```

**Expected Response:**
```json
{
  "status_code": 400,
  "message": "Invalid OTP provided",
  "errors": [
    "OTP verification failed"
  ]
}
```

**Validations:**
- ✓ Status code is 400
- ✓ OTP attempt is recorded
- ✓ Verification status remains PENDING
- ✓ User can retry (if under max attempts)

### Scenario 5: Expired OTP

**Prerequisites:**
- Generate OTP
- Wait for 5 minutes (300 seconds)

```bash
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "test-user-123",
    "reference_id": "REF123456789",
    "otp": "123456"
  }'
```

**Expected Response:**
```json
{
  "status_code": 400,
  "message": "OTP has expired. Please request a new OTP",
  "errors": [
    "OTP expired"
  ]
}
```

**Validations:**
- ✓ Status code is 400
- ✓ Expiry error is returned
- ✓ User must generate new OTP

### Scenario 6: Max OTP Attempts Exceeded

**Prerequisites:**
- Generate OTP
- Attempt verification 3 times with wrong OTP

```bash
# Fourth attempt
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "test-user-123",
    "reference_id": "REF123456789",
    "otp": "123456"
  }'
```

**Expected Response:**
```json
{
  "status_code": 429,
  "message": "Maximum OTP verification attempts exceeded",
  "errors": [
    "too many failed attempts"
  ]
}
```

**Validations:**
- ✓ Status code is 429
- ✓ Rate limit error is returned
- ✓ User must wait for cooldown period or generate new OTP

### Scenario 7: Unauthorized Access to KYC Status

```bash
curl -X GET http://localhost:8080/api/v1/kyc/status/other-user-456 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-User-ID: test-user-123"
```

**Expected Response:**
```json
{
  "status_code": 403,
  "message": "Unauthorized to access this user's KYC status",
  "errors": [
    "forbidden"
  ]
}
```

**Validations:**
- ✓ Status code is 403
- ✓ Unauthorized error is returned
- ✓ No KYC data is leaked

### Scenario 8: Sandbox API Rate Limiting

**Prerequisites:**
- Make multiple rapid requests to exceed Sandbox API rate limits

```bash
# Make 100 requests in quick succession
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer YOUR_JWT_TOKEN" \
    -d '{
      "user_id": "test-user-123",
      "aadhaar_number": "123456789012",
      "consent": "Y"
    }' &
done
wait
```

**Expected Response (after rate limit):**
```json
{
  "status_code": 429,
  "message": "Rate limit exceeded. Please try again later",
  "errors": [
    "too many requests"
  ]
}
```

**Validations:**
- ✓ Status code is 429
- ✓ Rate limit error is returned
- ✓ Retry logic is working (3 retries with exponential backoff)
- ✓ Requests are throttled gracefully

## Database Validation Queries

### Check Verification Record
```sql
SELECT
  id,
  user_id,
  verification_status,
  kyc_status,
  reference_id,
  transaction_id,
  name,
  gender,
  date_of_birth,
  photo_url,
  created_at,
  updated_at
FROM aadhaar_verifications
WHERE user_id = 'test-user-123'
ORDER BY created_at DESC
LIMIT 1;
```

### Check OTP Attempts
```sql
SELECT
  id,
  aadhaar_verification_id,
  otp_hash,
  attempt_status,
  ip_address,
  user_agent,
  created_at
FROM otp_attempts
WHERE aadhaar_verification_id = (
  SELECT id FROM aadhaar_verifications WHERE user_id = 'test-user-123'
)
ORDER BY created_at DESC;
```

### Check User Profile Update
```sql
SELECT
  id,
  full_name,
  aadhaar_verified,
  kyc_status,
  updated_at
FROM user_profiles
WHERE user_id = 'test-user-123';
```

### Check Address Creation
```sql
SELECT
  id,
  user_id,
  address_type,
  street,
  district,
  state,
  pincode,
  created_at
FROM addresses
WHERE user_id = 'test-user-123'
ORDER BY created_at DESC
LIMIT 1;
```

## S3 Validation

### Check Photo Upload
```bash
# Using AWS CLI
aws s3 ls s3://your-bucket-name/aadhaar/photos/test-user-123/ --recursive

# Expected output:
# 2025-11-18 12:34:56    45678 aadhaar/photos/test-user-123/aadhaar_1234567890.jpg
```

### Verify Photo Accessibility
```bash
# Download and verify
aws s3 cp s3://your-bucket-name/aadhaar/photos/test-user-123/aadhaar_1234567890.jpg ./test_photo.jpg

# Check file size
ls -lh test_photo.jpg

# Verify it's a valid JPEG
file test_photo.jpg
```

## Monitoring & Logging

### Application Logs
```bash
# Check OTP generation logs
tail -f logs/aaa-service.log | grep "OTP generation"

# Check OTP verification logs
tail -f logs/aaa-service.log | grep "OTP verification"

# Check Sandbox API logs
tail -f logs/aaa-service.log | grep "Sandbox API"

# Check photo upload logs
tail -f logs/aaa-service.log | grep "Photo upload"
```

### Expected Log Patterns

**OTP Generation:**
```
INFO  Sending OTP generation request to Sandbox API  endpoint=/kyc/aadhaar/okyc/otp method=POST aadhaar_masked=XXXX-XXXX-9012
INFO  Received response from Sandbox API  status_code=200 response_time_ms=234 endpoint=/kyc/aadhaar/okyc/otp
INFO  OTP generated successfully  transaction_id=TXN1234567890 reference_id=123456789
```

**OTP Verification:**
```
INFO  Sending OTP verification request to Sandbox API  endpoint=/kyc/aadhaar/okyc/otp/verify method=POST reference_id=123456789
INFO  Received response from Sandbox API  status_code=200 response_time_ms=456 endpoint=/kyc/aadhaar/okyc/otp/verify
INFO  OTP verified successfully  transaction_id=TXN1234567890 name=JOHN DOE status=success
INFO  Photo upload successful  user_id=test-user-123 photo_url=https://s3.amazonaws.com/...
```

## Performance Testing

### Load Test Configuration
```bash
# Install Apache Bench
sudo apt-get install apache2-utils

# Test OTP generation (100 requests, 10 concurrent)
ab -n 100 -c 10 -T 'application/json' -H 'Authorization: Bearer YOUR_JWT_TOKEN' \
  -p otp_request.json \
  http://localhost:8080/api/v1/kyc/aadhaar/otp

# Test OTP verification (50 requests, 5 concurrent)
ab -n 50 -c 5 -T 'application/json' -H 'Authorization: Bearer YOUR_JWT_TOKEN' \
  -p verify_request.json \
  http://localhost:8080/api/v1/kyc/aadhaar/otp/verify
```

**otp_request.json:**
```json
{
  "user_id": "test-user-123",
  "aadhaar_number": "123456789012",
  "consent": "Y"
}
```

**verify_request.json:**
```json
{
  "user_id": "test-user-123",
  "reference_id": "REF123456789",
  "otp": "123456"
}
```

### Performance Benchmarks

**Target Metrics:**
- OTP generation: < 500ms average response time
- OTP verification: < 1000ms average response time (includes photo upload)
- KYC status retrieval: < 100ms average response time
- Success rate: > 95%
- Concurrent users: 100+

## Security Testing

### Test SQL Injection
```bash
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "test'\'' OR '\''1'\''='\''1",
    "aadhaar_number": "123456789012",
    "consent": "Y"
  }'
```

**Expected:** Request should be safely handled, no SQL injection occurs

### Test XSS
```bash
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "<script>alert(1)</script>",
    "aadhaar_number": "123456789012",
    "consent": "Y"
  }'
```

**Expected:** Script tags should be escaped/sanitized

### Test Authentication
```bash
# No token
curl -X POST http://localhost:8080/api/v1/kyc/aadhaar/otp \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user-123",
    "aadhaar_number": "123456789012",
    "consent": "Y"
  }'
```

**Expected:** 401 Unauthorized

## Cleanup After Testing

```sql
-- Delete test verification records
DELETE FROM otp_attempts WHERE aadhaar_verification_id IN (
  SELECT id FROM aadhaar_verifications WHERE user_id = 'test-user-123'
);

DELETE FROM aadhaar_verifications WHERE user_id = 'test-user-123';

-- Reset user profile
UPDATE user_profiles
SET aadhaar_verified = false, kyc_status = 'PENDING'
WHERE user_id = 'test-user-123';

-- Delete test addresses
DELETE FROM addresses WHERE user_id = 'test-user-123';
```

```bash
# Delete S3 test photos
aws s3 rm s3://your-bucket-name/aadhaar/photos/test-user-123/ --recursive
```

## Troubleshooting

### Issue: OTP Not Received
- Check Sandbox API logs for errors
- Verify Aadhaar number is valid (12 digits)
- Check Sandbox API credentials
- Verify network connectivity to Sandbox API

### Issue: OTP Verification Fails
- Ensure OTP hasn't expired (5 minute window)
- Verify reference ID matches the one from generation
- Check attempts haven't exceeded max (3)
- Validate OTP format (6 digits)

### Issue: Photo Upload Fails
- Check S3 credentials
- Verify bucket exists and is accessible
- Check photo size is under limit (5MB)
- Ensure base64 decoding works correctly

### Issue: User Profile Not Updated
- Check user exists in database
- Verify user service adapter is working
- Check database transaction logs
- Ensure no database constraints are violated

## Automated E2E Test Script

```bash
#!/bin/bash
# e2e-test.sh

set -e

BASE_URL="http://localhost:8080/api/v1"
USER_ID="test-user-$(date +%s)"
AADHAAR="123456789012"
TOKEN="YOUR_JWT_TOKEN"

echo "=== Aadhaar Verification E2E Test ==="
echo "User ID: $USER_ID"

# 1. Generate OTP
echo ""
echo "Step 1: Generating OTP..."
GENERATE_RESPONSE=$(curl -s -X POST "$BASE_URL/kyc/aadhaar/otp" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"aadhaar_number\": \"$AADHAAR\",
    \"consent\": \"Y\"
  }")

echo "$GENERATE_RESPONSE" | jq .

REFERENCE_ID=$(echo "$GENERATE_RESPONSE" | jq -r '.reference_id')
echo "Reference ID: $REFERENCE_ID"

# 2. Verify OTP (using test OTP from Sandbox)
echo ""
echo "Step 2: Verifying OTP..."
sleep 2  # Wait a bit

VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/kyc/aadhaar/otp/verify" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"reference_id\": \"$REFERENCE_ID\",
    \"otp\": \"123456\"
  }")

echo "$VERIFY_RESPONSE" | jq .

# 3. Get KYC Status
echo ""
echo "Step 3: Checking KYC status..."
STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/kyc/status/$USER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-User-ID: $USER_ID")

echo "$STATUS_RESPONSE" | jq .

echo ""
echo "=== Test Complete ==="
```

## Success Criteria

All E2E tests pass when:
- ✓ OTP generation completes in < 500ms
- ✓ OTP verification completes in < 1000ms
- ✓ Photo is uploaded to S3 successfully
- ✓ User profile is updated with Aadhaar data
- ✓ Address record is created
- ✓ KYC status shows VERIFIED
- ✓ All validations work correctly
- ✓ Error handling is robust
- ✓ Retry logic works for Sandbox API failures
- ✓ Audit logs are created for all operations
