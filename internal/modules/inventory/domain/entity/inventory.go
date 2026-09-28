package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type StockMovementType string

const (
	StockMovementTypeIn        StockMovementType = "in"
	StockMovementTypeOut       StockMovementType = "out"
	StockMovementTypeAdjustment StockMovementType = "adjustment"
	StockMovementTypeTransfer  StockMovementType = "transfer"
	StockMovementTypeReturn    StockMovementType = "return"
	StockMovementTypeLoss      StockMovementType = "loss"
	StockMovementTypeDamage    StockMovementType = "damage"
	StockMovementTypeExpired   StockMovementType = "expired"
)

type StockMovement struct {
	ID              uuid.UUID           `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID           `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID          `json:"company_id" gorm:"type:uuid;index"`
	ProductID       uuid.UUID           `json:"product_id" gorm:"type:uuid;not null;index"`
	WarehouseID     *uuid.UUID          `json:"warehouse_id" gorm:"type:uuid;index"`
	BatchID         *uuid.UUID          `json:"batch_id" gorm:"type:uuid;index"`
	SerialNumber    string              `json:"serial_number" gorm:"type:varchar(100);index"`
	Type            StockMovementType   `json:"type" gorm:"type:varchar(20);not null"`
	ReferenceType   string              `json:"reference_type" gorm:"type:varchar(50)"`
	ReferenceID     *uuid.UUID          `json:"reference_id" gorm:"type:uuid;index"`
	Quantity        decimal.Decimal     `json:"quantity" gorm:"type:decimal(15,4);not null"`
	PreviousStock   decimal.Decimal     `json:"previous_stock" gorm:"type:decimal(15,4);not null"`
	NewStock        decimal.Decimal     `json:"new_stock" gorm:"type:decimal(15,4);not null"`
	UnitCost        decimal.Decimal     `json:"unit_cost" gorm:"type:decimal(15,4);default:0"`
	TotalCost       decimal.Decimal     `json:"total_cost" gorm:"type:decimal(15,4);default:0"`
	Notes           string              `json:"notes" gorm:"type:text"`
	CreatedBy       uuid.UUID           `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt       time.Time           `json:"created_at" gorm:"autoCreateTime"`
}

func NewStockMovement(clock clock.Clock, tenantID, productID, createdBy uuid.UUID, movementType StockMovementType, quantity, previousStock, newStock decimal.Decimal, companyID, warehouseID, batchID *uuid.UUID, serialNumber, referenceType string, referenceID *uuid.UUID, unitCost decimal.Decimal, notes string) *StockMovement {
	now := clock.NowUTC()
	return &StockMovement{
		TenantID:      tenantID,
		CompanyID:     companyID,
		ProductID:     productID,
		WarehouseID:   warehouseID,
		BatchID:       batchID,
		SerialNumber:  serialNumber,
		Type:          movementType,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
		Quantity:      quantity,
		PreviousStock: previousStock,
		NewStock:      newStock,
		UnitCost:      unitCost,
		TotalCost:     unitCost.Mul(quantity.Abs()),
		Notes:         notes,
		CreatedBy:     createdBy,
		CreatedAt:     now,
	}
}

func (StockMovement) TableName() string {
	return "stock_movements"
}

type Warehouse struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID   *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Code        string     `json:"code" gorm:"type:varchar(50);uniqueIndex:idx_warehouses_tenant_code;not null"`
	Address     string     `json:"address" gorm:"type:text"`
	City        string     `json:"city" gorm:"type:varchar(100)"`
	State       string     `json:"state" gorm:"type:varchar(100)"`
	Country     string     `json:"country" gorm:"type:varchar(100)"`
	PostalCode  string     `json:"postal_code" gorm:"type:varchar(20)"`
	IsDefault   bool       `json:"is_default" gorm:"default:false"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	Settings    WarehouseSettings `json:"settings" gorm:"type:jsonb"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type WarehouseSettings struct {
	AllowNegativeStock bool `json:"allow_negative_stock" gorm:"default:false"`
	RequireBatch       bool `json:"require_batch" gorm:"default:false"`
	RequireSerial      bool `json:"require_serial" gorm:"default:false"`
}

func NewWarehouse(clock clock.Clock, tenantID uuid.UUID, name, code string, companyID *uuid.UUID, isDefault bool) *Warehouse {
	now := clock.NowUTC()
	return &Warehouse{
		TenantID:  tenantID,
		CompanyID: companyID,
		Name:      name,
		Code:      code,
		IsDefault: isDefault,
		IsActive:  true,
		Settings: WarehouseSettings{
			AllowNegativeStock: false,
			RequireBatch:       false,
			RequireSerial:      false,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (Warehouse) TableName() string {
	return "warehouses"
}

type Batch struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID   `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID  `json:"company_id" gorm:"type:uuid;index"`
	ProductID       uuid.UUID   `json:"product_id" gorm:"type:uuid;not null;index"`
	WarehouseID     *uuid.UUID  `json:"warehouse_id" gorm:"type:uuid;index"`
	BatchNumber     string      `json:"batch_number" gorm:"type:varchar(100);not null"`
	ManufactureDate *time.Time  `json:"manufacture_date" gorm:"index"`
	ExpiryDate      *time.Time  `json:"expiry_date" gorm:"index"`
	Quantity        decimal.Decimal `json:"quantity" gorm:"type:decimal(15,4);default:0"`
	UnitCost        decimal.Decimal `json:"unit_cost" gorm:"type:decimal(15,4);default:0"`
	SupplierID      *uuid.UUID  `json:"supplier_id" gorm:"type:uuid;index"`
	Notes           string      `json:"notes" gorm:"type:text"`
	IsActive        bool        `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time  `json:"deleted_at,omitempty" gorm:"index"`
}

func NewBatch(clock clock.Clock, tenantID, productID uuid.UUID, batchNumber string, companyID, warehouseID, supplierID *uuid.UUID, manufactureDate, expiryDate *time.Time, quantity, unitCost decimal.Decimal, notes string) *Batch {
	now := clock.NowUTC()
	return &Batch{
		TenantID:         tenantID,
		CompanyID:        companyID,
		ProductID:        productID,
		WarehouseID:      warehouseID,
		BatchNumber:      batchNumber,
		ManufactureDate:  manufactureDate,
		ExpiryDate:       expiryDate,
		Quantity:         quantity,
		UnitCost:         unitCost,
		SupplierID:       supplierID,
		Notes:            notes,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (Batch) TableName() string {
	return "batches"
}

type Supplier struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID   *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	ContactName string     `json:"contact_name" gorm:"type:varchar(255)"`
	Email       string     `json:"email" gorm:"type:varchar(255)"`
	Phone       string     `json:"phone" gorm:"type:varchar(50)"`
	Address     string     `json:"address" gorm:"type:text"`
	City        string     `json:"city" gorm:"type:varchar(100)"`
	State       string     `json:"state" gorm:"type:varchar(100)"`
	Country     string     `json:"country" gorm:"type:varchar(100)"`
	PostalCode  string     `json:"postal_code" gorm:"type:varchar(20)"`
	TaxID       string     `json:"tax_id" gorm:"type:varchar(50)"`
	PaymentTerms string    `json:"payment_terms" gorm:"type:varchar(100)"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	Notes       string     `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func NewSupplier(clock clock.Clock, tenantID uuid.UUID, name string, companyID *uuid.UUID) *Supplier {
	now := clock.NowUTC()
	return &Supplier{
		TenantID:  tenantID,
		CompanyID: companyID,
		Name:      name,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (Supplier) TableName() string {
	return "suppliers"
}

type PurchaseOrder struct {
	ID              uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID            `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID           `json:"company_id" gorm:"type:uuid;index"`
	SupplierID      uuid.UUID            `json:"supplier_id" gorm:"type:uuid;not null;index"`
	WarehouseID     *uuid.UUID           `json:"warehouse_id" gorm:"type:uuid;index"`
	OrderNumber     string               `json:"order_number" gorm:"type:varchar(100);uniqueIndex:idx_purchase_orders_tenant_number;not null"`
	Status          PurchaseOrderStatus  `json:"status" gorm:"type:varchar(20);default:'draft'"`
	OrderDate       time.Time            `json:"order_date" gorm:"not null;index"`
	ExpectedDate    *time.Time           `json:"expected_date" gorm:"index"`
	ReceivedDate    *time.Time           `json:"received_date" gorm:"index"`
	Subtotal        decimal.Decimal      `json:"subtotal" gorm:"type:decimal(15,4);default:0"`
	TaxAmount       decimal.Decimal      `json:"tax_amount" gorm:"type:decimal(15,4);default:0"`
	DiscountAmount  decimal.Decimal      `json:"discount_amount" gorm:"type:decimal(15,4);default:0"`
	TotalAmount     decimal.Decimal      `json:"total_amount" gorm:"type:decimal(15,4);default:0"`
	Notes           string               `json:"notes" gorm:"type:text"`
	CreatedBy       uuid.UUID            `json:"created_by" gorm:"type:uuid;not null"`
	ApprovedBy      *uuid.UUID           `json:"approved_by" gorm:"type:uuid"`
	ApprovedAt      *time.Time           `json:"approved_at"`
	CreatedAt       time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time           `json:"deleted_at,omitempty" gorm:"index"`

	Items      []*PurchaseOrderItem `json:"items" gorm:"foreignKey:PurchaseOrderID"`
}

type PurchaseOrderStatus string

const (
	PurchaseOrderStatusDraft     PurchaseOrderStatus = "draft"
	PurchaseOrderStatusPending   PurchaseOrderStatus = "pending"
	PurchaseOrderStatusApproved  PurchaseOrderStatus = "approved"
	PurchaseOrderStatusOrdered   PurchaseOrderStatus = "ordered"
	PurchaseOrderStatusPartial   PurchaseOrderStatus = "partial"
	PurchaseOrderStatusReceived  PurchaseOrderStatus = "received"
	PurchaseOrderStatusCancelled PurchaseOrderStatus = "cancelled"
)

func NewPurchaseOrder(clock clock.Clock, tenantID, supplierID, createdBy uuid.UUID, orderNumber string, companyID, warehouseID *uuid.UUID, orderDate time.Time, expectedDate *time.Time, notes string) *PurchaseOrder {
	now := clock.NowUTC()
	return &PurchaseOrder{
		TenantID:     tenantID,
		CompanyID:    companyID,
		SupplierID:   supplierID,
		WarehouseID:  warehouseID,
		OrderNumber:  orderNumber,
		Status:       PurchaseOrderStatusDraft,
		OrderDate:    orderDate,
		ExpectedDate: expectedDate,
		Notes:        notes,
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

type PurchaseOrderItem struct {
	ID                 uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PurchaseOrderID    uuid.UUID       `json:"purchase_order_id" gorm:"type:uuid;not null;index"`
	ProductID          uuid.UUID       `json:"product_id" gorm:"type:uuid;not null;index"`
	Quantity           decimal.Decimal `json:"quantity" gorm:"type:decimal(15,4);not null"`
	ReceivedQuantity   decimal.Decimal `json:"received_quantity" gorm:"type:decimal(15,4);default:0"`
	UnitCost           decimal.Decimal `json:"unit_cost" gorm:"type:decimal(15,4);not null"`
	TaxRate            decimal.Decimal `json:"tax_rate" gorm:"type:decimal(5,4);default:0"`
	TaxAmount          decimal.Decimal `json:"tax_amount" gorm:"type:decimal(15,4);default:0"`
	DiscountRate       decimal.Decimal `json:"discount_rate" gorm:"type:decimal(5,4);default:0"`
	DiscountAmount     decimal.Decimal `json:"discount_amount" gorm:"type:decimal(15,4);default:0"`
	TotalAmount        decimal.Decimal `json:"total_amount" gorm:"type:decimal(15,4);not null"`
	Notes              string          `json:"notes" gorm:"type:text"`
	CreatedAt          time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func (PurchaseOrderItem) TableName() string {
	return "purchase_order_items"
}