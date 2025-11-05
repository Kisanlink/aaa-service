# Address API Standardization - Quick Reference

## Problem Summary
- **Current Issue**: gRPC API uses international format (address_line_1, city, postal_code) while HTTP API and domain model use Indian format
- **Impact**: Data loss, complex mapping logic, inconsistent API experience
- **Solution**: Standardize on Indian format across all APIs

## Recommended Approach: API Versioning with Backward Compatibility

### Option 1: Direct Proto Update (Breaking Change) ❌
**NOT RECOMMENDED** - Would break all existing gRPC clients

### Option 2: Versioned Migration (Recommended) ✅
Create proto v2 with Indian format while maintaining v1 for backward compatibility

## Implementation Steps

### Step 1: Create New Proto (Week 1)
```bash
# Create new proto file
mkdir -p pkg/proto/v2
# Copy provided proto definition to pkg/proto/v2/address.proto

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/proto/v2/address.proto
```

### Step 2: Add Version Detection (Week 1)
- Check for `api-version` header in gRPC metadata
- Default to v1 for existing clients
- Use v2 for new clients

### Step 3: Implement Dual Handler (Week 2)
```go
// Pseudo-code for handler
if version == "v2" {
    // Direct Indian format handling
    return handleIndianFormat(request)
} else {
    // Backward compatibility with conversion
    return handleInternationalFormat(request)
}
```

### Step 4: Test Both Paths (Week 2)
- Unit tests for converters
- Integration tests for both v1 and v2 clients
- Performance tests to ensure no degradation

### Step 5: Gradual Rollout (Week 3-6)
1. Deploy to staging
2. Test with sample clients
3. Roll out to 10% production
4. Monitor metrics
5. Increase to 50%, then 100%

## Critical Files to Modify

### New Files to Create
```
pkg/proto/v2/address.proto                          # New proto definition
internal/grpc_server/version/detector.go            # Version detection
internal/grpc_server/converters/address_converter.go # Format conversion
internal/validators/indian_address_validator.go     # Indian format validation
```

### Files to Update
```
internal/grpc_server/address_handler.go  # Add version-aware handling
internal/entities/models/address.go      # Add UserID field if missing
cmd/server/grpc.go                       # Register v2 service
```

## Indian Address Format Fields

| Field | Description | Validation |
|-------|-------------|------------|
| house | House/Flat number | Max 255 chars |
| street | Street name | Max 255 chars |
| landmark | Nearby landmark | Max 255 chars |
| post_office | Post office name | Max 255 chars |
| subdistrict | Subdistrict/Taluk | Max 255 chars |
| district | District name | Max 255 chars, Required |
| vtc | Village/Town/City | Max 255 chars |
| state | Indian state | Max 255 chars, Required, Must be valid state |
| country | Country | Max 255 chars, Default "India" |
| pincode | Indian postal code | Exactly 6 digits, Format: [1-9][0-9]{5} |
| full_address | Complete address | Max 1000 chars |

## Backward Compatibility Mapping

| V1 Field (International) | V2 Field (Indian) | Notes |
|-------------------------|-------------------|-------|
| address_line_1 | house + street | Split on comma |
| address_line_2 | landmark | Direct mapping |
| city | vtc | Direct mapping |
| state | state | Direct mapping |
| postal_code | pincode | Direct mapping |
| country | country | Direct mapping |
| metadata["post_office"] | post_office | Extract from metadata |
| metadata["subdistrict"] | subdistrict | Extract from metadata |
| metadata["district"] | district | Extract from metadata |

## Testing Checklist

- [ ] V1 clients continue to work without changes
- [ ] V2 clients can use Indian format directly
- [ ] Data conversion preserves all fields
- [ ] Validation works for Indian addresses
- [ ] Performance meets SLA targets
- [ ] Rollback procedure tested

## Monitoring Metrics

```prometheus
# Track API version usage
address_api_version_total{version="v1", method="CreateAddress"}
address_api_version_total{version="v2", method="CreateAddress"}

# Track conversion errors
address_conversion_errors_total

# Track validation failures
address_validation_errors_total{field="pincode"}
```

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Client breakage | Maintain v1 compatibility for 6+ months |
| Data loss | Comprehensive field mapping and testing |
| Performance impact | Caching and optimized converters |
| Rollback needed | Feature flags and versioned deployments |

## Decision Points

1. **Timeline**: 6-week implementation with 3-month migration window
2. **Client Communication**: Notify all clients by Week 2
3. **Deprecation Date**: V1 support ends after 6 months
4. **Performance Target**: No more than 5ms overhead for conversion

## Commands for Implementation

```bash
# Generate proto
make proto-gen

# Run tests
go test ./internal/grpc_server/... -v

# Check for breaking changes
buf breaking pkg/proto/v2 --against pkg/proto

# Deploy with feature flag
AAA_ENABLE_ADDRESS_V2=true make deploy
```

## Contact for Questions

- Architecture: Backend Architecture Team
- Implementation: Backend Development Team
- Client Migration: API Platform Team
- Security Review: Security Team

## References

- [Proto3 Language Guide](https://developers.google.com/protocol-buffers/docs/proto3)
- [gRPC Versioning Best Practices](https://grpc.io/docs/guides/versioning/)
- [Indian Postal Code Database](https://www.indiapost.gov.in/)
