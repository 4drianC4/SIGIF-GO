package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type PromotionType string

const (
	PromotionTypePercentage PromotionType = "percentage"
	PromotionTypeFixed      PromotionType = "fixed"
	PromotionTypeBOGO       PromotionType = "bogo"
	PromotionTypeBundle     PromotionType = "bundle"
	PromotionTypeLoyalty    PromotionType = "loyalty"
)

type Promotion struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID      `json:"company_id" gorm:"type:uuid;index"`
	Name            string          `json:"name" gorm:"type:varchar(255);not null"`
	Description     string          `json:"description" gorm:"type:text"`
	Type            PromotionType   `json:"type" gorm:"type:varchar(20);not null"`
	Code            string          `json:"code" gorm:"type:varchar(50);uniqueIndex:idx_promotions_tenant_code;not null"`
	Value           decimal.Decimal `json:"value" gorm:"type:decimal(15,4);not null"`
	MinAmount       decimal.Decimal `json:"min_amount" gorm:"type:decimal(15,4);default:0"`
	MaxDiscount     decimal.Decimal `json:"max_discount" gorm:"type:decimal(15,4);default:0"`
	StartDate       time.Time       `json:"start_date" gorm:"not null;index"`
	EndDate         time.Time       `json:"end_date" gorm:"not null;index"`
	UsageLimit      int             `json:"usage_limit" gorm:"default:0"`
	UsageCount      int             `json:"usage_count" gorm:"default:0"`
	PerCustomerLimit int            `json:"per_customer_limit" gorm:"default:0"`
	IsActive        bool            `json:"is_active" gorm:"default:true"`
	ApplicableProducts string        `json:"applicable_products" gorm:"type:jsonb"`
	ApplicableCategories string     `json:"applicable_categories" gorm:"type:jsonb"`
	ApplicableCustomers string      `json:"applicable_customers" gorm:"type:jsonb"`
	CreatedBy       uuid.UUID       `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty" gorm:"index"`
}

func NewPromotion(clock clock.Clock, tenantID, createdBy uuid.UUID, name, code string, promoType PromotionType, value decimal.Decimal, startDate, endDate time.Time, companyID *uuid.UUID) *Promotion {
	now := clock.NowUTC()
	return &Promotion{
		TenantID:           tenantID,
		CompanyID:          companyID,
		Name:               name,
		Code:               code,
		Type:               promoType,
		Value:              value,
		StartDate:          startDate,
		EndDate:            endDate,
		IsActive:           true,
		CreatedBy:          createdBy,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func (p *Promotion) IsValid() bool {
	now := time.Now()
	return p.IsActive && now.After(p.StartDate) && now.Before(p.EndDate) && (p.UsageLimit == 0 || p.UsageCount < p.UsageLimit)
}

func (Promotion) TableName() string {
	return "promotions"
}

type LoyaltyProgram struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID      `json:"company_id" gorm:"type:uuid;index"`
	Name            string          `json:"name" gorm:"type:varchar(255);not null"`
	Description     string          `json:"description" gorm:"type:text"`
	PointsPerAmount decimal.Decimal `json:"points_per_amount" gorm:"type:decimal(10,4);default:1"`
	PointsValue     decimal.Decimal `json:"points_value" gorm:"type:decimal(10,4);default:0.01"`
	MinPointsRedeem int             `json:"min_points_redeem" gorm:"default:100"`
	PointsExpiryDays int            `json:"points_expiry_days" gorm:"default:365"`
	Tiers           string          `json:"tiers" gorm:"type:jsonb"`
	IsActive        bool            `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func NewLoyaltyProgram(clock clock.Clock, tenantID uuid.UUID, name string, pointsPerAmount, pointsValue decimal.Decimal, companyID *uuid.UUID) *LoyaltyProgram {
	now := clock.NowUTC()
	return &LoyaltyProgram{
		TenantID:         tenantID,
		CompanyID:        companyID,
		Name:             name,
		PointsPerAmount:  pointsPerAmount,
		PointsValue:      pointsValue,
		MinPointsRedeem:  100,
		PointsExpiryDays: 365,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (LoyaltyProgram) TableName() string {
	return "loyalty_programs"
}

type CustomerLoyalty struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CustomerID      uuid.UUID       `json:"customer_id" gorm:"type:uuid;not null;uniqueIndex:idx_customer_loyalty_tenant_customer;not null"`
	ProgramID       uuid.UUID       `json:"program_id" gorm:"type:uuid;not null;index"`
	CurrentPoints   int             `json:"current_points" gorm:"default:0"`
	LifetimePoints  int             `json:"lifetime_points" gorm:"default:0"`
	CurrentTier     string          `json:"current_tier" gorm:"type:varchar(50)"`
	LastActivityAt  *time.Time      `json:"last_activity_at" gorm:"index"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func NewCustomerLoyalty(clock clock.Clock, tenantID, customerID, programID uuid.UUID) *CustomerLoyalty {
	now := clock.NowUTC()
	return &CustomerLoyalty{
		TenantID:       tenantID,
		CustomerID:     customerID,
		ProgramID:      programID,
		CurrentPoints:  0,
		LifetimePoints: 0,
		CurrentTier:    "bronze",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (CustomerLoyalty) TableName() string {
	return "customer_loyalty"
}

type LoyaltyTransaction struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CustomerID      uuid.UUID       `json:"customer_id" gorm:"type:uuid;not null;index"`
	ProgramID       uuid.UUID       `json:"program_id" gorm:"type:uuid;not null;index"`
	SaleID          *uuid.UUID      `json:"sale_id" gorm:"type:uuid;index"`
	TransactionType string          `json:"transaction_type" gorm:"type:varchar(20);not null"`
	Points          int             `json:"points" gorm:"not null"`
	BalanceBefore   int             `json:"balance_before" gorm:"not null"`
	BalanceAfter    int             `json:"balance_after" gorm:"not null"`
	Description     string          `json:"description" gorm:"type:text"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

func NewLoyaltyTransaction(clock clock.Clock, tenantID, customerID, programID uuid.UUID, transactionType string, points, balanceBefore, balanceAfter int, saleID *uuid.UUID, description string) *LoyaltyTransaction {
	now := clock.NowUTC()
	return &LoyaltyTransaction{
		TenantID:         tenantID,
		CustomerID:       customerID,
		ProgramID:        programID,
		SaleID:           saleID,
		TransactionType:  transactionType,
		Points:           points,
		BalanceBefore:    balanceBefore,
		BalanceAfter:     balanceAfter,
		Description:      description,
		CreatedAt:        now,
	}
}

func (LoyaltyTransaction) TableName() string {
	return "loyalty_transactions"
}