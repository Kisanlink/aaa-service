# Address API Migration Guide - Indian Format Standardization

## Overview
This guide helps you migrate from the old gRPC Address API (v1 - international format) to the new v2 API (Indian format). The HTTP API remains unchanged as it already uses the Indian format.

## What Changed?

### Before (v1 - International Format)
```protobuf
message CreateAddressRequest {
    string user_id = 1;
    string address_line_1 = 3;      // Combined house + street
    string address_line_2 = 4;      // Landmark
    string city = 5;                // Village/Town/City
    string state = 6;
    string postal_code = 7;         // 6-digit pincode
    string country = 8;
    map<string, string> metadata = 10;  // Indian fields hidden here
}
```

### After (v2 - Indian Format)
```protobuf
message CreateAddressRequest {
    string user_id = 1;
    string house = 3;               // Separate field
    string street = 4;              // Separate field
    string landmark = 5;            // Direct field
    string post_office = 6;         // Direct field
    string subdistrict = 7;         // Direct field (taluka/tehsil)
    string district = 8;            // Direct field
    string vtc = 9;                 // Village/Town/City
    string state = 10;
    string country = 11;
    string pincode = 12;            // Indian postal code
}
```

## Why Migrate?

1. **No Data Loss**: House and street are separate fields, not concatenated
2. **Better Validation**: Proper Indian pincode and state validation
3. **First-Class Indian Fields**: post_office, subdistrict, district are direct fields, not in metadata
4. **Domain Alignment**: Matches Indian postal system hierarchy
5. **Future-Proof**: Better support for government integrations (PM-KISAN, e-NAM)

## Migration Path

### Option 1: Add Version Header (Recommended)
Add the API version header to your gRPC calls:

```go
// Before (v1 - no header)
ctx := context.Background()
resp, err := client.CreateAddress(ctx, &pb.CreateAddressRequest{
    UserId: "USER00000001",
    AddressLine_1: "123 Main St",
    City: "Mumbai",
    PostalCode: "400001",
})

// After (v2 - with header)
import "google.golang.org/grpc/metadata"

ctx := metadata.AppendToOutgoingContext(context.Background(), "api-version", "v2")
resp, err := client.CreateAddress(ctx, &pbv2.CreateAddressRequest{
    UserId: "USER00000001",
    House: "123",
    Street: "Main Street",
    Vtc: "Mumbai",
    District: "Mumbai Suburban",
    State: "Maharashtra",
    Pincode: "400001",
})
```

### Option 2: Continue Using V1 (Temporary)
V1 will be supported for **6 months** after v2 release. No changes needed, but you won't get the benefits of v2.

```go
// V1 continues to work (backward compatible)
resp, err := client.CreateAddress(ctx, &pb.CreateAddressRequest{
    UserId: "USER00000001",
    AddressLine_1: "123 Main St",
    City: "Mumbai",
    PostalCode: "400001",
    Metadata: map[string]string{
        "post_office": "Andheri",
        "subdistrict": "Andheri West",
        "district": "Mumbai Suburban",
    },
})
```

## Field Mapping Reference

| V1 Field (International) | V2 Field (Indian) | Notes |
|-------------------------|-------------------|-------|
| `address_line_1` | `house` + `street` | Split by comma in v1 |
| `address_line_2` | `landmark` | Direct mapping |
| `city` | `vtc` | Village/Town/City |
| `state` | `state` | Must be valid Indian state |
| `postal_code` | `pincode` | 6 digits, cannot start with 0 |
| `country` | `country` | Direct mapping |
| `metadata["post_office"]` | `post_office` | Now a direct field |
| `metadata["subdistrict"]` | `subdistrict` | Now a direct field |
| `metadata["district"]` | `district` | Now a direct field |

## Validation Changes

### V1 Validation (Lenient)
- `postal_code`: Any 6-digit string

### V2 Validation (Strict Indian Format)
- `pincode`: Must be 6 digits, cannot start with 0 (e.g., "400001" ✅, "000001" ❌)
- `state`: Must be a valid Indian state or union territory
- At least one location field required (house, street, vtc, or district)

## Code Examples

### Example 1: Creating an Address

#### V1 (Old Way)
```go
req := &pb.CreateAddressRequest{
    UserId: "USER00000001",
    AddressLine_1: "123, Main Street",
    AddressLine_2: "Near City Mall",
    City: "Mumbai",
    State: "Maharashtra",
    PostalCode: "400053",
    Country: "India",
    Metadata: map[string]string{
        "post_office": "Andheri",
        "subdistrict": "Andheri West",
        "district": "Mumbai Suburban",
    },
}
```

#### V2 (New Way)
```go
ctx := metadata.AppendToOutgoingContext(context.Background(), "api-version", "v2")

req := &pbv2.CreateAddressRequest{
    UserId: "USER00000001",
    House: "123",
    Street: "Main Street",
    Landmark: "Near City Mall",
    PostOffice: "Andheri",
    Subdistrict: "Andheri West",
    District: "Mumbai Suburban",
    Vtc: "Mumbai",
    State: "Maharashtra",
    Country: "India",
    Pincode: "400053",
}
```

### Example 2: Reading an Address

#### V1 Response
```go
address := resp.GetAddress()
// address_line_1 = "123, Main Street" (combined)
// city = "Mumbai"
// postal_code = "400053"
// metadata["district"] = "Mumbai Suburban"
```

#### V2 Response
```go
address := resp.GetAddress()
// house = "123" (separate)
// street = "Main Street" (separate)
// vtc = "Mumbai"
// district = "Mumbai Suburban" (direct field)
// pincode = "400053"
```

## Common Migration Issues

### Issue 1: Pincode Starting with 0
**Error**: `invalid pincode format: must be 6 digits, cannot start with 0`

**Cause**: V2 validates Indian pincode format strictly. No Indian pincode starts with 0.

**Solution**: Check your data source. If pincode is "000123", it's likely invalid data.

### Issue 2: Invalid State Name
**Error**: `invalid Indian state: Maharastra`

**Cause**: Typo in state name (correct spelling: "Maharashtra")

**Solution**: Use exact state names from the validation list (see below).

### Issue 3: Missing Location Fields
**Error**: `at least one location field is required`

**Cause**: All location fields (house, street, vtc, district) are empty.

**Solution**: Provide at least one location identifier.

## Valid Indian States

```
States:
- Andhra Pradesh
- Arunachal Pradesh
- Assam
- Bihar
- Chhattisgarh
- Goa
- Gujarat
- Haryana
- Himachal Pradesh
- Jharkhand
- Karnataka
- Kerala
- Madhya Pradesh
- Maharashtra
- Manipur
- Meghalaya
- Mizoram
- Nagaland
- Odisha
- Punjab
- Rajasthan
- Sikkim
- Tamil Nadu
- Telangana
- Tripura
- Uttar Pradesh
- Uttarakhand
- West Bengal

Union Territories:
- Andaman and Nicobar Islands
- Chandigarh
- Dadra and Nagar Haveli
- Daman and Diu
- Delhi
- Jammu and Kashmir
- Ladakh
- Lakshadweep
- Puducherry
```

## Testing Your Migration

### Step 1: Test V2 Endpoint
```bash
# Test v2 gRPC endpoint
grpcurl -H "api-version: v2" \
  -d '{
    "user_id": "USER00000001",
    "house": "123",
    "street": "Main Street",
    "vtc": "Mumbai",
    "state": "Maharashtra",
    "pincode": "400001"
  }' \
  localhost:50051 pb.v2.AddressService/CreateAddress
```

### Step 2: Verify Backward Compatibility
```bash
# Test v1 endpoint (should still work)
grpcurl \
  -d '{
    "user_id": "USER00000001",
    "address_line_1": "123 Main St",
    "city": "Mumbai",
    "postal_code": "400001"
  }' \
  localhost:50051 pb.AddressService/CreateAddress
```

### Step 3: Compare Results
Both should create the address successfully, but v2 will preserve field granularity.

## Timeline

| Phase | Date | Action |
|-------|------|--------|
| **V2 Release** | 2025-11-05 | V2 available, V1 still supported |
| **Client Migration** | 2025-11 - 2026-04 | Clients migrate to v2 |
| **V1 Deprecation Warning** | 2026-04-05 | Warning logs for v1 usage |
| **V1 End of Life** | 2026-05-05 | V1 removed, v2 only |

## Need Help?

- **Technical Issues**: Open an issue in the AAA service repository
- **Migration Questions**: Contact the Backend Platform Team
- **API Documentation**: See `.kiro/specs/address-api-standardization/`
- **Client SDK Updates**: Check language-specific SDK repositories

## Summary Checklist

- [ ] Update proto imports to use `pkg/proto/v2`
- [ ] Add `api-version: v2` header to gRPC calls
- [ ] Split `address_line_1` into separate `house` and `street` fields
- [ ] Move metadata fields (`post_office`, `subdistrict`, `district`) to direct fields
- [ ] Rename `city` → `vtc`, `postal_code` → `pincode`
- [ ] Validate pincode format (6 digits, no leading 0)
- [ ] Validate state name against valid Indian states
- [ ] Test with staging environment
- [ ] Update client documentation
- [ ] Monitor for errors after deployment

## What Happens After Migration?

1. **Better Data Quality**: Validation catches errors early
2. **Richer Analytics**: Separate house/street enables better address parsing
3. **Government Integration**: Ready for PM-KISAN, e-NAM integrations
4. **Future Features**: Address auto-completion, pincode lookup, etc.
