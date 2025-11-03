package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/entities/models"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	pb "github.com/Kisanlink/aaa-service/v2/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AddressHandler implements address-related gRPC services
type AddressHandler struct {
	pb.UnimplementedAddressServiceServer
	addressService interfaces.AddressService
	logger         *zap.Logger
}

// NewAddressHandler creates a new address handler
func NewAddressHandler(
	addressService interfaces.AddressService,
	logger *zap.Logger,
) *AddressHandler {
	return &AddressHandler{
		addressService: addressService,
		logger:         logger,
	}
}

// modelToProto converts an Address model to proto
// Note: The proto Address fields don't match the actual model structure perfectly
// We map the fields as best as possible
func (h *AddressHandler) modelToProto(addr *models.Address) *pb.Address {
	if addr == nil {
		return nil
	}

	protoAddr := &pb.Address{
		Id: addr.GetID(),
	}

	// Map model fields to proto fields
	// Proto expects: address_line_1, address_line_2, city, state, postal_code, country
	// Model has: House, Street, Landmark, PostOffice, Subdistrict, District, VTC, State, Country, Pincode

	// Combine House and Street into address_line_1
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

	// Set metadata with additional address fields
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

	// Default values for fields not in model
	protoAddr.IsPrimary = false
	protoAddr.IsActive = true

	return protoAddr
}

// protoToModel converts proto Address fields to model
func (h *AddressHandler) protoToModel(req interface{}) *models.Address {
	address := models.NewAddress()

	// Handle different request types
	switch r := req.(type) {
	case *pb.CreateAddressRequest:
		// Split address_line_1 into house and street (simple split by comma)
		if r.AddressLine_1 != "" {
			house := r.AddressLine_1
			address.House = &house
		}

		// Use address_line_2 as landmark
		if r.AddressLine_2 != "" {
			landmark := r.AddressLine_2
			address.Landmark = &landmark
		}

		// Map city to VTC
		if r.City != "" {
			vtc := r.City
			address.VTC = &vtc
		}

		// Map state
		if r.State != "" {
			state := r.State
			address.State = &state
		}

		// Map postal_code to pincode
		if r.PostalCode != "" {
			pincode := r.PostalCode
			address.Pincode = &pincode
		}

		// Map country
		if r.Country != "" {
			country := r.Country
			address.Country = &country
		}

		// Extract additional fields from metadata
		if r.Metadata != nil {
			if postOffice, ok := r.Metadata["post_office"]; ok && postOffice != "" {
				address.PostOffice = &postOffice
			}
			if subdistrict, ok := r.Metadata["subdistrict"]; ok && subdistrict != "" {
				address.Subdistrict = &subdistrict
			}
			if district, ok := r.Metadata["district"]; ok && district != "" {
				address.District = &district
			}
		}

	case *pb.UpdateAddressRequest:
		// Similar mapping for update requests
		if r.AddressLine_1 != "" {
			house := r.AddressLine_1
			address.House = &house
		}
		if r.AddressLine_2 != "" {
			landmark := r.AddressLine_2
			address.Landmark = &landmark
		}
		if r.City != "" {
			vtc := r.City
			address.VTC = &vtc
		}
		if r.State != "" {
			state := r.State
			address.State = &state
		}
		if r.PostalCode != "" {
			pincode := r.PostalCode
			address.Pincode = &pincode
		}
		if r.Country != "" {
			country := r.Country
			address.Country = &country
		}
	}

	return address
}

// CreateAddress creates a new address
func (h *AddressHandler) CreateAddress(ctx context.Context, req *pb.CreateAddressRequest) (*pb.CreateAddressResponse, error) {
	h.logger.Info("gRPC CreateAddress request", zap.String("user_id", req.UserId))

	// Validate request
	if req.UserId == "" {
		h.logger.Warn("CreateAddress called with empty user_id")
		return &pb.CreateAddressResponse{
			StatusCode: 400,
			Message:    "user_id is required",
		}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.AddressLine_1 == "" {
		h.logger.Warn("CreateAddress called with empty address_line_1")
		return &pb.CreateAddressResponse{
			StatusCode: 400,
			Message:    "address_line_1 is required",
		}, status.Error(codes.InvalidArgument, "address_line_1 is required")
	}

	// Convert proto to model
	address := h.protoToModel(req)

	// Note: The current Address model doesn't store UserID directly
	// It should be linked through a separate relationship or stored in metadata
	// For now, we'll store it in the BaseModel's metadata or handle it at repository level

	// Create address
	err := h.addressService.CreateAddress(ctx, address)
	if err != nil {
		h.logger.Error("Failed to create address", zap.String("user_id", req.UserId), zap.Error(err))
		return &pb.CreateAddressResponse{
			StatusCode: 500,
			Message:    "failed to create address: " + err.Error(),
		}, status.Error(codes.Internal, "failed to create address")
	}

	h.logger.Info("Address created successfully", zap.String("id", address.GetID()))

	return &pb.CreateAddressResponse{
		StatusCode: 201,
		Message:    "Address created successfully",
		Address:    h.modelToProto(address),
	}, nil
}

// GetAddress retrieves an address by ID
func (h *AddressHandler) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.GetAddressResponse, error) {
	h.logger.Info("gRPC GetAddress request", zap.String("id", req.Id))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("GetAddress called with empty ID")
		return &pb.GetAddressResponse{
			StatusCode: 400,
			Message:    "address ID is required",
		}, status.Error(codes.InvalidArgument, "address ID is required")
	}

	// Get address
	address, err := h.addressService.GetAddressByID(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get address", zap.String("id", req.Id), zap.Error(err))
		return &pb.GetAddressResponse{
			StatusCode: 404,
			Message:    "address not found",
		}, status.Error(codes.NotFound, "address not found")
	}

	h.logger.Info("Address retrieved successfully", zap.String("id", req.Id))

	return &pb.GetAddressResponse{
		StatusCode: 200,
		Message:    "Address retrieved successfully",
		Address:    h.modelToProto(address),
	}, nil
}

// GetAddressesByUser retrieves all addresses for a user
func (h *AddressHandler) GetAddressesByUser(ctx context.Context, req *pb.GetAddressesByUserRequest) (*pb.GetAddressesByUserResponse, error) {
	h.logger.Info("gRPC GetAddressesByUser request", zap.String("user_id", req.UserId))

	// Validate request
	if req.UserId == "" {
		h.logger.Warn("GetAddressesByUser called with empty user_id")
		return &pb.GetAddressesByUserResponse{
			StatusCode: 400,
			Message:    "user_id is required",
		}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get addresses
	addresses, err := h.addressService.GetAddressesByUserID(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get addresses", zap.String("user_id", req.UserId), zap.Error(err))
		return &pb.GetAddressesByUserResponse{
			StatusCode: 500,
			Message:    "failed to retrieve addresses",
		}, status.Error(codes.Internal, "failed to retrieve addresses")
	}

	// Convert to proto
	protoAddresses := make([]*pb.Address, 0, len(addresses))
	for _, addr := range addresses {
		protoAddresses = append(protoAddresses, h.modelToProto(addr))
	}

	h.logger.Info("Addresses retrieved successfully",
		zap.String("user_id", req.UserId),
		zap.Int("count", len(protoAddresses)))

	return &pb.GetAddressesByUserResponse{
		StatusCode: 200,
		Message:    "Addresses retrieved successfully",
		Addresses:  protoAddresses,
	}, nil
}

// UpdateAddress updates an existing address
func (h *AddressHandler) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.UpdateAddressResponse, error) {
	h.logger.Info("gRPC UpdateAddress request", zap.String("id", req.Id))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("UpdateAddress called with empty ID")
		return &pb.UpdateAddressResponse{
			StatusCode: 400,
			Message:    "address ID is required",
		}, status.Error(codes.InvalidArgument, "address ID is required")
	}

	// Get existing address first
	existingAddr, err := h.addressService.GetAddressByID(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get address for update", zap.String("id", req.Id), zap.Error(err))
		return &pb.UpdateAddressResponse{
			StatusCode: 404,
			Message:    "address not found",
		}, status.Error(codes.NotFound, "address not found")
	}

	// Update fields from request
	updatedAddr := h.protoToModel(req)

	// Merge non-empty fields
	if updatedAddr.House != nil {
		existingAddr.House = updatedAddr.House
	}
	if updatedAddr.Street != nil {
		existingAddr.Street = updatedAddr.Street
	}
	if updatedAddr.Landmark != nil {
		existingAddr.Landmark = updatedAddr.Landmark
	}
	if updatedAddr.VTC != nil {
		existingAddr.VTC = updatedAddr.VTC
	}
	if updatedAddr.State != nil {
		existingAddr.State = updatedAddr.State
	}
	if updatedAddr.Pincode != nil {
		existingAddr.Pincode = updatedAddr.Pincode
	}
	if updatedAddr.Country != nil {
		existingAddr.Country = updatedAddr.Country
	}
	if updatedAddr.District != nil {
		existingAddr.District = updatedAddr.District
	}
	if updatedAddr.PostOffice != nil {
		existingAddr.PostOffice = updatedAddr.PostOffice
	}
	if updatedAddr.Subdistrict != nil {
		existingAddr.Subdistrict = updatedAddr.Subdistrict
	}

	// Update address
	err = h.addressService.UpdateAddress(ctx, existingAddr)
	if err != nil {
		h.logger.Error("Failed to update address", zap.String("id", req.Id), zap.Error(err))
		return &pb.UpdateAddressResponse{
			StatusCode: 500,
			Message:    "failed to update address",
		}, status.Error(codes.Internal, "failed to update address")
	}

	h.logger.Info("Address updated successfully", zap.String("id", req.Id))

	return &pb.UpdateAddressResponse{
		StatusCode: 200,
		Message:    "Address updated successfully",
		Address:    h.modelToProto(existingAddr),
	}, nil
}

// DeleteAddress deletes an address
func (h *AddressHandler) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*pb.DeleteAddressResponse, error) {
	h.logger.Info("gRPC DeleteAddress request", zap.String("id", req.Id), zap.Bool("soft_delete", req.SoftDelete))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("DeleteAddress called with empty ID")
		return &pb.DeleteAddressResponse{
			StatusCode: 400,
			Message:    "address ID is required",
		}, status.Error(codes.InvalidArgument, "address ID is required")
	}

	// For now, we only support hard delete (the service interface doesn't have soft delete)
	err := h.addressService.DeleteAddress(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to delete address", zap.String("id", req.Id), zap.Error(err))
		return &pb.DeleteAddressResponse{
			StatusCode: 500,
			Message:    "failed to delete address",
		}, status.Error(codes.Internal, "failed to delete address")
	}

	h.logger.Info("Address deleted successfully", zap.String("id", req.Id))

	return &pb.DeleteAddressResponse{
		StatusCode: 200,
		Message:    "Address deleted successfully",
	}, nil
}

// ListAddresses lists addresses with pagination and filters
func (h *AddressHandler) ListAddresses(ctx context.Context, req *pb.ListAddressesRequest) (*pb.ListAddressesResponse, error) {
	h.logger.Info("gRPC ListAddresses request",
		zap.Int32("page", req.Page),
		zap.Int32("page_size", req.PageSize),
		zap.String("user_id", req.UserId))

	// User ID is required for this implementation
	if req.UserId == "" {
		h.logger.Warn("ListAddresses called without user_id")
		return &pb.ListAddressesResponse{
			StatusCode: 400,
			Message:    "user_id is required",
		}, status.Error(codes.InvalidArgument, "user_id is required for listing addresses")
	}

	// Get addresses by user ID
	addresses, err := h.addressService.GetAddressesByUserID(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to list addresses", zap.String("user_id", req.UserId), zap.Error(err))
		return &pb.ListAddressesResponse{
			StatusCode: 500,
			Message:    "failed to list addresses",
		}, status.Error(codes.Internal, "failed to list addresses")
	}

	// Convert to proto
	protoAddresses := make([]*pb.Address, 0, len(addresses))
	for _, addr := range addresses {
		protoAddresses = append(protoAddresses, h.modelToProto(addr))
	}

	h.logger.Info("Addresses listed successfully", zap.Int("count", len(protoAddresses)))

	return &pb.ListAddressesResponse{
		StatusCode: 200,
		Message:    "Addresses retrieved successfully",
		Addresses:  protoAddresses,
		TotalCount: int32(len(protoAddresses)),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}
