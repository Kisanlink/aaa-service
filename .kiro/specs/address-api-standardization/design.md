# Address API Standardization - Architectural Design

## Executive Summary

This document provides architectural guidance for standardizing the address API system on the Indian address format across both HTTP and gRPC interfaces. Currently, there is a mismatch between the HTTP API (using Indian format) and gRPC API (using international format), which creates maintenance overhead and potential data loss.

## Current State Analysis

### HTTP API (Indian Format)
- **Fields**: house, street, landmark, post_office, subdistrict, district, vtc, state, country, pincode, full_address
- **Implementation**: Fully aligned with the domain model
- **Status**: Working correctly

### gRPC API (International Format)
- **Fields**: address_line_1, address_line_2, city, state, postal_code, country, metadata
- **Implementation**: Uses lossy mapping through metadata field
- **Issues**:
  - Data loss during conversion (post_office, subdistrict, district stored in metadata)
  - Complex bidirectional mapping logic in grpc_server/address_handler.go
  - Inconsistent API experience

### Domain Model
- **Location**: internal/entities/models/address.go
- **Format**: Indian address format
- **Database**: Stores all Indian format fields natively

## Architectural Recommendations

### 1. Approach: Versioned Proto Migration

**Recommendation**: Create a new proto version (v2) with Indian format while maintaining backward compatibility.

**Implementation Strategy**:
```protobuf
// address_v2.proto
syntax = "proto3";

package pb.v2;

message Address {
    string id = 1;
    string user_id = 2;
    string type = 3;

    // Indian address fields
    string house = 4;
    string street = 5;
    string landmark = 6;
    string post_office = 7;
    string subdistrict = 8;
    string district = 9;
    string vtc = 10; // Village/Town/City
    string state = 11;
    string country = 12;
    string pincode = 13;
    string full_address = 14;

    // Metadata and system fields
    bool is_primary = 15;
    bool is_active = 16;
    map<string, string> metadata = 17;
    google.protobuf.Timestamp created_at = 18;
    google.protobuf.Timestamp updated_at = 19;

    // Deprecated international format fields for backward compatibility
    // These will be populated for old clients
    string address_line_1 = 100 [deprecated = true];
    string address_line_2 = 101 [deprecated = true];
    string city = 102 [deprecated = true];
    string postal_code = 103 [deprecated = true];
}
```

### 2. Backward Compatibility Strategy

**Phase 1: Dual Support (3-6 months)**
- Maintain both v1 and v2 proto definitions
- Implement a compatibility layer in the gRPC handler
- Auto-populate deprecated fields for backward compatibility
- Add API version header support

**Phase 2: Migration (3 months)**
- Notify clients about deprecation
- Provide migration tools/scripts
- Monitor usage of v1 endpoints
- Support both formats with preference for v2

**Phase 3: Deprecation (1 month)**
- Remove v1 support
- Clean up compatibility code
- Final optimization

### 3. Migration Implementation

#### Step 1: Create Versioned Proto Structure
```
pkg/proto/
├── v1/
│   └── address.proto (current)
└── v2/
    └── address.proto (Indian format)
```

#### Step 2: Implement Compatibility Layer
```go
// internal/grpc_server/compatibility/address_mapper.go
type AddressMapper struct {
    logger *zap.Logger
}

func (m *AddressMapper) ToV2(v1addr *pbv1.Address) *pbv2.Address {
    // Map v1 to v2 format
}

func (m *AddressMapper) ToV1(v2addr *pbv2.Address) *pbv1.Address {
    // Map v2 to v1 format for backward compatibility
}

func (m *AddressMapper) SupportsIndianFormat(headers metadata.MD) bool {
    // Check client version from headers
}
```

#### Step 3: Update gRPC Handler
```go
func (h *AddressHandler) CreateAddress(ctx context.Context, req interface{}) (*pb.CreateAddressResponse, error) {
    md, _ := metadata.FromIncomingContext(ctx)

    if h.mapper.SupportsIndianFormat(md) {
        // Handle v2 request with Indian format
    } else {
        // Handle v1 request with compatibility layer
    }
}
```

### 4. Validation and Security Considerations

#### Input Validation
```go
// internal/validators/address_validator.go
type AddressValidator struct {
    pincodeService PincodeService // External validation service
}

func (v *AddressValidator) ValidateIndianAddress(addr *models.Address) error {
    // Validate pincode format (6 digits)
    if err := v.validatePincode(addr.Pincode); err != nil {
        return err
    }

    // Validate state against known Indian states
    if err := v.validateState(addr.State); err != nil {
        return err
    }

    // Cross-validate pincode with district/state
    if err := v.crossValidate(addr); err != nil {
        return err
    }

    return nil
}
```

#### Security Measures
1. **Input Sanitization**: Prevent XSS/SQL injection in address fields
2. **Rate Limiting**: Implement per-user address creation limits
3. **Authorization**: Ensure users can only modify their own addresses
4. **Audit Logging**: Track all address modifications
5. **PII Protection**: Encrypt sensitive address data at rest

### 5. Database Migration Strategy

No database changes required as the model already uses Indian format. However:

1. **Add Indexes**: Create composite indexes for efficient searching
```sql
CREATE INDEX idx_addresses_pincode ON addresses(pincode);
CREATE INDEX idx_addresses_district_state ON addresses(district, state);
CREATE INDEX idx_addresses_user_id ON addresses(user_id);
```

2. **Add Constraints**: Ensure data integrity
```sql
ALTER TABLE addresses ADD CONSTRAINT chk_pincode CHECK (pincode ~ '^\d{6}$');
```

### 6. Client Migration Guide

#### For gRPC Clients
```go
// Old client code (v1)
client := pb.NewAddressServiceClient(conn)
req := &pb.CreateAddressRequest{
    AddressLine_1: "123 Main St",
    City: "Mumbai",
    PostalCode: "400001",
}

// New client code (v2)
client := pbv2.NewAddressServiceClient(conn)
req := &pbv2.CreateAddressRequest{
    House: "123",
    Street: "Main Street",
    VTC: "Mumbai",
    District: "Mumbai",
    State: "Maharashtra",
    Pincode: "400001",
}
```

### 7. Monitoring and Observability

#### Metrics to Track
- API version usage distribution
- Conversion errors between formats
- Validation failure rates by field
- Client migration progress

#### Logging Strategy
```go
logger.Info("Address API request",
    zap.String("api_version", version),
    zap.String("format", format),
    zap.String("client_id", clientID),
    zap.Bool("conversion_required", needsConversion),
)
```

### 8. Performance Considerations

1. **Caching**: Cache validated addresses and pincode mappings
2. **Batch Operations**: Support bulk address operations
3. **Async Validation**: Perform complex validations asynchronously
4. **Connection Pooling**: Optimize database connections

### 9. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing clients | High | Gradual migration with backward compatibility |
| Data loss during migration | Medium | Comprehensive testing and rollback plan |
| Performance degradation | Low | Caching and optimization strategies |
| Validation complexity | Medium | External pincode validation service |

### 10. Implementation Timeline

| Phase | Duration | Activities |
|-------|----------|------------|
| Planning | 1 week | Finalize design, notify stakeholders |
| Development | 2 weeks | Implement v2 proto, compatibility layer |
| Testing | 1 week | Unit, integration, and backward compatibility tests |
| Staged Rollout | 2 weeks | Deploy to staging, then gradual production rollout |
| Client Migration | 3 months | Support clients in migration |
| Cleanup | 1 week | Remove deprecated code |

## Conclusion

The recommended approach provides a safe, gradual migration path that:
- Maintains backward compatibility
- Aligns with the domain model
- Provides better data fidelity
- Enables future enhancements
- Minimizes risk to existing clients

The Indian address format standardization will improve consistency, reduce complexity, and provide better support for Indian-specific address requirements.
