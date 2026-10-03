package service

import (
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

// CustomerService expone los casos de uso de dominio de clientes.
// Cada operación está en su propio archivo (create.go, get.go, etc.).
type CustomerService struct {
	repo  repository.CustomerRepository
	clock clock.Clock
}

func NewCustomerService(repo repository.CustomerRepository, clk clock.Clock) *CustomerService {
	return &CustomerService{
		repo:  repo,
		clock: clk,
	}
}
