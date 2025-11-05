package version

import (
	"google.golang.org/grpc/metadata"
)

// Detector handles API version detection from gRPC metadata
type Detector struct {
	defaultVersion string
}

// NewDetector creates a new version detector with default version set to v1
func NewDetector() *Detector {
	return &Detector{
		defaultVersion: "v1",
	}
}

// GetAPIVersion extracts the API version from gRPC metadata
// Checks the following headers in order:
// 1. api-version
// 2. x-api-version
// Returns defaultVersion (v1) if no version header is found
func (d *Detector) GetAPIVersion(md metadata.MD) string {
	// Check api-version header
	if versions := md.Get("api-version"); len(versions) > 0 {
		return versions[0]
	}

	// Check x-api-version header
	if versions := md.Get("x-api-version"); len(versions) > 0 {
		return versions[0]
	}

	// Return default version for backward compatibility
	return d.defaultVersion
}

// SupportsIndianFormat returns true if the client supports Indian address format (v2)
// v1 clients use international format (address_line_1, city, postal_code)
// v2 clients use Indian format (house, street, vtc, pincode)
func (d *Detector) SupportsIndianFormat(md metadata.MD) bool {
	version := d.GetAPIVersion(md)
	return version == "v2" || version == "2.0" || version == "2"
}

// IsV2 returns true if the API version is v2
func (d *Detector) IsV2(md metadata.MD) bool {
	return d.SupportsIndianFormat(md)
}

// IsV1 returns true if the API version is v1
func (d *Detector) IsV1(md metadata.MD) bool {
	return !d.IsV2(md)
}
