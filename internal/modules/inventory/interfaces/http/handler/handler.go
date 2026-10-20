package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/inventory/application/command"
	inventorydto "github.com/sigif/sigif-go/internal/modules/inventory/application/dto"
	apphandler "github.com/sigif/sigif-go/internal/modules/inventory/application/handler"
	"github.com/sigif/sigif-go/internal/modules/inventory/application/query"
	"github.com/sigif/sigif-go/internal/modules/inventory/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/inventory/interfaces/http/dtos"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/middleware"
	"github.com/sigif/sigif-go/internal/shared/pagination"
	"github.com/sigif/sigif-go/internal/shared/response"
	"github.com/sigif/sigif-go/internal/shared/validator"
)

type InventoryHTTPHandler struct {
	cmdHandler   *apphandler.InventoryCommandHandler
	queryHandler *apphandler.InventoryQueryHandler
	validator    *validator.Validator
}

func NewInventoryHTTPHandler(
	cmdHandler *apphandler.InventoryCommandHandler,
	queryHandler *apphandler.InventoryQueryHandler,
	val *validator.Validator,
) *InventoryHTTPHandler {
	return &InventoryHTTPHandler{cmdHandler: cmdHandler, queryHandler: queryHandler, validator: val}
}

func (h *InventoryHTTPHandler) RegisterMovement(c *fiber.Ctx) error {
	var req dtos.RegisterMovementRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	req.Normalize()
	if err := h.validator.Validate(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	userID, _ := middleware.UserIDFromContext(c.UserContext())

	result, err := h.cmdHandler.HandleRegisterMovement(c.UserContext(), command.RegisterMovement{
		CompanyID:  companyID,
		ProductID:  uuid.MustParse(req.ProductID),
		BranchID:   req.BranchID,
		LocationID: req.LocationID,
		Type:       entity.MovementType(req.Type),
		Quantity:   *req.Quantity,
		Reason:     req.Reason,
		UserID:     userID,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return response.Created(c, inventorydto.FromRegisterResult(result))
}

func (h *InventoryHTTPHandler) ListBranches(c *fiber.Ctx) error {
	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	branches, err := h.queryHandler.HandleListBranches(c.UserContext(), query.ListBranches{CompanyID: companyID})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Success(c, inventorydto.FromBranches(branches))
}

func (h *InventoryHTTPHandler) ProductStock(c *fiber.Ctx) error {
	var q dtos.ProductStockQuery
	if err := c.QueryParser(&q); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	if err := h.validator.Validate(&q); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	items, err := h.queryHandler.HandleProductStock(c.UserContext(), query.ProductStock{
		CompanyID: companyID,
		ProductID: uuid.MustParse(q.ProductID),
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Success(c, inventorydto.FromBalances(items))
}

func (h *InventoryHTTPHandler) ListMovements(c *fiber.Ctx) error {
	var q dtos.ListMovementsQuery
	if err := c.QueryParser(&q); err != nil {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	if err := h.validator.Validate(&q); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}
	companyID, ok := middleware.CompanyIDFromContext(c.UserContext())
	if !ok {
		return response.Error(c, fiber.StatusBadRequest, sharedErrors.ErrBadRequest)
	}
	page, limit := pagination.Parse(c, 20, 100)
	items, total, err := h.queryHandler.HandleListMovements(c.UserContext(), query.ListMovements{
		CompanyID: companyID,
		ProductID: uuid.MustParse(q.ProductID),
		Offset:    (page - 1) * limit,
		Limit:     limit,
	})
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}
	return response.Paginated(c, inventorydto.FromMovements(items), pagination.New(page, limit, total))
}
