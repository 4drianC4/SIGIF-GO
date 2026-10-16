package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/application/command"
	"github.com/sigif/sigif-go/internal/modules/user/application/dto"
	"github.com/sigif/sigif-go/internal/modules/user/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// ChangeStatus activates or deactivates a role (HU-082-04). The role is never
// deleted: its configuration and permission assignments are preserved.
func (h *RoleHTTPHandler) ChangeStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid role id", 400))
	}

	var req dtos.ChangeRoleStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	var companyID *uuid.UUID
	if callerCompany, ok := middleware.CompanyIDFromContext(c.UserContext()); ok {
		companyID = &callerCompany
	}

	role, err := h.cmdHandler.HandleChangeRoleStatus(c.UserContext(), command.ChangeRoleStatus{
		ID:        id,
		CompanyID: companyID,
		Status:    dtos.RoleStatusFromRequest(req.Status),
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.RoleSummaryFromEntity(role))
}
