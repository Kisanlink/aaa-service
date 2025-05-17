package model

type User struct {
	Base
	Username      string  `json:"username" gorm:"unique" validate:"required,username"`
	Password      string  `json:"password" validate:"required,min=8,max=128"`
	IsValidated   bool    `json:"isValidate" validate:"default:false"`
	AadhaarNumber *string `json:"aadhaar_number" gorm:"type:varchar(12)"`
	Status        *string `json:"status" gorm:"type:string"`
	Name          *string `json:"name" gorm:"type:string"`
	CareOf        *string `json:"care_of" gorm:"type:string"`
	DateOfBirth   *string `json:"date_of_birth" gorm:"type:string"`
	Photo         *string `json:"photo" gorm:"type:string"`
	EmailHash     *string `json:"email_hash" gorm:"type:string"`
	ShareCode     *string `json:"share_code" gorm:"type:string"`
	YearOfBirth   *string `json:"year_of_birth" gorm:"type:string"`
	MobileNumber  uint64  `json:"mobile_number" gorm:"type:bigint"`
	CountryCode   *string `json:"country_code" gorm:"type:varchar(10);default:'+91'"`
	Message       *string `json:"message" gorm:"type:string"`
	AddressID     *string `json:"address_id"`
	//Address       Address    `gorm:"foreignKey:ID;references:AddressID"`
	Address Address    `gorm:"foreignKey:AddressID;references:ID"`
	Roles   []UserRole `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// for credits
	Tokens int `json:"tokens" gorm:"default:1000"`
}

type Address struct {
	Base
	House       *string `json:"house" gorm:"type:string"`
	Street      *string `json:"street" gorm:"type:string"`
	Landmark    *string `json:"landmark" gorm:"type:string"`
	PostOffice  *string `json:"post_office" gorm:"type:string"`
	Subdistrict *string `json:"subdistrict" gorm:"type:string"`
	District    *string `json:"district" gorm:"type:string"`
	VTC         *string `json:"vtc" gorm:"type:string"`
	State       *string `json:"state" gorm:"type:string"`
	Country     *string `json:"country" gorm:"type:string"`
	Pincode     *string `json:"pincode" gorm:"type:string"`
	FullAddress *string `json:"full_address" gorm:"type:string"`
	// Remove the User field as it is not needed
}

type UserRole struct {
	Base
	UserID   string `gorm:"type:uuid"`
	User     User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoleID   string `gorm:"type:uuid"`
	Role     *Role  `gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsActive bool   `json:"is_active" validate:"default:true"`
}
