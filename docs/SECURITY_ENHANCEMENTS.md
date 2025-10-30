# Security Enhancements Documentation

This document outlines the comprehensive security enhancements implemented in the AAA service to address requirements 9.1, 9.2, 9.4, 9.5, and 10.5.

## Overview

The security enhancements provide multiple layers of protection against common security threats including:

- Rate limiting and brute force protection
- Comprehensive audit logging
- Input sanitization and validation
- Secure error handling
- Enhanced security headers and CORS configuration

## 1. Rate Limiting Implementation

### 1.1 Authentication Rate Limiting

**Location**: `internal/middleware/middleware.go`

**Features**:

- 5 login attempts per minute per IP (configurable)
- Exponential backoff after failed attempts
- Temporary IP blocking after 10 failed attempts
- Automatic cleanup of old rate limiting data

**Configuration**:

```bash
AAA_AUTH_RATE_LIMIT_REQUESTS_PER_MINUTE=5
AAA_AUTH_RATE_LIMIT_BURST_SIZE=3
AAA_AUTH_MAX_FAILED_ATTEMPTS=10
AAA_AUTH_BLOCK_DURATION=15m
```

### 1.2 MPIN Rate Limiting

**Features**:

- 3 MPIN operations per minute per IP
- Stricter blocking after 5 failed attempts
- 15-minute block duration for MPIN failures

**Configuration**:

```bash
AAA_MPIN_RATE_LIMIT_REQUESTS_PER_MINUTE=3
AAA_MPIN_RATE_LIMIT_BURST_SIZE=2
AAA_MPIN_MAX_FAILED_ATTEMPTS=5
AAA_MPIN_BLOCK_DURATION=15m
```

### 1.3 General Rate Limiting

**Features**:

- 100 requests per minute for general endpoints
- Configurable burst size and cleanup intervals

## 2. Audit Logging

### 2.1 Security Audit Middleware

**Location**: `internal/middleware/audit.go`

**Features**:

- Comprehensive logging of all security-sensitive operations
- Automatic detection of suspicious activity patterns
- Detailed request/response metadata capture
- IP-based tracking and analysis

**Logged Events**:

- Authentication attempts (success/failure)
- MPIN operations (setup/update)
- User registration and management
- Role assignments and removals
- Administrative operations
- Suspicious activity detection

### 2.2 Suspicious Activity Detection

**Patterns Detected**:

- Missing or suspicious user agents
- Unusual request patterns
- Rate limiting triggers
- Multiple failed authentication attempts
- Requests from suspicious IP addresses

**Configuration**:

```bash
AAA_ENABLE_SECURITY_AUDIT=true
AAA_ENABLE_DATA_ACCESS_AUDIT=true
AAA_LOG_SUSPICIOUS_ACTIVITY=true
AAA_AUDIT_RETENTION_DAYS=90
```

## 3. Input Sanitization and Validation

### 3.1 Input Sanitization Middleware

**Location**: `internal/middleware/sanitization.go`

**Features**:

- Comprehensive input sanitization for all request data
- XSS pattern removal
- SQL injection pattern detection
- Control character filtering
- JSON structure validation

**Sanitization Process**:

1. Query parameter sanitization
2. Path parameter sanitization
3. Header sanitization (selective)
4. Request body sanitization (JSON and text)

### 3.2 Enhanced Validation

**Location**: `utils/validator.go`

**New Validation Functions**:

- `ValidateMPin()` - MPIN format and strength validation
- `ValidateAndSanitizeString()` - Combined validation and sanitization
- `SanitizeInput()` - Input sanitization
- Enhanced password validation with common pattern detection

**Features**:

- MPIN strength validation (prevents weak patterns)
- Password strength assessment
- Input length limits
- SQL injection pattern detection
- XSS pattern detection

### 3.3 Security Utilities

**Location**: `internal/security/utils.go`

**Features**:

- Secure password hashing with bcrypt
- MPIN hashing with additional salt
- Cryptographically secure token generation
- Password strength assessment
- Constant-time string comparison
- Email and phone number validation

## 4. Secure Error Handling

### 4.1 Enhanced Error Types

**Location**: `pkg/errors/errors.go`

**Features**:

- Secure error constructors that don't leak sensitive information
- Automatic sanitization of error messages
- Standardized error response format
- Specific error types for different security scenarios

**New Error Types**:

- `NewSecureUnauthorizedError()`
- `NewSecureForbiddenError()`
- `NewSecureNotFoundError()`
- `NewAuthenticationFailedError()`
- `NewAccountLockedError()`
- `NewRateLimitError()`

### 4.2 Secure Error Handler Middleware

**Location**: `internal/middleware/errorHandler.go`

**Features**:

- Prevents information leakage in error responses
- Logs detailed errors server-side for debugging
- Returns sanitized error messages to clients
- Includes request ID for traceability

## 5. Security Headers and CORS

### 5.1 Enhanced Security Headers

**Location**: `internal/middleware/middleware.go`

**Headers Implemented**:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security` (production only)
- `Content-Security-Policy` (configurable)
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy` for feature restrictions
- Cache control for sensitive endpoints

### 5.2 CORS Configuration

**Features**:

- Configurable allowed origins, methods, and headers
- Proper preflight request handling
- Credential support configuration
- Configurable max age for preflight caching

**Configuration**:

```bash
AAA_CORS_ALLOWED_ORIGINS=https://example.com,https://app.example.com
AAA_CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS,PATCH
AAA_CORS_ALLOW_CREDENTIALS=true
AAA_CORS_MAX_AGE=86400
```

## 6. Security Configuration

### 6.1 Centralized Security Configuration

**Location**: `internal/config/security.go`

**Features**:

- Environment-based configuration
- Default secure values
- Validation of security settings
- Production vs development mode detection

### 6.2 Configuration Categories

1. **Rate Limiting Configuration**
2. **Input Validation Configuration**
3. **Security Headers Configuration**
4. **CORS Configuration**
5. **Audit Configuration**
6. **Encryption Configuration**

## 7. Route-Level Security

### 7.1 Enhanced Route Setup

**Location**: `internal/routes/setup.go` and `internal/routes/auth_routes.go`

**Features**:

- Global security middleware application
- Endpoint-specific rate limiting
- Input validation for all routes
- Content type validation
- JSON structure validation

### 7.2 Security Middleware Stack

**Order of Application**:

1. Security headers
2. CORS handling
3. Request ID generation
4. Request size limiting
5. Timeout handling
6. Rate limiting (general)
7. Input sanitization
8. Content type validation
9. JSON structure validation
10. Authentication (protected routes)
11. Authorization (protected routes)
12. Endpoint-specific rate limiting

## 8. Testing

### 8.1 Security Test Suite

**Location**: `internal/security/security_test.go`

**Test Coverage**:

- Password hashing and verification
- MPIN hashing and verification
- Token generation
- Password strength assessment
- Input sanitization
- Email and phone validation
- Secure string comparison

### 8.2 Benchmark Tests

**Performance Tests**:

- Password hashing performance
- Password verification performance
- Token generation performance
- Input sanitization performance

## 9. Deployment Considerations

### 9.1 Environment Variables

**Required Environment Variables**:

```bash
# JWT Configuration
AAA_JWT_SECRET=your-secret-key

# Rate Limiting
AAA_AUTH_RATE_LIMIT_REQUESTS_PER_MINUTE=5
AAA_MPIN_RATE_LIMIT_REQUESTS_PER_MINUTE=3

# Security Features
AAA_ENABLE_INPUT_SANITIZATION=true
AAA_ENABLE_SECURITY_AUDIT=true
AAA_ENABLE_HSTS=true

# CORS Configuration
AAA_CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

### 9.2 Production Recommendations

1. **Enable HSTS** in production environments
2. **Configure specific CORS origins** instead of wildcards
3. **Set appropriate rate limits** based on expected traffic
4. **Enable all audit logging** for compliance
5. **Use strong JWT secrets** (minimum 32 characters)
6. **Configure proper CSP headers** for your application

### 9.3 Monitoring and Alerting

**Recommended Monitoring**:

- Rate limiting trigger frequency
- Authentication failure rates
- Suspicious activity detection alerts
- Error rate monitoring
- Performance impact of security middleware

## 10. Security Best Practices Implemented

1. **Defense in Depth**: Multiple layers of security controls
2. **Fail Secure**: Secure defaults and fail-safe mechanisms
3. **Least Privilege**: Minimal permissions and access controls
4. **Input Validation**: Comprehensive input sanitization and validation
5. **Audit Logging**: Complete audit trail for security events
6. **Rate Limiting**: Protection against brute force and DoS attacks
7. **Secure Error Handling**: No information leakage in error responses
8. **Security Headers**: Protection against common web vulnerabilities

## 11. Compliance and Standards

The implemented security enhancements help meet various compliance requirements:

- **OWASP Top 10** protection
- **PCI DSS** requirements for authentication and logging
- **GDPR** requirements for data protection and audit trails
- **SOC 2** requirements for security controls and monitoring

## 12. Future Enhancements

Potential future security improvements:

1. **IP Geolocation Filtering**: Block requests from specific countries
2. **Device Fingerprinting**: Enhanced device-based security
3. **Machine Learning**: Anomaly detection for suspicious patterns
4. **Multi-Factor Authentication**: Additional authentication factors
5. **Certificate Pinning**: Enhanced transport security
6. **Web Application Firewall**: Additional layer of protection

## 13. Troubleshooting

### Common Issues and Solutions

1. **Rate Limiting Too Aggressive**:

   - Adjust `AAA_AUTH_RATE_LIMIT_REQUESTS_PER_MINUTE`
   - Increase burst size if needed

2. **CORS Issues**:

   - Verify `AAA_CORS_ALLOWED_ORIGINS` configuration
   - Check preflight request handling

3. **Input Sanitization Breaking Functionality**:

   - Review sanitization rules
   - Add exceptions for specific endpoints if needed

4. **Performance Impact**:
   - Monitor middleware performance
   - Adjust rate limiting cleanup intervals
   - Consider caching for validation results

### Debug Mode

For debugging security issues, enable detailed logging:

```bash
AAA_LOG_LEVEL=debug
AAA_ENABLE_SECURITY_AUDIT=true
AAA_LOG_SUSPICIOUS_ACTIVITY=true
```

This comprehensive security enhancement implementation provides robust protection against common security threats while maintaining performance and usability.
