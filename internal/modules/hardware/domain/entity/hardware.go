package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type SerialNumberStatus string

const (
	SerialStatusAvailable  SerialNumberStatus = "available"
	SerialStatusReserved   SerialNumberStatus = "reserved"
	SerialStatusSold       SerialNumberStatus = "sold"
	SerialStatusReturned   SerialNumberStatus = "returned"
	SerialStatusDefective  SerialNumberStatus = "defective"
	SerialStatusInRepair   SerialNumberStatus = "in_repair"
)

type SerialNumber struct {
	ID              uuid.UUID           `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID           `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID          `json:"company_id" gorm:"type:uuid;index"`
	ProductID       uuid.UUID           `json:"product_id" gorm:"type:uuid;not null;index"`
	BatchID         *uuid.UUID          `json:"batch_id" gorm:"type:uuid;index"`
	WarehouseID     *uuid.UUID          `json:"warehouse_id" gorm:"type:uuid;index"`
	SerialNumber    string              `json:"serial_number" gorm:"type:varchar(100);uniqueIndex:idx_serial_numbers_tenant_serial;not null"`
	ManufacturerSN  string              `json:"manufacturer_sn" gorm:"type:varchar(100);index"`
	Status          SerialNumberStatus  `json:"status" gorm:"type:varchar(20);default:'available'"`
	PurchaseOrderID *uuid.UUID          `json:"purchase_order_id" gorm:"type:uuid;index"`
	SaleID          *uuid.UUID          `json:"sale_id" gorm:"type:uuid;index"`
	CustomerID      *uuid.UUID          `json:"customer_id" gorm:"type:uuid;index"`
	PurchaseCost    decimal.Decimal     `json:"purchase_cost" gorm:"type:decimal(15,4);default:0"`
	SalePrice       decimal.Decimal     `json:"sale_price" gorm:"type:decimal(15,4);default:0"`
	WarrantyStart   *time.Time          `json:"warranty_start" gorm:"index"`
	WarrantyEnd     *time.Time          `json:"warranty_end" gorm:"index"`
	Notes           string              `json:"notes" gorm:"type:text"`
	ReceivedAt      *time.Time          `json:"received_at"`
	SoldAt          *time.Time          `json:"sold_at"`
	CreatedAt       time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time           `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time          `json:"deleted_at,omitempty" gorm:"index"`
}

func NewSerialNumber(clock clock.Clock, tenantID, productID uuid.UUID, serialNumber string, companyID, batchID, warehouseID *uuid.UUID, manufacturerSN string, purchaseCost decimal.Decimal) *SerialNumber {
	now := clock.NowUTC()
	return &SerialNumber{
		TenantID:       tenantID,
		CompanyID:      companyID,
		ProductID:      productID,
		BatchID:        batchID,
		WarehouseID:    warehouseID,
		SerialNumber:   serialNumber,
		ManufacturerSN: manufacturerSN,
		Status:         SerialStatusAvailable,
		PurchaseCost:   purchaseCost,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *SerialNumber) Reserve(clock clock.Clock) {
	s.Status = SerialStatusReserved
	s.UpdatedAt = clock.NowUTC()
}

func (s *SerialNumber) Sell(clock clock.Clock, saleID, customerID uuid.UUID, salePrice decimal.Decimal) {
	s.Status = SerialStatusSold
	s.SaleID = &saleID
	s.CustomerID = &customerID
	s.SalePrice = salePrice
	now := clock.NowUTC()
	s.SoldAt = &now
	s.UpdatedAt = now
}

func (s *SerialNumber) Return(clock clock.Clock) {
	s.Status = SerialStatusReturned
	s.SaleID = nil
	s.CustomerID = nil
	s.UpdatedAt = clock.NowUTC()
}

func (s *SerialNumber) MarkDefective(clock clock.Clock) {
	s.Status = SerialStatusDefective
	s.UpdatedAt = clock.NowUTC()
}

func (s *SerialNumber) SendToRepair(clock clock.Clock) {
	s.Status = SerialStatusInRepair
	s.UpdatedAt = clock.NowUTC()
}

func (s *SerialNumber) IsAvailable() bool {
	return s.Status == SerialStatusAvailable
}

func (SerialNumber) TableName() string {
	return "serial_numbers"
}

type WarrantyClaim struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	SerialNumberID  uuid.UUID  `json:"serial_number_id" gorm:"type:uuid;not null;index"`
	CustomerID      *uuid.UUID `json:"customer_id" gorm:"type:uuid;index"`
	ClaimNumber     string     `json:"claim_number" gorm:"type:varchar(100);uniqueIndex:idx_warranty_claims_tenant_number;not null"`
	Status          string     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	IssueType       string     `json:"issue_type" gorm:"type:varchar(50)"`
	Description     string     `json:"description" gorm:"type:text"`
	Diagnosis       string     `json:"diagnosis" gorm:"type:text"`
	Resolution      string     `json:"resolution" gorm:"type:text"`
	ClaimDate       time.Time  `json:"claim_date" gorm:"not null;index"`
	ResolvedDate    *time.Time `json:"resolved_date" gorm:"index"`
	ApprovedBy      *uuid.UUID `json:"approved_by" gorm:"type:uuid"`
	ApprovedAt      *time.Time `json:"approved_at"`
	CreatedBy       uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func NewWarrantyClaim(clock clock.Clock, tenantID, serialNumberID, createdBy uuid.UUID, claimNumber, issueType, description string, companyID, customerID *uuid.UUID) *WarrantyClaim {
	now := clock.NowUTC()
	return &WarrantyClaim{
		TenantID:       tenantID,
		CompanyID:      companyID,
		SerialNumberID: serialNumberID,
		CustomerID:     customerID,
		ClaimNumber:    claimNumber,
		Status:         "pending",
		IssueType:      issueType,
		Description:    description,
		ClaimDate:      now,
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (WarrantyClaim) TableName() string {
	return "warranty_claims"
}

type ServiceOrder struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID      `json:"company_id" gorm:"type:uuid;index"`
	SerialNumberID  *uuid.UUID      `json:"serial_number_id" gorm:"type:uuid;index"`
	CustomerID      *uuid.UUID      `json:"customer_id" gorm:"type:uuid;index"`
	OrderNumber     string          `json:"order_number" gorm:"type:varchar(100);uniqueIndex:idx_service_orders_tenant_number;not null"`
	Status          string          `json:"status" gorm:"type:varchar(20);default:'pending'"`
	ServiceType     string          `json:"service_type" gorm:"type:varchar(50)"`
	Description     string          `json:"description" gorm:"type:text"`
	Diagnosis       string          `json:"diagnosis" gorm:"type:text"`
	Resolution      string          `json:"resolution" gorm:"type:text"`
	TechnicianID    *uuid.UUID      `json:"technician_id" gorm:"type:uuid;index"`
	ReceivedDate    time.Time       `json:"received_date" gorm:"not null;index"`
	StartedDate     *time.Time      `json:"started_date" gorm:"index"`
	CompletedDate   *time.Time      `json:"completed_date" gorm:"index"`
	EstimatedCost   decimal.Decimal `json:"estimated_cost" gorm:"type:decimal(15,4);default:0"`
	ActualCost      decimal.Decimal `json:"actual_cost" gorm:"type:decimal(15,4);default:0"`
	PartsCost       decimal.Decimal `json:"parts_cost" gorm:"type:decimal(15,4);default:0"`
	LaborCost       decimal.Decimal `json:"labor_cost" gorm:"type:decimal(15,4);default:0"`
	Notes           string          `json:"notes" gorm:"type:text"`
	CreatedBy       uuid.UUID       `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty" gorm:"index"`
}

func NewServiceOrder(clock clock.Clock, tenantID, createdBy uuid.UUID, orderNumber, serviceType, description string, companyID, serialNumberID, customerID *uuid.UUID, receivedDate time.Time, estimatedCost decimal.Decimal) *ServiceOrder {
	now := clock.NowUTC()
	return &ServiceOrder{
		TenantID:       tenantID,
		CompanyID:      companyID,
		SerialNumberID: serialNumberID,
		CustomerID:     customerID,
		OrderNumber:    orderNumber,
		Status:         "pending",
		ServiceType:    serviceType,
		Description:    description,
		ReceivedDate:   receivedDate,
		EstimatedCost:  estimatedCost,
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (ServiceOrder) TableName() string {
	return "service_orders"
}