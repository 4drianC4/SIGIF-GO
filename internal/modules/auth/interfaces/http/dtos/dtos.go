package dtos

import (
	"github.com/sigif/sigif-go/internal/modules/auth/application/dto"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is the login body. The access token travels only in the
// HttpOnly auth cookie; expires_in lets the client know when to log in again.
type LoginResponse struct {
	User      dto.User `json:"user"`
	ExpiresIn int      `json:"expires_in"`
}
