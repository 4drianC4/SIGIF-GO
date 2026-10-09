package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/customer/application/command"
	"github.com/sigif/sigif-go/internal/modules/customer/application/dto"
	"github.com/sigif/sigif-go/internal/modules/customer/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/customer/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/response"
)

// Patch partially updates a customer (HU-08-04). The body is strict: it must
// carry at least one editable field and nothing else.
func (h *CustomerHTTPHandler) Patch(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.New(sharedErrors.CodeBadRequest, "invalid customer id", 400))
	}

	req, cleared, details, err := parsePatchBody(c.Body())
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	if err := h.validator.Validate(&req); err != nil {
		var appErr *sharedErrors.AppError
		if !errors.As(err, &appErr) {
			return response.Error(c, fiber.StatusBadRequest, err)
		}
		fieldErrors, ok := appErr.Details.(map[string]string)
		if !ok {
			return response.Error(c, fiber.StatusBadRequest, err)
		}
		for field, message := range fieldErrors {
			details[field] = message
		}
	}
	if len(details) > 0 {
		return response.ValidationError(c, details)
	}

	companyID, err := companyIDFromContext(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	cmd := command.PatchCustomer{
		ID:             id,
		CompanyID:      companyID,
		LegalName:      req.LegalName,
		DocumentNumber: valueOrCleared(req.DocumentNumber, cleared["document_number"]),
		Phone:          valueOrCleared(req.Phone, cleared["phone"]),
		Email:          valueOrCleared(req.Email, cleared["email"]),
		Address:        valueOrCleared(req.Address, cleared["address"]),
	}
	if req.DocumentType != nil {
		documentType := entity.DocumentType(*req.DocumentType)
		cmd.DocumentType = &documentType
	}

	customer, err := h.cmdHandler.HandlePatch(c.UserContext(), cmd)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Success(c, dto.FromEntity(customer))
}

// parsePatchBody decodes a JSON object field by field. It returns the request,
// the optional fields to clear (sent as null or blank) and per-field errors.
func parsePatchBody(body []byte) (dtos.PatchCustomerRequest, map[string]bool, map[string]string, error) {
	var req dtos.PatchCustomerRequest
	if len(bytes.TrimSpace(body)) == 0 {
		return req, nil, nil, sharedErrors.New(sharedErrors.CodeValidation, "at least one editable field is required", 400)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		return req, nil, nil, sharedErrors.New(sharedErrors.CodeBadRequest, "request body must be a JSON object", 400)
	}
	if len(raw) == 0 {
		return req, nil, nil, sharedErrors.New(sharedErrors.CodeValidation, "at least one editable field is required", 400)
	}

	targets := map[string]**string{
		"legal_name":      &req.LegalName,
		"document_type":   &req.DocumentType,
		"document_number": &req.DocumentNumber,
		"phone":           &req.Phone,
		"email":           &req.Email,
		"address":         &req.Address,
	}
	cleared := map[string]bool{}
	details := map[string]string{}

	for key, value := range raw {
		optional, editable := dtos.PatchEditableFields[key]
		if !editable {
			if dtos.PatchReadOnlyFields[key] {
				details[key] = key + " is not editable"
			} else {
				details[key] = key + " is not a known field"
			}
			continue
		}

		var s *string
		if err := json.Unmarshal(value, &s); err != nil {
			details[key] = key + " must be a string"
			continue
		}
		if s == nil || strings.TrimSpace(*s) == "" {
			if optional {
				cleared[key] = true
				continue
			}
			details[key] = key + " cannot be empty"
			continue
		}
		*targets[key] = s
	}

	return req, cleared, details, nil
}

// valueOrCleared maps a cleared optional field to an empty string, which the
// customer entity stores as NULL.
func valueOrCleared(value *string, cleared bool) *string {
	if cleared {
		empty := ""
		return &empty
	}
	return value
}
