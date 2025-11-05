# ADR-001: Indian Address Format Standardization

## Status
Accepted - 2025-11-05

## Context
The AAA service currently has inconsistent address formats between HTTP and gRPC APIs:

1. **HTTP API**: Uses Indian address format with fields like `house`, `street`, `vtc`, `pincode`, `post_office`, `subdistrict`, `district`
2. **gRPC API (v1)**: Uses international format with fields like `address_line_1`, `address_line_2`, `city`, `postal_code`

This inconsistency creates:
- **Data Loss**: Converting between formats loses granularity (e.g., `house` and `street` get combined into `address_line_1`)
- **Complex Mapping Logic**: Requires bidirectional conversion with workarounds using metadata
- **Poor Developer Experience**: Clients must maintain two different request formats for the same operation
- **Documentation Burden**: Must explain both formats and their differences

## Decision
We will standardize on the **Indian address format** across all APIs (HTTP and gRPC) by:

1. Creating a new gRPC proto v2 that uses Indian address fields directly
2. Maintaining backward compatibility with v1 proto for existing clients
3. Adding the `user_id` field to the Address model to establish proper user-address relationships
4. Implementing proper Indian address validation (6-digit pincode format, valid state names)

### Rationale for Indian Format

1. **Domain Alignment**: Kisanlink serves the Indian agricultural ecosystem (farmers, FPOs, cooperatives)
2. **Government Integration**: PM-KISAN, e-NAM, and other government programs require Indian postal hierarchy
3. **Data Accuracy**: Critical for subsidy distribution and preventing leakage
4. **Existing Infrastructure**: The domain model and HTTP API already use Indian format
5. **No Data Loss**: Direct mapping between API and database with no lossy conversions

## Implementation Details

### Proto V2 Structure
```protobuf
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
    string vtc = 10;        // Village/Town/City
    string state = 11;
    string country = 12;
    string pincode = 13;
    string full_address = 14;

    // System fields
    bool is_primary = 15;
    bool is_active = 16;
    map<string, string> metadata = 17;
}
```

### Database Changes
- Added `user_id` column to `addresses` table
- Added indexes: `idx_addresses_user_id`, `idx_addresses_pincode`, `idx_addresses_district_state`
- Added check constraint for pincode format: `^[1-9][0-9]{5}$`

### Validation Rules
1. **Pincode**: 6 digits, cannot start with 0 (Indian postal code standard)
2. **State**: Must be a valid Indian state or union territory
3. **Required Fields**: `user_id`, at least one location field (house, street, vtc, or district)

### Version Detection
- V2 clients send `api-version: v2` or `x-api-version: v2` in gRPC metadata
- V1 clients (no header) continue to work with backward compatibility layer
- Converter layer handles bidirectional mapping between v1/v2 and domain model

## Consequences

### Positive
- ✅ **Consistency**: Single address format across all APIs
- ✅ **No Data Loss**: Direct mapping preserves all fields
- ✅ **Better DX**: Developers use the same format everywhere
- ✅ **Domain Alignment**: Matches Indian postal system and government requirements
- ✅ **Maintainability**: Simpler codebase without complex conversion logic
- ✅ **Performance**: No lossy conversion overhead in v2 API

### Negative
- ⚠️ **Client Migration**: Existing gRPC clients need to upgrade to v2
- ⚠️ **Dual Maintenance**: Must maintain v1 compatibility for transition period (6 months)
- ⚠️ **Learning Curve**: New clients must learn Indian address structure

### Mitigation Strategies
1. **Backward Compatibility**: v1 API continues to work for 6 months minimum
2. **Client SDKs**: Provide updated SDKs with v2 support
3. **Migration Guide**: Comprehensive documentation and examples
4. **Gradual Rollout**: Monitor adoption and extend v1 support if needed
5. **Version Detection**: Automatic based on headers, no breaking changes

## Alternatives Considered

### 1. Standardize on International Format ❌
**Rejected** because:
- Loses Indian-specific fields (post_office, subdistrict)
- Requires changing database schema
- Breaks existing HTTP API clients
- Doesn't align with domain requirements

### 2. Support Both Formats Equally ❌
**Rejected** because:
- Doubles maintenance burden
- Increases code complexity
- Confuses API consumers
- Doesn't solve the consistency problem

### 3. Use Adapter Pattern Only ❌
**Rejected** because:
- Still requires lossy conversion
- Doesn't eliminate complexity
- Temporary workaround, not a solution

## Implementation Timeline

| Phase | Duration | Status |
|-------|----------|--------|
| Proto v2 creation | 2025-11-05 | ✅ Completed |
| Address model update | 2025-11-05 | ✅ Completed |
| Handlers and converters | 2025-11-05 | ✅ Completed |
| Validation implementation | 2025-11-05 | ✅ Completed |
| Database migration | 2025-11-05 | ✅ Completed |
| Testing | Pending | 🔄 In Progress |
| Documentation | Pending | 🔄 In Progress |
| Client migration | 6 months | ⏳ Not Started |
| V1 deprecation | After 6 months | ⏳ Not Started |

## Success Metrics

1. **Zero Breaking Changes**: V1 clients continue to work without modification
2. **High Adoption**: >50% of clients using v2 within 3 months
3. **No Data Loss**: All address fields preserved in v2 API
4. **Performance**: <5ms overhead for format conversion in v1 compatibility layer
5. **Error Rate**: <1% increase in validation errors

## References

1. Indian Postal System: https://www.indiapost.gov.in/
2. Government Digital Agriculture Programs: PM-KISAN, e-NAM
3. gRPC Versioning Best Practices: https://grpc.io/docs/guides/versioning/
4. Proto3 Language Guide: https://developers.google.com/protocol-buffers/docs/proto3

## Approval

- **Decision Date**: 2025-11-05
- **Approved By**: Backend Architecture Team
- **Review Status**: ✅ Approved with implementation completed

## Notes

- This ADR supersedes any previous informal decisions about address format
- V1 deprecation timeline may be extended based on client adoption metrics
- Additional validation rules may be added based on government integration requirements
- Future consideration: Add pincode-to-location lookup service for auto-population
