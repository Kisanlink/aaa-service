# Address API Standardization - Implementation Plan

## Immediate Actions Required

### Phase 1: Proto Definition Update (Priority: HIGH)

#### 1.1 Create New Proto Version
**File**: `pkg/proto/v2/address.proto`

```protobuf
syntax = "proto3";

package pb.v2;

option go_package = "github.com/Kisanlink/aaa-service/v2/pkg/proto/v2;pbv2";

import "google/protobuf/timestamp.proto";

// Indian Address model aligned with domain
message Address {
    string id = 1;
    string user_id = 2;
    string type = 3; // HOME, WORK, OTHER

    // Core Indian address fields
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

    // System fields
    bool is_primary = 15;
    bool is_active = 16;
    map<string, string> metadata = 17;
    google.protobuf.Timestamp created_at = 18;
    google.protobuf.Timestamp updated_at = 19;
}

// Create Address Request
message CreateAddressRequest {
    string user_id = 1;
    string type = 2;

    // Indian format fields
    string house = 3;
    string street = 4;
    string landmark = 5;
    string post_office = 6;
    string subdistrict = 7;
    string district = 8;
    string vtc = 9;
    string state = 10;
    string country = 11;
    string pincode = 12;

    bool is_primary = 13;
    map<string, string> metadata = 14;
}
```

#### 1.2 Generate Proto Code
```bash
# Command to generate Go code from proto
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pkg/proto/v2/address.proto
```

### Phase 2: Compatibility Layer Implementation

#### 2.1 Create Version Detector
**File**: `internal/grpc_server/version/detector.go`

```go
package version

import (
    "google.golang.org/grpc/metadata"
)

type Detector struct {
    defaultVersion string
}

func NewDetector() *Detector {
    return &Detector{
        defaultVersion: "v1",
    }
}

func (d *Detector) GetAPIVersion(md metadata.MD) string {
    if versions := md.Get("api-version"); len(versions) > 0 {
        return versions[0]
    }
    if versions := md.Get("x-api-version"); len(versions) > 0 {
        return versions[0]
    }
    return d.defaultVersion
}

func (d *Detector) SupportsIndianFormat(md metadata.MD) bool {
    version := d.GetAPIVersion(md)
    return version == "v2" || version == "2.0"
}
```

#### 2.2 Create Address Converter
**File**: `internal/grpc_server/converters/address_converter.go`

```go
package converters

import (
    pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
    pbv2 "github.com/Kisanlink/aaa-service/v2/pkg/proto/v2"
    "github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

type AddressConverter struct{}

func NewAddressConverter() *AddressConverter {
    return &AddressConverter{}
}

// ModelToV2Proto converts domain model to v2 proto (direct mapping)
func (c *AddressConverter) ModelToV2Proto(addr *models.Address) *pbv2.Address {
    if addr == nil {
        return nil
    }

    proto := &pbv2.Address{
        Id:     addr.GetID(),
        UserId: addr.UserID, // Need to add UserID field to model
    }

    // Direct field mapping
    if addr.House != nil {
        proto.House = *addr.House
    }
    if addr.Street != nil {
        proto.Street = *addr.Street
    }
    if addr.Landmark != nil {
        proto.Landmark = *addr.Landmark
    }
    if addr.PostOffice != nil {
        proto.PostOffice = *addr.PostOffice
    }
    if addr.Subdistrict != nil {
        proto.Subdistrict = *addr.Subdistrict
    }
    if addr.District != nil {
        proto.District = *addr.District
    }
    if addr.VTC != nil {
        proto.Vtc = *addr.VTC
    }
    if addr.State != nil {
        proto.State = *addr.State
    }
    if addr.Country != nil {
        proto.Country = *addr.Country
    }
    if addr.Pincode != nil {
        proto.Pincode = *addr.Pincode
    }
    if addr.FullAddress != nil {
        proto.FullAddress = *addr.FullAddress
    }

    return proto
}

// ModelToV1Proto converts domain model to v1 proto (with data loss)
func (c *AddressConverter) ModelToV1Proto(addr *models.Address) *pb.Address {
    // Implementation for backward compatibility
    // This is the existing modelToProto logic from address_handler.go
}

// V2ProtoToModel converts v2 proto to domain model
func (c *AddressConverter) V2ProtoToModel(proto *pbv2.CreateAddressRequest) *models.Address {
    addr := models.NewAddress()

    if proto.House != "" {
        addr.House = &proto.House
    }
    if proto.Street != "" {
        addr.Street = &proto.Street
    }
    // ... continue for all fields

    return addr
}
```

### Phase 3: Handler Updates

#### 3.1 Update gRPC Handler
**File**: `internal/grpc_server/address_handler_v2.go`

```go
package grpc_server

import (
    "context"
    pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
    pbv2 "github.com/Kisanlink/aaa-service/v2/pkg/proto/v2"
    "google.golang.org/grpc/metadata"
)

// Enhanced handler supporting both versions
func (h *AddressHandler) CreateAddressV2(ctx context.Context, req interface{}) (interface{}, error) {
    md, _ := metadata.FromIncomingContext(ctx)
    version := h.versionDetector.GetAPIVersion(md)

    switch version {
    case "v2", "2.0":
        return h.handleV2CreateAddress(ctx, req.(*pbv2.CreateAddressRequest))
    default:
        return h.handleV1CreateAddress(ctx, req.(*pb.CreateAddressRequest))
    }
}

func (h *AddressHandler) handleV2CreateAddress(ctx context.Context, req *pbv2.CreateAddressRequest) (*pbv2.CreateAddressResponse, error) {
    // Direct handling with Indian format
    address := h.converter.V2ProtoToModel(req)

    // Validate Indian address format
    if err := h.validator.ValidateIndianAddress(address); err != nil {
        return &pbv2.CreateAddressResponse{
            StatusCode: 400,
            Message:    "Invalid address: " + err.Error(),
        }, nil
    }

    // Create address
    if err := h.addressService.CreateAddress(ctx, address); err != nil {
        return &pbv2.CreateAddressResponse{
            StatusCode: 500,
            Message:    "Failed to create address",
        }, nil
    }

    return &pbv2.CreateAddressResponse{
        StatusCode: 201,
        Message:    "Address created successfully",
        Address:    h.converter.ModelToV2Proto(address),
    }, nil
}
```

### Phase 4: Validation Enhancement

#### 4.1 Indian Address Validator
**File**: `internal/validators/indian_address_validator.go`

```go
package validators

import (
    "fmt"
    "regexp"
    "github.com/Kisanlink/aaa-service/v2/internal/entities/models"
)

type IndianAddressValidator struct {
    pincodeRegex *regexp.Regexp
    validStates  map[string]bool
}

func NewIndianAddressValidator() *IndianAddressValidator {
    return &IndianAddressValidator{
        pincodeRegex: regexp.MustCompile(`^[1-9][0-9]{5}$`),
        validStates: map[string]bool{
            "Andhra Pradesh": true,
            "Maharashtra": true,
            "Karnataka": true,
            // ... add all Indian states
        },
    }
}

func (v *IndianAddressValidator) ValidateIndianAddress(addr *models.Address) error {
    // Validate required fields
    if addr.District == nil || *addr.District == "" {
        return fmt.Errorf("district is required")
    }

    if addr.State == nil || *addr.State == "" {
        return fmt.Errorf("state is required")
    }

    // Validate pincode format
    if addr.Pincode != nil && *addr.Pincode != "" {
        if !v.pincodeRegex.MatchString(*addr.Pincode) {
            return fmt.Errorf("invalid pincode format, must be 6 digits")
        }
    }

    // Validate state
    if addr.State != nil && !v.validStates[*addr.State] {
        return fmt.Errorf("invalid Indian state: %s", *addr.State)
    }

    return nil
}
```

### Phase 5: Testing Strategy

#### 5.1 Unit Tests for Converter
**File**: `internal/grpc_server/converters/address_converter_test.go`

```go
package converters_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestModelToV2Proto(t *testing.T) {
    tests := []struct {
        name     string
        input    *models.Address
        expected *pbv2.Address
    }{
        {
            name: "complete indian address",
            input: &models.Address{
                House: ptr("123"),
                Street: ptr("MG Road"),
                Landmark: ptr("Near City Mall"),
                PostOffice: ptr("Andheri"),
                Subdistrict: ptr("Andheri West"),
                District: ptr("Mumbai"),
                VTC: ptr("Mumbai"),
                State: ptr("Maharashtra"),
                Country: ptr("India"),
                Pincode: ptr("400053"),
            },
            expected: &pbv2.Address{
                House: "123",
                Street: "MG Road",
                // ... rest of fields
            },
        },
    }

    converter := NewAddressConverter()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := converter.ModelToV2Proto(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### 5.2 Integration Tests
**File**: `internal/grpc_server/address_handler_integration_test.go`

```go
func TestAddressHandlerV2Integration(t *testing.T) {
    // Test v1 client compatibility
    t.Run("v1 client backward compatibility", func(t *testing.T) {
        // Create v1 request
        // Verify it still works
    })

    // Test v2 client with Indian format
    t.Run("v2 client indian format", func(t *testing.T) {
        // Create v2 request with Indian fields
        // Verify proper handling
    })
}
```

### Phase 6: Database Optimizations

#### 6.1 Add Indexes for Performance
**File**: `migrations/20240101_add_address_indexes.sql`

```sql
-- Optimize address queries
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_addresses_pincode ON addresses(pincode);
CREATE INDEX IF NOT EXISTS idx_addresses_district_state ON addresses(district, state);
CREATE INDEX IF NOT EXISTS idx_addresses_created_at ON addresses(created_at);

-- Add check constraints for data integrity
ALTER TABLE addresses
    ADD CONSTRAINT chk_pincode_format
    CHECK (pincode ~ '^[1-9][0-9]{5}$' OR pincode IS NULL);
```

### Phase 7: Monitoring and Rollback

#### 7.1 Add Metrics Collection
**File**: `internal/metrics/address_metrics.go`

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    AddressAPIVersion = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "address_api_version_total",
            Help: "Total number of address API calls by version",
        },
        []string{"version", "method"},
    )

    AddressConversionErrors = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "address_conversion_errors_total",
            Help: "Total number of address format conversion errors",
        },
    )
)

func init() {
    prometheus.MustRegister(AddressAPIVersion)
    prometheus.MustRegister(AddressConversionErrors)
}
```

#### 7.2 Rollback Plan
```bash
# If issues arise, rollback strategy:

# 1. Feature flag to disable v2
export AAA_ENABLE_ADDRESS_V2=false

# 2. Revert to previous deployment
kubectl rollout undo deployment/aaa-service

# 3. Database rollback (if needed)
psql -d aaa_db -f migrations/rollback/remove_address_indexes.sql
```

## Security Checklist

- [ ] Input validation for all Indian address fields
- [ ] SQL injection prevention in address queries
- [ ] XSS prevention in address display
- [ ] Rate limiting on address creation/updates
- [ ] Authorization checks for address modifications
- [ ] Audit logging for all address operations
- [ ] PII data encryption at rest
- [ ] GDPR compliance for address data

## Performance Targets

- Address creation: < 100ms p95
- Address retrieval: < 50ms p95
- Bulk address operations: < 500ms for 100 addresses
- Validation overhead: < 10ms per address

## Rollout Schedule

| Week | Task | Owner | Status |
|------|------|-------|--------|
| 1 | Proto v2 definition & generation | Backend Team | Pending |
| 1 | Compatibility layer implementation | Backend Team | Pending |
| 2 | Validator implementation | Backend Team | Pending |
| 2 | Unit & integration tests | QA Team | Pending |
| 3 | Staging deployment | DevOps | Pending |
| 3 | Client SDK updates | Client Teams | Pending |
| 4 | Production rollout (10%) | DevOps | Pending |
| 5 | Production rollout (50%) | DevOps | Pending |
| 6 | Production rollout (100%) | DevOps | Pending |

## Success Criteria

1. Zero downtime during migration
2. 100% backward compatibility for existing clients
3. < 1% error rate increase
4. No performance degradation
5. All Indian address fields properly stored
6. Successful validation of Indian addresses

## Next Steps

1. Review and approve design with stakeholders
2. Set up feature flags for gradual rollout
3. Create client migration guide
4. Schedule client team meetings
5. Begin implementation of Phase 1
