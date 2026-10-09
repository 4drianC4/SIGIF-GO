package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/sigif/sigif-go/internal/modules/company/domain/entity"
)

type CompanyRepository interface {
	List(context.Context) ([]entity.Company, error)
	ExistsByID(context.Context, uuid.UUID) (bool, error)
}
