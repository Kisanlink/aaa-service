package grpc_server

import (
	"context"

	"github.com/Kisanlink/aaa-service/v2/internal/grpc_server/converters"
	"github.com/Kisanlink/aaa-service/v2/internal/grpc_server/version"
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	pbv2 "github.com/Kisanlink/aaa-service/v2/pkg/proto/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AddressHandlerV2 implements address-related gRPC services with v2 Indian format
type AddressHandlerV2 struct {
	pbv2.UnimplementedAddressServiceServer
	addressService  interfaces.AddressService
	logger          *zap.Logger
	converter       *converters.AddressConverter
	versionDetector *version.Detector
}

// NewAddressHandlerV2 creates a new v2 address handler with Indian format support
func NewAddressHandlerV2(
	addressService interfaces.AddressService,
	logger *zap.Logger,
) *AddressHandlerV2 {
	return &AddressHandlerV2{
		addressService:  addressService,
		logger:          logger,
		converter:       converters.NewAddressConverter(),
		versionDetector: version.NewDetector(),
	}
}

// CreateAddress creates a new address with Indian format
func (h *AddressHandlerV2) CreateAddress(ctx context.Context, req *pbv2.CreateAddressRequest) (*pbv2.CreateAddressResponse, error) {
	h.logger.Info("gRPC v2 CreateAddress request", zap.String("user_id", req.UserId))

	// Validate request
	if req.UserId == "" {
		h.logger.Warn("CreateAddress called with empty user_id")
		return &pbv2.CreateAddressResponse{
			StatusCode: 400,
			Message:    "user_id is required",
		}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Validate at least one address field is provided
	if req.House == "" && req.Street == "" && req.Vtc == "" {
		h.logger.Warn("CreateAddress called without any address fields")
		return &pbv2.CreateAddressResponse{
			StatusCode: 400,
			Message:    "at least one address field (house, street, or vtc) is required",
		}, status.Error(codes.InvalidArgument, "address information is required")
	}

	// Convert proto to model
	address := h.converter.V2ProtoToModel(req)

	// Create address
	err := h.addressService.CreateAddress(ctx, address)
	if err != nil {
		h.logger.Error("Failed to create address", zap.String("user_id", req.UserId), zap.Error(err))
		return &pbv2.CreateAddressResponse{
			StatusCode: 500,
			Message:    "failed to create address: " + err.Error(),
		}, status.Error(codes.Internal, "failed to create address")
	}

	h.logger.Info("Address created successfully with v2 format", zap.String("id", address.GetID()))

	return &pbv2.CreateAddressResponse{
		StatusCode: 201,
		Message:    "Address created successfully",
		Address:    h.converter.ModelToV2Proto(address),
	}, nil
}

// GetAddress retrieves an address by ID
func (h *AddressHandlerV2) GetAddress(ctx context.Context, req *pbv2.GetAddressRequest) (*pbv2.GetAddressResponse, error) {
	h.logger.Info("gRPC v2 GetAddress request", zap.String("id", req.Id))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("GetAddress called with empty ID")
		return &pbv2.GetAddressResponse{
			StatusCode: 400,
			Message:    "address ID is required",
		}, status.Error(codes.InvalidArgument, "address ID is required")
	}

	// Get address
	address, err := h.addressService.GetAddressByID(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get address", zap.String("id", req.Id), zap.Error(err))
		return &pbv2.GetAddressResponse{
			StatusCode: 404,
			Message:    "address not found",
		}, status.Error(codes.NotFound, "address not found")
	}

	h.logger.Info("Address retrieved successfully with v2 format", zap.String("id", req.Id))

	return &pbv2.GetAddressResponse{
		StatusCode: 200,
		Message:    "Address retrieved successfully",
		Address:    h.converter.ModelToV2Proto(address),
	}, nil
}

// GetAddressesByUser retrieves all addresses for a user
func (h *AddressHandlerV2) GetAddressesByUser(ctx context.Context, req *pbv2.GetAddressesByUserRequest) (*pbv2.GetAddressesByUserResponse, error) {
	h.logger.Info("gRPC v2 GetAddressesByUser request", zap.String("user_id", req.UserId))

	// Validate request
	if req.UserId == "" {
		h.logger.Warn("GetAddressesByUser called with empty user_id")
		return &pbv2.GetAddressesByUserResponse{
			StatusCode: 400,
			Message:    "user_id is required",
		}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get addresses
	addresses, err := h.addressService.GetAddressesByUserID(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get addresses", zap.String("user_id", req.UserId), zap.Error(err))
		return &pbv2.GetAddressesByUserResponse{
			StatusCode: 500,
			Message:    "failed to retrieve addresses",
		}, status.Error(codes.Internal, "failed to retrieve addresses")
	}

	// Convert to proto
	protoAddresses := make([]*pbv2.Address, 0, len(addresses))
	for _, addr := range addresses {
		protoAddresses = append(protoAddresses, h.converter.ModelToV2Proto(addr))
	}

	h.logger.Info("Addresses retrieved successfully with v2 format",
		zap.String("user_id", req.UserId),
		zap.Int("count", len(protoAddresses)))

	return &pbv2.GetAddressesByUserResponse{
		StatusCode: 200,
		Message:    "Addresses retrieved successfully",
		Addresses:  protoAddresses,
	}, nil
}

// UpdateAddress updates an existing address
func (h *AddressHandlerV2) UpdateAddress(ctx context.Context, req *pbv2.UpdateAddressRequest) (*pbv2.UpdateAddressResponse, error) {
	h.logger.Info("gRPC v2 UpdateAddress request", zap.String("id", req.Id))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("UpdateAddress called with empty ID")
		return &pbv2.UpdateAddressResponse{
			StatusCode: 400,
			Message:    "address ID is required",
		}, status.Error(codes.InvalidArgument, "address ID is required")
	}

	// Get existing address first
	existingAddr, err := h.addressService.GetAddressByID(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get address for update", zap.String("id", req.Id), zap.Error(err))
		return &pbv2.UpdateAddressResponse{
			StatusCode: 404,
			Message:    "address not found",
		}, status.Error(codes.NotFound, "address not found")
	}

	// Update fields from request
	updatedAddr := h.converter.V2UpdateProtoToModel(req)

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
	if updatedAddr.PostOffice != nil {
		existingAddr.PostOffice = updatedAddr.PostOffice
	}
	if updatedAddr.Subdistrict != nil {
		existingAddr.Subdistrict = updatedAddr.Subdistrict
	}
	if updatedAddr.District != nil {
		existingAddr.District = updatedAddr.District
	}
	if updatedAddr.VTC != nil {
		existingAddr.VTC = updatedAddr.VTC
	}
	if updatedAddr.State != nil {
		existingAddr.State = updatedAddr.State
	}
	if updatedAddr.Country != nil {
		existingAddr.Country = updatedAddr.Country
	}
	if updatedAddr.Pincode != nil {
		existingAddr.Pincode = updatedAddr.Pincode
	}

	// Rebuild full address
	existingAddr.BuildFullAddress()

	// Update address
	err = h.addressService.UpdateAddress(ctx, existingAddr)
	if err != nil {
		h.logger.Error("Failed to update address", zap.String("id", req.Id), zap.Error(err))
		return &pbv2.UpdateAddressResponse{
			StatusCode: 500,
			Message:    "failed to update address",
		}, status.Error(codes.Internal, "failed to update address")
	}

	h.logger.Info("Address updated successfully with v2 format", zap.String("id", req.Id))

	return &pbv2.UpdateAddressResponse{
		StatusCode: 200,
		Message:    "Address updated successfully",
		Address:    h.converter.ModelToV2Proto(existingAddr),
	}, nil
}

// DeleteAddress deletes an address
func (h *AddressHandlerV2) DeleteAddress(ctx context.Context, req *pbv2.DeleteAddressRequest) (*pbv2.DeleteAddressResponse, error) {
	h.logger.Info("gRPC v2 DeleteAddress request", zap.String("id", req.Id), zap.Bool("soft_delete", req.SoftDelete))

	// Validate request
	if req.Id == "" {
		h.logger.Warn("DeleteAddress called with empty ID")
		return &pbv2.DeleteAddressResponse{
			StatusCode: 400,
			Message:    "address ID is required",
		}, status.Error(codes.InvalidArgument, "address ID is required")
	}

	// For now, we only support hard delete (the service interface doesn't have soft delete)
	err := h.addressService.DeleteAddress(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to delete address", zap.String("id", req.Id), zap.Error(err))
		return &pbv2.DeleteAddressResponse{
			StatusCode: 500,
			Message:    "failed to delete address",
		}, status.Error(codes.Internal, "failed to delete address")
	}

	h.logger.Info("Address deleted successfully with v2 format", zap.String("id", req.Id))

	return &pbv2.DeleteAddressResponse{
		StatusCode: 200,
		Message:    "Address deleted successfully",
	}, nil
}

// ListAddresses lists addresses with pagination and filters
func (h *AddressHandlerV2) ListAddresses(ctx context.Context, req *pbv2.ListAddressesRequest) (*pbv2.ListAddressesResponse, error) {
	h.logger.Info("gRPC v2 ListAddresses request",
		zap.Int32("page", req.Page),
		zap.Int32("page_size", req.PageSize),
		zap.String("user_id", req.UserId))

	// User ID is required for this implementation
	if req.UserId == "" {
		h.logger.Warn("ListAddresses called without user_id")
		return &pbv2.ListAddressesResponse{
			StatusCode: 400,
			Message:    "user_id is required",
		}, status.Error(codes.InvalidArgument, "user_id is required for listing addresses")
	}

	// Get addresses by user ID
	addresses, err := h.addressService.GetAddressesByUserID(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to list addresses", zap.String("user_id", req.UserId), zap.Error(err))
		return &pbv2.ListAddressesResponse{
			StatusCode: 500,
			Message:    "failed to list addresses",
		}, status.Error(codes.Internal, "failed to list addresses")
	}

	// Convert to proto
	protoAddresses := make([]*pbv2.Address, 0, len(addresses))
	for _, addr := range addresses {
		protoAddresses = append(protoAddresses, h.converter.ModelToV2Proto(addr))
	}

	h.logger.Info("Addresses listed successfully with v2 format", zap.Int("count", len(protoAddresses)))

	return &pbv2.ListAddressesResponse{
		StatusCode: 200,
		Message:    "Addresses retrieved successfully",
		Addresses:  protoAddresses,
		TotalCount: int32(len(protoAddresses)),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}
