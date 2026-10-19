package handler

import "github.com/sigif/sigif-go/internal/modules/supplier/domain/service"

type SupplierCommandHandler struct{ service *service.SupplierService }
type SupplierQueryHandler struct{ service *service.SupplierService }

func NewSupplierCommandHandler(s *service.SupplierService) *SupplierCommandHandler {
	return &SupplierCommandHandler{service: s}
}
func NewSupplierQueryHandler(s *service.SupplierService) *SupplierQueryHandler {
	return &SupplierQueryHandler{service: s}
}
