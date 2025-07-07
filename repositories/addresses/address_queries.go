package addresses

import (
	"context"

	"aaa-service/entities/models"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// ListAll retrieves all addresses
func (r *AddressRepository) ListAll(ctx context.Context, limit, offset int) ([]models.Address, error) {
	return r.List(ctx, []db.Filter{}, limit, offset)
}

// SearchByPincode searches addresses by pincode
func (r *AddressRepository) SearchByPincode(ctx context.Context, pincode string, limit, offset int) ([]models.Address, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("pincode", db.FilterOpEqual, pincode),
	}

	return r.List(ctx, filters, limit, offset)
}

// SearchByDistrict searches addresses by district
func (r *AddressRepository) SearchByDistrict(ctx context.Context, district string, limit, offset int) ([]models.Address, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("district", db.FilterOpContains, district),
	}

	return r.List(ctx, filters, limit, offset)
}

// SearchByState searches addresses by state
func (r *AddressRepository) SearchByState(ctx context.Context, state string, limit, offset int) ([]models.Address, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("state", db.FilterOpContains, state),
	}

	return r.List(ctx, filters, limit, offset)
}

// SearchByVTC searches addresses by Village/Town/City
func (r *AddressRepository) SearchByVTC(ctx context.Context, vtc string, limit, offset int) ([]models.Address, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("vtc", db.FilterOpContains, vtc),
	}

	return r.List(ctx, filters, limit, offset)
}

// SearchByKeyword searches addresses by keyword in various fields
func (r *AddressRepository) SearchByKeyword(ctx context.Context, keyword string, limit, offset int) ([]models.Address, error) {
	filters := []db.Filter{
		r.dbManager.BuildFilter("house", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("street", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("landmark", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("post_office", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("subdistrict", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("district", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("vtc", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("state", db.FilterOpContains, keyword),
		r.dbManager.BuildFilter("country", db.FilterOpContains, keyword),
	}

	return r.List(ctx, filters, limit, offset)
}
