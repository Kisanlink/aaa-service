package responses

// UserInfo represents minimal user information in auth responses
type UserInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	CountryCode string `json:"country_code"`
	IsValidated bool   `json:"is_validated"`
}

// LoginResponse represents the response for a successful login
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	User         *UserInfo `json:"user"`
	Message      string    `json:"message"`
}

// GetType returns the response type
func (r *LoginResponse) GetType() string {
	return "login"
}

// IsSuccess returns whether the response indicates success
func (r *LoginResponse) IsSuccess() bool {
	return r.AccessToken != ""
}

// RegisterResponse represents the response for a successful registration
type RegisterResponse struct {
	User    *UserInfo `json:"user"`
	Message string    `json:"message"`
}

// GetType returns the response type
func (r *RegisterResponse) GetType() string {
	return "register"
}

// IsSuccess returns whether the response indicates success
func (r *RegisterResponse) IsSuccess() bool {
	return r.User != nil && r.User.ID != ""
}

// RefreshTokenResponse represents the response for a token refresh
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Message      string `json:"message"`
}

// GetType returns the response type
func (r *RefreshTokenResponse) GetType() string {
	return "refresh_token"
}

// IsSuccess returns whether the response indicates success
func (r *RefreshTokenResponse) IsSuccess() bool {
	return r.AccessToken != ""
}

// ForgotPasswordResponse represents the response for a forgot password request
type ForgotPasswordResponse struct {
	Message string `json:"message"`
	SentTo  string `json:"sent_to,omitempty"`
}

// GetType returns the response type
func (r *ForgotPasswordResponse) GetType() string {
	return "forgot_password"
}

// IsSuccess returns whether the response indicates success
func (r *ForgotPasswordResponse) IsSuccess() bool {
	return r.Message != ""
}

// ResetPasswordResponse represents the response for a reset password request
type ResetPasswordResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetType returns the response type
func (r *ResetPasswordResponse) GetType() string {
	return "reset_password"
}

// IsSuccess returns whether the response indicates success
func (r *ResetPasswordResponse) IsSuccess() bool {
	return r.Success
}

// LogoutResponse represents the response for a logout request
type LogoutResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// GetType returns the response type
func (r *LogoutResponse) GetType() string {
	return "logout"
}

// IsSuccess returns whether the response indicates success
func (r *LogoutResponse) IsSuccess() bool {
	return r.Success
}

// TokenValidationResponse represents the response for token validation
type TokenValidationResponse struct {
	Valid   bool                   `json:"valid"`
	User    *UserInfo              `json:"user,omitempty"`
	Claims  map[string]interface{} `json:"claims,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// GetType returns the response type
func (r *TokenValidationResponse) GetType() string {
	return "token_validation"
}

// IsSuccess returns whether the response indicates success
func (r *TokenValidationResponse) IsSuccess() bool {
	return r.Valid
}
