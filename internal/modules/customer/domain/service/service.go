package service

import (
	"github.com/sigif/sigif-go/internal/modules/customer/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type CustomerService struct {
	repo  repository.CustomerRepository
	deps  repository.DependencyChecker
	clock clock.Clock
}

func NewCustomerService(repo repository.CustomerRepository, deps repository.DependencyChecker, clk clock.Clock) *CustomerService {
	return &CustomerService{
		repo:  repo,
		deps:  deps,
		clock: clk,
	}
}
