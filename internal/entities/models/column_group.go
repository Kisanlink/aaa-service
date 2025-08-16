package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
	"gorm.io/gorm"
)

// ColumnGroup represents a named group of columns for authorization
type ColumnGroup struct {
	*base.BaseModel
	Name           string  `json:"name" gorm:"size:100;not null"`
	Description    string  `json:"description" gorm:"type:text"`
	Table          string  `json:"table" gorm:"size:100;not null;column:table_name"`
	OrganizationID string  `json:"organization_id" gorm:"type:varchar(255);not null"`
	IsActive       bool    `json:"is_active" gorm:"default:true"`
	Metadata       *string `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization  *Organization       `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
	ColumnMembers []ColumnGroupMember `json:"column_members" gorm:"foreignKey:ColumnGroupID"`
	ColumnSets    []ColumnSet         `json:"column_sets" gorm:"foreignKey:ColumnGroupID"`
}

// ColumnGroupMember represents a column that belongs to a column group
type ColumnGroupMember struct {
	*base.BaseModel
	ColumnGroupID  string `json:"column_group_id" gorm:"type:varchar(255);not null"`
	ColumnName     string `json:"column_name" gorm:"size:100;not null"`
	ColumnPosition int    `json:"column_position" gorm:"not null"` // Position in the table
	IsActive       bool   `json:"is_active" gorm:"default:true"`

	// Relationships
	ColumnGroup *ColumnGroup `json:"column_group" gorm:"foreignKey:ColumnGroupID;references:ID"`
}

// BitSet represents a bitmap for efficient column set operations
type BitSet []byte

// Scan implements the Scanner interface for database reads
func (bs *BitSet) Scan(value interface{}) error {
	if value == nil {
		*bs = make(BitSet, 0)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*bs = BitSet(v)
		return nil
	default:
		return errors.New("cannot scan bitset from database")
	}
}

// Value implements the Valuer interface for database writes
func (bs BitSet) Value() (driver.Value, error) {
	if bs == nil {
		return nil, nil
	}
	return []byte(bs), nil
}

// ColumnSet represents an optimized bitmap of allowed columns
type ColumnSet struct {
	*base.BaseModel
	Name           string  `json:"name" gorm:"size:100;not null"`
	Table          string  `json:"table" gorm:"size:100;not null;column:table_name"`
	ColumnGroupID  *string `json:"column_group_id" gorm:"type:varchar(255)"` // Optional reference to a group
	Bitmap         BitSet  `json:"bitmap" gorm:"type:bytea;not null"`        // Bitmap representation
	ColumnCount    int     `json:"column_count" gorm:"not null"`             // Total columns in the bitmap
	OrganizationID string  `json:"organization_id" gorm:"type:varchar(255);not null"`
	IsActive       bool    `json:"is_active" gorm:"default:true"`

	// Relationships
	Organization *Organization `json:"organization" gorm:"foreignKey:OrganizationID;references:ID"`
	ColumnGroup  *ColumnGroup  `json:"column_group" gorm:"foreignKey:ColumnGroupID;references:ID"`
}

func (cg *ColumnGroup) TableName() string {
	return "column_groups"
}

func (cgm *ColumnGroupMember) TableName() string {
	return "column_group_members"
}

func (cs *ColumnSet) TableName() string {
	return "column_sets"
}

// NewColumnGroup creates a new ColumnGroup instance
func NewColumnGroup(name, description, tableName, organizationID string) *ColumnGroup {
	return &ColumnGroup{
		BaseModel:      base.NewBaseModel("COLGROUP", hash.Small),
		Name:           name,
		Description:    description,
		Table:          tableName,
		OrganizationID: organizationID,
		IsActive:       true,
	}
}

// NewColumnGroupMember creates a new ColumnGroupMember instance
func NewColumnGroupMember(columnGroupID, columnName string, position int) *ColumnGroupMember {
	return &ColumnGroupMember{
		BaseModel:      base.NewBaseModel("COLGRMEM", hash.Small),
		ColumnGroupID:  columnGroupID,
		ColumnName:     columnName,
		ColumnPosition: position,
		IsActive:       true,
	}
}

// NewColumnSet creates a new ColumnSet instance
func NewColumnSet(name, tableName string, columnCount int, organizationID string) *ColumnSet {
	// Initialize bitmap with enough bytes to hold all columns
	bitmapSize := (columnCount + 7) / 8 // Round up to nearest byte
	bitmap := make(BitSet, bitmapSize)

	return &ColumnSet{
		BaseModel:      base.NewBaseModel("COLSET", hash.Small),
		Name:           name,
		Table:          tableName,
		Bitmap:         bitmap,
		ColumnCount:    columnCount,
		OrganizationID: organizationID,
		IsActive:       true,
	}
}

// BeforeCreate hooks
func (cg *ColumnGroup) BeforeCreate() error {
	return cg.BaseModel.BeforeCreate()
}

func (cgm *ColumnGroupMember) BeforeCreate() error {
	return cgm.BaseModel.BeforeCreate()
}

func (cs *ColumnSet) BeforeCreate() error {
	return cs.BaseModel.BeforeCreate()
}

// GORM Hooks
func (cg *ColumnGroup) BeforeCreateGORM(tx *gorm.DB) error {
	return cg.BeforeCreate()
}

func (cgm *ColumnGroupMember) BeforeCreateGORM(tx *gorm.DB) error {
	return cgm.BeforeCreate()
}

func (cs *ColumnSet) BeforeCreateGORM(tx *gorm.DB) error {
	return cs.BeforeCreate()
}

// Helper methods
func (cg *ColumnGroup) GetTableIdentifier() string   { return "COL" }
func (cg *ColumnGroup) GetTableSize() hash.TableSize { return hash.Medium }

// Explicit method implementations to satisfy linter
func (cg *ColumnGroup) GetID() string   { return cg.BaseModel.GetID() }
func (cg *ColumnGroup) SetID(id string) { cg.BaseModel.SetID(id) }

func (cgm *ColumnGroupMember) GetTableIdentifier() string   { return "CGM" }
func (cgm *ColumnGroupMember) GetTableSize() hash.TableSize { return hash.Medium }

func (cs *ColumnSet) GetTableIdentifier() string   { return "cls" }
func (cs *ColumnSet) GetTableSize() hash.TableSize { return hash.Small }

// SetColumn sets a column bit to allowed (1) in the bitmap
func (cs *ColumnSet) SetColumn(position int) error {
	if position < 0 || position >= cs.ColumnCount {
		return errors.New("column position out of range")
	}

	byteIndex := position / 8
	bitIndex := uint(position % 8)

	if byteIndex >= len(cs.Bitmap) {
		return errors.New("bitmap too small for column position")
	}

	cs.Bitmap[byteIndex] |= (1 << bitIndex)
	return nil
}

// UnsetColumn sets a column bit to not allowed (0) in the bitmap
func (cs *ColumnSet) UnsetColumn(position int) error {
	if position < 0 || position >= cs.ColumnCount {
		return errors.New("column position out of range")
	}

	byteIndex := position / 8
	bitIndex := uint(position % 8)

	if byteIndex >= len(cs.Bitmap) {
		return errors.New("bitmap too small for column position")
	}

	cs.Bitmap[byteIndex] &^= (1 << bitIndex)
	return nil
}

// IsColumnAllowed checks if a column is allowed in the bitmap
func (cs *ColumnSet) IsColumnAllowed(position int) (bool, error) {
	if position < 0 || position >= cs.ColumnCount {
		return false, errors.New("column position out of range")
	}

	byteIndex := position / 8
	bitIndex := uint(position % 8)

	if byteIndex >= len(cs.Bitmap) {
		return false, errors.New("bitmap too small for column position")
	}

	return (cs.Bitmap[byteIndex] & (1 << bitIndex)) != 0, nil
}

// Union performs a bitwise OR with another column set
func (cs *ColumnSet) Union(other *ColumnSet) error {
	if cs.Table != other.Table {
		return errors.New("cannot union column sets from different tables")
	}

	if cs.ColumnCount != other.ColumnCount {
		return errors.New("column counts do not match")
	}

	if len(cs.Bitmap) != len(other.Bitmap) {
		return errors.New("bitmap sizes do not match")
	}

	for i := 0; i < len(cs.Bitmap); i++ {
		cs.Bitmap[i] |= other.Bitmap[i]
	}

	return nil
}

// Intersect performs a bitwise AND with another column set
func (cs *ColumnSet) Intersect(other *ColumnSet) error {
	if cs.Table != other.Table {
		return errors.New("cannot intersect column sets from different tables")
	}

	if cs.ColumnCount != other.ColumnCount {
		return errors.New("column counts do not match")
	}

	if len(cs.Bitmap) != len(other.Bitmap) {
		return errors.New("bitmap sizes do not match")
	}

	for i := 0; i < len(cs.Bitmap); i++ {
		cs.Bitmap[i] &= other.Bitmap[i]
	}

	return nil
}

// IsSubsetOf checks if this column set is a subset of another
func (cs *ColumnSet) IsSubsetOf(other *ColumnSet) (bool, error) {
	if cs.Table != other.Table {
		return false, errors.New("cannot compare column sets from different tables")
	}

	if cs.ColumnCount != other.ColumnCount {
		return false, errors.New("column counts do not match")
	}

	if len(cs.Bitmap) != len(other.Bitmap) {
		return false, errors.New("bitmap sizes do not match")
	}

	for i := 0; i < len(cs.Bitmap); i++ {
		if (cs.Bitmap[i] & ^other.Bitmap[i]) != 0 {
			return false, nil
		}
	}

	return true, nil
}

// GetAllowedColumns returns a list of allowed column positions
func (cs *ColumnSet) GetAllowedColumns() []int {
	allowed := make([]int, 0)

	for pos := 0; pos < cs.ColumnCount; pos++ {
		if isAllowed, err := cs.IsColumnAllowed(pos); err == nil && isAllowed {
			allowed = append(allowed, pos)
		}
	}

	return allowed
}

// ToBigInt converts the bitmap to a big.Int for efficient operations
func (cs *ColumnSet) ToBigInt() *big.Int {
	return new(big.Int).SetBytes(cs.Bitmap)
}

// FromBigInt sets the bitmap from a big.Int
func (cs *ColumnSet) FromBigInt(bigInt *big.Int) {
	bytes := bigInt.Bytes()
	if len(bytes) > len(cs.Bitmap) {
		// If the big.Int is larger, we need to resize our bitmap
		cs.Bitmap = make(BitSet, len(bytes))
	}
	copy(cs.Bitmap, bytes)
}

// MarshalJSON implements json.Marshaler
func (cs *ColumnSet) MarshalJSON() ([]byte, error) {
	type Alias ColumnSet
	return json.Marshal(&struct {
		Bitmap string `json:"bitmap"`
		*Alias
	}{
		Bitmap: cs.ToBigInt().Text(16), // Hex representation
		Alias:  (*Alias)(cs),
	})
}

// AddColumn adds a column to the column group
func (cg *ColumnGroup) AddColumn(columnName string, position int) *ColumnGroupMember {
	member := NewColumnGroupMember(cg.GetID(), columnName, position)
	cg.ColumnMembers = append(cg.ColumnMembers, *member)
	return member
}
