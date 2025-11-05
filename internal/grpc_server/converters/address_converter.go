package converters

import (
	"strings"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	pbv2 "github.com/Kisanlink/aaa-service/v2/pkg/proto/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AddressConverter handles conversion between domain models and proto messages
type AddressConverter struct{}

// NewAddressConverter creates a new address converter
func NewAddressConverter() *AddressConverter {
	return &AddressConverter{}
}

// ModelToV2Proto converts domain model to v2 proto (direct mapping, no data loss)
func (c *AddressConverter) ModelToV2Proto(addr *models.Address) *pbv2.Address {
	if addr == nil {
		return nil
	}

	proto := &pbv2.Address{
		Id:     addr.GetID(),
		UserId: addr.UserID,
	}

	// Direct field mapping - no data loss
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

	// Add timestamps from BaseModel
	if addr.BaseModel != nil {
		if createdAt := addr.BaseModel.GetCreatedAt(); !createdAt.IsZero() {
			proto.CreatedAt = timestamppb.New(createdAt)
		}
		if updatedAt := addr.BaseModel.GetUpdatedAt(); !updatedAt.IsZero() {
			proto.UpdatedAt = timestamppb.New(updatedAt)
		}
	}

	// Default values
	proto.IsPrimary = false
	proto.IsActive = true

	return proto
}

// ModelToV1Proto converts domain model to v1 proto (with data loss for backward compatibility)
func (c *AddressConverter) ModelToV1Proto(addr *models.Address) *pb.Address {
	if addr == nil {
		return nil
	}

	protoAddr := &pb.Address{
		Id:     addr.GetID(),
		UserId: addr.UserID,
	}

	// Lossy mapping: Combine House and Street into address_line_1
	if addr.House != nil && addr.Street != nil {
		combined := *addr.House + ", " + *addr.Street
		protoAddr.AddressLine_1 = combined
	} else if addr.House != nil {
		protoAddr.AddressLine_1 = *addr.House
	} else if addr.Street != nil {
		protoAddr.AddressLine_1 = *addr.Street
	}

	// Use Landmark as address_line_2
	if addr.Landmark != nil {
		protoAddr.AddressLine_2 = *addr.Landmark
	}

	// Use VTC (Village/Town/City) as city
	if addr.VTC != nil {
		protoAddr.City = *addr.VTC
	}

	// Map State directly
	if addr.State != nil {
		protoAddr.State = *addr.State
	}

	// Map Pincode to postal_code
	if addr.Pincode != nil {
		protoAddr.PostalCode = *addr.Pincode
	}

	// Map Country directly
	if addr.Country != nil {
		protoAddr.Country = *addr.Country
	}

	// Set metadata with additional address fields not in v1 proto
	protoAddr.Metadata = make(map[string]string)
	if addr.PostOffice != nil {
		protoAddr.Metadata["post_office"] = *addr.PostOffice
	}
	if addr.Subdistrict != nil {
		protoAddr.Metadata["subdistrict"] = *addr.Subdistrict
	}
	if addr.District != nil {
		protoAddr.Metadata["district"] = *addr.District
	}
	if addr.FullAddress != nil {
		protoAddr.Metadata["full_address"] = *addr.FullAddress
	}

	// Add timestamps from BaseModel
	if addr.BaseModel != nil {
		if createdAt := addr.BaseModel.GetCreatedAt(); !createdAt.IsZero() {
			protoAddr.CreatedAt = timestamppb.New(createdAt)
		}
		if updatedAt := addr.BaseModel.GetUpdatedAt(); !updatedAt.IsZero() {
			protoAddr.UpdatedAt = timestamppb.New(updatedAt)
		}
	}

	// Default values
	protoAddr.IsPrimary = false
	protoAddr.IsActive = true

	return protoAddr
}

// V2ProtoToModel converts v2 proto to domain model (direct mapping)
func (c *AddressConverter) V2ProtoToModel(proto *pbv2.CreateAddressRequest) *models.Address {
	if proto == nil {
		return nil
	}

	addr := models.NewAddress()

	// Set user ID
	addr.UserID = proto.UserId

	// Direct field mapping
	if proto.House != "" {
		addr.House = &proto.House
	}
	if proto.Street != "" {
		addr.Street = &proto.Street
	}
	if proto.Landmark != "" {
		addr.Landmark = &proto.Landmark
	}
	if proto.PostOffice != "" {
		addr.PostOffice = &proto.PostOffice
	}
	if proto.Subdistrict != "" {
		addr.Subdistrict = &proto.Subdistrict
	}
	if proto.District != "" {
		addr.District = &proto.District
	}
	if proto.Vtc != "" {
		addr.VTC = &proto.Vtc
	}
	if proto.State != "" {
		addr.State = &proto.State
	}
	if proto.Country != "" {
		addr.Country = &proto.Country
	}
	if proto.Pincode != "" {
		addr.Pincode = &proto.Pincode
	}

	// Build full address from components
	addr.BuildFullAddress()

	return addr
}

// V2UpdateProtoToModel converts v2 update proto to domain model
func (c *AddressConverter) V2UpdateProtoToModel(proto *pbv2.UpdateAddressRequest) *models.Address {
	if proto == nil {
		return nil
	}

	addr := models.NewAddress()
	addr.SetID(proto.Id)

	// Direct field mapping
	if proto.House != "" {
		addr.House = &proto.House
	}
	if proto.Street != "" {
		addr.Street = &proto.Street
	}
	if proto.Landmark != "" {
		addr.Landmark = &proto.Landmark
	}
	if proto.PostOffice != "" {
		addr.PostOffice = &proto.PostOffice
	}
	if proto.Subdistrict != "" {
		addr.Subdistrict = &proto.Subdistrict
	}
	if proto.District != "" {
		addr.District = &proto.District
	}
	if proto.Vtc != "" {
		addr.VTC = &proto.Vtc
	}
	if proto.State != "" {
		addr.State = &proto.State
	}
	if proto.Country != "" {
		addr.Country = &proto.Country
	}
	if proto.Pincode != "" {
		addr.Pincode = &proto.Pincode
	}

	return addr
}

// V1ProtoToModel converts v1 proto to domain model (with best-effort reconstruction)
func (c *AddressConverter) V1ProtoToModel(proto *pb.CreateAddressRequest) *models.Address {
	if proto == nil {
		return nil
	}

	addr := models.NewAddress()

	// Set user ID
	addr.UserID = proto.UserId

	// Best-effort split of address_line_1 into house and street
	// If address_line_1 contains a comma, split it
	if proto.AddressLine_1 != "" {
		parts := strings.SplitN(proto.AddressLine_1, ",", 2)
		house := strings.TrimSpace(parts[0])
		addr.House = &house

		if len(parts) > 1 {
			street := strings.TrimSpace(parts[1])
			addr.Street = &street
		}
	}

	// Use address_line_2 as landmark
	if proto.AddressLine_2 != "" {
		landmark := proto.AddressLine_2
		addr.Landmark = &landmark
	}

	// Map city to VTC
	if proto.City != "" {
		vtc := proto.City
		addr.VTC = &vtc
	}

	// Map state
	if proto.State != "" {
		state := proto.State
		addr.State = &state
	}

	// Map postal_code to pincode
	if proto.PostalCode != "" {
		pincode := proto.PostalCode
		addr.Pincode = &pincode
	}

	// Map country
	if proto.Country != "" {
		country := proto.Country
		addr.Country = &country
	}

	// Extract additional fields from metadata
	if proto.Metadata != nil {
		if postOffice, ok := proto.Metadata["post_office"]; ok && postOffice != "" {
			addr.PostOffice = &postOffice
		}
		if subdistrict, ok := proto.Metadata["subdistrict"]; ok && subdistrict != "" {
			addr.Subdistrict = &subdistrict
		}
		if district, ok := proto.Metadata["district"]; ok && district != "" {
			addr.District = &district
		}
	}

	// Build full address from components
	addr.BuildFullAddress()

	return addr
}

// V1UpdateProtoToModel converts v1 update proto to domain model
func (c *AddressConverter) V1UpdateProtoToModel(proto *pb.UpdateAddressRequest) *models.Address {
	if proto == nil {
		return nil
	}

	addr := models.NewAddress()
	addr.SetID(proto.Id)

	// Best-effort split of address_line_1 into house and street
	if proto.AddressLine_1 != "" {
		parts := strings.SplitN(proto.AddressLine_1, ",", 2)
		house := strings.TrimSpace(parts[0])
		addr.House = &house

		if len(parts) > 1 {
			street := strings.TrimSpace(parts[1])
			addr.Street = &street
		}
	}

	// Use address_line_2 as landmark
	if proto.AddressLine_2 != "" {
		landmark := proto.AddressLine_2
		addr.Landmark = &landmark
	}

	// Map city to VTC
	if proto.City != "" {
		vtc := proto.City
		addr.VTC = &vtc
	}

	// Map state
	if proto.State != "" {
		state := proto.State
		addr.State = &state
	}

	// Map postal_code to pincode
	if proto.PostalCode != "" {
		pincode := proto.PostalCode
		addr.Pincode = &pincode
	}

	// Map country
	if proto.Country != "" {
		country := proto.Country
		addr.Country = &country
	}

	// Extract additional fields from metadata
	if proto.Metadata != nil {
		if postOffice, ok := proto.Metadata["post_office"]; ok && postOffice != "" {
			addr.PostOffice = &postOffice
		}
		if subdistrict, ok := proto.Metadata["subdistrict"]; ok && subdistrict != "" {
			addr.Subdistrict = &subdistrict
		}
		if district, ok := proto.Metadata["district"]; ok && district != "" {
			addr.District = &district
		}
	}

	return addr
}
