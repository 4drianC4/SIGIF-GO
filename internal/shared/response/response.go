package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/pagination"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    any         `json:"data,omitempty"`
	Meta    any         `json:"meta,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func Success(c *fiber.Ctx, data any, meta ...any) error {
	resp := APIResponse{
		Success: true,
		Data:    data,
	}
	if len(meta) > 0 {
		resp.Meta = meta[0]
	}
	return c.JSON(resp)
}

func Created(c *fiber.Ctx, data any) error {
	resp := APIResponse{
		Success: true,
		Data:    data,
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func Paginated(c *fiber.Ctx, data any, page *pagination.Page) error {
	return Success(c, data, page)
}

func Cursor(c *fiber.Ctx, data any, cursor *pagination.Cursor) error {
	return Success(c, data, cursor)
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Error(c *fiber.Ctx, statusCode int, err error) error {
	var appErr *errors.AppError
	if e, ok := err.(*errors.AppError); ok {
		appErr = e
	} else {
		appErr = errors.ErrInternal
	}

	resp := APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}
	return c.Status(statusCode).JSON(resp)
}

func ValidationError(c *fiber.Ctx, details map[string]string) error {
	return Error(c, fiber.StatusBadRequest, errors.New(errors.CodeValidation, "validation failed", fiber.StatusBadRequest).WithDetails(details))
}