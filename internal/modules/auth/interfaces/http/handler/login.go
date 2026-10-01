package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/auth/application/command"
	"github.com/sigif/sigif-go/internal/modules/auth/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

func (h *AuthHTTPHandler) Login(c *fiber.Ctx) error {
	var req dtos.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	tenantID := c.Locals("tenant_id")
	tenantIDUUID, ok := tenantID.(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrInvalidTenant)
	}

	cmd := command.Login{
		TenantID:  tenantIDUUID,
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: c.Get("User-Agent"),
		IPAddress: c.IP(),
	}

	result, err := h.cmdHandler.HandleLogin(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, err)
	}

	return response.Success(c, result)
}
