package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type SaleStatus string

const (
	SaleStatusDraft       SaleStatus = "draft"
	SaleStatusPending     SaleStatus = "pending"
	SaleStatusCompleted   SaleStatus = "completed"
	SaleStatusCancelled   SaleStatus = "cancelled"
	SaleStatusRefunded    SaleStatus = "refunded"
	SaleStatusPartial     SaleStatus = "partial"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPartial   PaymentStatus = "partial"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusRefunded  PaymentStatus = "refunded"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type PaymentMethod string

const (
	PaymentMethodCash       PaymentMethod = "cash"
	PaymentMethodCard       PaymentMethod = "card"
	PaymentMethodTransfer   PaymentMethod = "transfer"
	PaymentMethodCheck      PaymentMethod = "check"
	PaymentMethodMobile     PaymentMethod = "mobile"
	PaymentMethodGiftCard   PaymentMethod = "gift_card"
	PaymentMethodCredit     PaymentMethod = "credit"
	PaymentMethodMixed      PaymentMethod = "mixed"
)

type Sale struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID      `json:"company_id" gorm:"type:uuid;index"`
	CustomerID      *uuid.UUID      `json:"customer_id" gorm:"type:uuid;index"`
	WarehouseID     *uuid.UUID      `json:"warehouse_id" gorm:"type:uuid;index"`
	SaleNumber      string          `json:"sale_number" gorm:"type:varchar(100);uniqueIndex:idx_sales_tenant_number;not null"`
	Status          SaleStatus      `json:"status" gorm:"type:varchar(20);default:'draft'"`
	PaymentStatus   PaymentStatus   `json:"payment_status" gorm:"type:varchar(20);default:'pending'"`
	PaymentMethod   PaymentMethod   `json:"payment_method" gorm:"type:varchar(20);default:'cash'"`
	SaleDate        time.Time       `json:"sale_date" gorm:"not null;index"`
	DueDate         *time.Time      `json:"due_date" gorm:"index"`
	Subtotal        decimal.Decimal `json:"subtotal" gorm:"type:decimal(15,4);default:0"`
	TaxAmount       decimal.Decimal `json:"tax_amount" gorm:"type:decimal(15,4);default:0"`
	DiscountAmount  decimal.Decimal `json:"discount_amount" gorm:"type:decimal(15,4);default:0"`
	TotalAmount     decimal.Decimal `json:"total_amount" gorm:"type:decimal(15,4);default:0"`
	PaidAmount      decimal.Decimal `json:"paid_amount" gorm:"type:decimal(15,4);default:0"`
	ChangeAmount    decimal.Decimal `json:"change_amount" gorm:"type:decimal(15,4);default:0"`
	Notes           string          `json:"notes" gorm:"type:text"`
	CreatedBy       uuid.UUID       `json:"created_by" gorm:"type:uuid;not null"`
	CompletedBy     *uuid.UUID      `json:"completed_by" gorm:"type:uuid"`
	CompletedAt     *time.Time      `json:"completed_at"`
	CancelledBy     *uuid.UUID      `json:"cancelled_by" gorm:"type:uuid"`
	CancelledAt     *time.Time      `json:"cancelled_at"`
	CancelReason    string          `json:"cancel_reason" gorm:"type:text"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty" gorm:"index"`

	Items      []*SaleItem `json:"items" gorm:"foreignKey:SaleID"`
	Payments   []*Payment  `json:"payments" gorm:"foreignKey:SaleID"`
}

func NewSale(clock clock.Clock, tenantID, createdBy uuid.UUID, saleNumber string, companyID, customerID, warehouseID *uuid.UUID, saleDate time.Time, paymentMethod PaymentMethod, notes string) *Sale {
	now := clock.NowUTC()
	return &Sale{
		TenantID:      tenantID,
		CompanyID:     companyID,
		CustomerID:    customerID,
		WarehouseID:   warehouseID,
		SaleNumber:    saleNumber,
		Status:        SaleStatusDraft,
		PaymentStatus: PaymentStatusPending,
		PaymentMethod: paymentMethod,
		SaleDate:      saleDate,
		Notes:         notes,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (s *Sale) AddItem(productID uuid.UUID, quantity, unitPrice, taxRate, discountRate decimal.Decimal, notes string) *SaleItem {
	item := &SaleItem{
		SaleID:       s.ID,
		ProductID:    productID,
		Quantity:     quantity,
		UnitPrice:    unitPrice,
		TaxRate:      taxRate,
		DiscountRate: discountRate,
		Notes:        notes,
	}
	item.Calculate()
	s.Items = append(s.Items, item)
	s.Recalculate()
	return item
}

func (s *Sale) Recalculate() {
	s.Subtotal = decimal.Zero
	s.TaxAmount = decimal.Zero
	s.DiscountAmount = decimal.Zero

	for _, item := range s.Items {
		s.Subtotal = s.Subtotal.Add(item.Subtotal)
		s.TaxAmount = s.TaxAmount.Add(item.TaxAmount)
		s.DiscountAmount = s.DiscountAmount.Add(item.DiscountAmount)
	}

	s.TotalAmount = s.Subtotal.Add(s.TaxAmount).Sub(s.DiscountAmount)
}

func (s *Sale) Complete(clock clock.Clock, completedBy uuid.UUID) {
	s.Status = SaleStatusCompleted
	s.CompletedBy = &completedBy
	now := clock.NowUTC()
	s.CompletedAt = &now
	s.UpdatedAt = now
}

func (s *Sale) Cancel(clock clock.Clock, cancelledBy uuid.UUID, reason string) {
	s.Status = SaleStatusCancelled
	s.CancelledBy = &cancelledBy
	s.CancelReason = reason
	now := clock.NowUTC()
	s.CancelledAt = &now
	s.UpdatedAt = now
}

func (Sale) TableName() string {
	return "sales"
}

type SaleItem struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SaleID        uuid.UUID       `json:"sale_id" gorm:"type:uuid;not null;index"`
	ProductID     uuid.UUID       `json:"product_id" gorm:"type:uuid;not null;index"`
	BatchID       *uuid.UUID      `json:"batch_id" gorm:"type:uuid;index"`
	SerialNumber  string          `json:"serial_number" gorm:"type:varchar(100)"`
	Quantity      decimal.Decimal `json:"quantity" gorm:"type:decimal(15,4);not null"`
	UnitPrice     decimal.Decimal `json:"unit_price" gorm:"type:decimal(15,4);not null"`
	TaxRate       decimal.Decimal `json:"tax_rate" gorm:"type:decimal(5,4);default:0"`
	DiscountRate  decimal.Decimal `json:"discount_rate" gorm:"type:decimal(5,4);default:0"`
	Subtotal      decimal.Decimal `json:"subtotal" gorm:"type:decimal(15,4);not null"`
	TaxAmount     decimal.Decimal `json:"tax_amount" gorm:"type:decimal(15,4);default:0"`
	DiscountAmount decimal.Decimal `json:"discount_amount" gorm:"type:decimal(15,4);default:0"`
	TotalAmount   decimal.Decimal `json:"total_amount" gorm:"type:decimal(15,4);not null"`
	Notes         string          `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	Sale  *Sale  `json:"-" gorm:"foreignKey:SaleID"`
}

func (i *SaleItem) Calculate() {
	i.Subtotal = i.UnitPrice.Mul(i.Quantity)
	i.DiscountAmount = i.Subtotal.Mul(i.DiscountRate).Div(decimal.NewFromInt(100))
	taxableAmount := i.Subtotal.Sub(i.DiscountAmount)
	i.TaxAmount = taxableAmount.Mul(i.TaxRate).Div(decimal.NewFromInt(100))
	i.TotalAmount = i.Subtotal.Sub(i.DiscountAmount).Add(i.TaxAmount)
}

func (SaleItem) TableName() string {
	return "sale_items"
}

type Customer struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID      uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID     *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	Code          string     `json:"code" gorm:"type:varchar(50);uniqueIndex:idx_customers_tenant_code;not null"`
	FirstName     string     `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName      string     `json:"last_name" gorm:"type:varchar(100);not null"`
	Email         string     `json:"email" gorm:"type:varchar(255)"`
	Phone         string     `json:"phone" gorm:"type:varchar(50)"`
	DocumentType  string     `json:"document_type" gorm:"type:varchar(20)"`
	DocumentNumber string    `json:"document_number" gorm:"type:varchar(50);index"`
	Address       string     `json:"address" gorm:"type:text"`
	City          string     `json:"city" gorm:"type:varchar(100)"`
	State         string     `json:"state" gorm:"type:varchar(100)"`
	Country       string     `json:"country" gorm:"type:varchar(100)"`
	PostalCode    string     `json:"postal_code" gorm:"type:varchar(20)"`
	BirthDate     *time.Time `json:"birth_date" gorm:"index"`
	Gender        string     `json:"gender" gorm:"type:varchar(20)"`
	CustomerType  string     `json:"customer_type" gorm:"type:varchar(20);default:'regular'"`
	CreditLimit   decimal.Decimal `json:"credit_limit" gorm:"type:decimal(15,4);default:0"`
	CurrentBalance decimal.Decimal `json:"current_balance" gorm:"type:decimal(15,4);default:0"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	Notes         string     `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func NewCustomer(clock clock.Clock, tenantID uuid.UUID, code, firstName, lastName string, companyID *uuid.UUID) *Customer {
	now := clock.NowUTC()
	return &Customer{
		TenantID:     tenantID,
		CompanyID:    companyID,
		Code:         code,
		FirstName:    firstName,
		LastName:     lastName,
		CustomerType: "regular",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (c *Customer) FullName() string {
	return c.FirstName + " " + c.LastName
}

func (Customer) TableName() string {
	return "customers"
}

type Payment struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID      uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	SaleID        uuid.UUID       `json:"sale_id" gorm:"type:uuid;not null;index"`
	CustomerID    *uuid.UUID      `json:"customer_id" gorm:"type:uuid;index"`
	PaymentNumber string          `json:"payment_number" gorm:"type:varchar(100);uniqueIndex:idx_payments_tenant_number;not null"`
	Amount        decimal.Decimal `json:"amount" gorm:"type:decimal(15,4);not null"`
	Method        PaymentMethod   `json:"method" gorm:"type:varchar(20);not null"`
	Reference     string          `json:"reference" gorm:"type:varchar(100)"`
	Status        PaymentStatus   `json:"status" gorm:"type:varchar(20);default:'pending'"`
	PaymentDate   time.Time       `json:"payment_date" gorm:"not null;index"`
	ProcessedBy   *uuid.UUID      `json:"processed_by" gorm:"type:uuid"`
	ProcessedAt   *time.Time      `json:"processed_at"`
	Notes         string          `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     *time.Time      `json:"deleted_at,omitempty" gorm:"index"`
}

func NewPayment(clock clock.Clock, tenantID, saleID uuid.UUID, paymentNumber string, amount decimal.Decimal, method PaymentMethod, customerID *uuid.UUID, reference, notes string) *Payment {
	now := clock.NowUTC()
	return &Payment{
		TenantID:      tenantID,
		SaleID:        saleID,
		CustomerID:    customerID,
		PaymentNumber: paymentNumber,
		Amount:        amount,
		Method:        method,
		Reference:     reference,
		Status:        PaymentStatusPending,
		PaymentDate:   now,
		Notes:         notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (Payment) TableName() string {
	return "payments"
}

type Return struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID      `json:"company_id" gorm:"type:uuid;index"`
	SaleID          uuid.UUID       `json:"sale_id" gorm:"type:uuid;not null;index"`
	CustomerID      *uuid.UUID      `json:"customer_id" gorm:"type:uuid;index"`
	ReturnNumber    string          `json:"return_number" gorm:"type:varchar(100);uniqueIndex:idx_returns_tenant_number;not null"`
	Status          ReturnStatus    `json:"status" gorm:"type:varchar(20);default:'draft'"`
	ReturnDate      time.Time       `json:"return_date" gorm:"not null;index"`
	Reason          string          `json:"reason" gorm:"type:varchar(100)"`
	Subtotal        decimal.Decimal `json:"subtotal" gorm:"type:decimal(15,4);default:0"`
	TaxAmount       decimal.Decimal `json:"tax_amount" gorm:"type:decimal(15,4);default:0"`
	TotalAmount     decimal.Decimal `json:"total_amount" gorm:"type:decimal(15,4);default:0"`
	RefundAmount    decimal.Decimal `json:"refund_amount" gorm:"type:decimal(15,4);default:0"`
	Notes           string          `json:"notes" gorm:"type:text"`
	ProcessedBy     *uuid.UUID      `json:"processed_by" gorm:"type:uuid"`
	ProcessedAt     *time.Time      `json:"processed_at"`
	CreatedBy       uuid.UUID       `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty" gorm:"index"`

	Items    []*ReturnItem `json:"items" gorm:"foreignKey:ReturnID"`
}

type ReturnStatus string

const (
	ReturnStatusDraft      ReturnStatus = "draft"
	ReturnStatusPending    ReturnStatus = "pending"
	ReturnStatusApproved   ReturnStatus = "approved"
	ReturnStatusCompleted  ReturnStatus = "completed"
	ReturnStatusCancelled  ReturnStatus = "cancelled"
)

func NewReturn(clock clock.Clock, tenantID, saleID, createdBy uuid.UUID, returnNumber string, companyID, customerID *uuid.UUID, returnDate time.Time, reason, notes string) *Return {
	now := clock.NowUTC()
	return &Return{
		TenantID:     tenantID,
		CompanyID:    companyID,
		SaleID:       saleID,
		CustomerID:   customerID,
		ReturnNumber: returnNumber,
		Status:       ReturnStatusDraft,
		ReturnDate:   returnDate,
		Reason:       reason,
		Notes:        notes,
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (Return) TableName() string {
	return "returns"
}

type ReturnItem struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ReturnID        uuid.UUID       `json:"return_id" gorm:"type:uuid;not null;index"`
	SaleItemID      uuid.UUID       `json:"sale_item_id" gorm:"type:uuid;not null;index"`
	ProductID       uuid.UUID       `json:"product_id" gorm:"type:uuid;not null;index"`
	BatchID         *uuid.UUID      `json:"batch_id" gorm:"type:uuid;index"`
	SerialNumber    string          `json:"serial_number" gorm:"type:varchar(100)"`
	Quantity        decimal.Decimal `json:"quantity" gorm:"type:decimal(15,4);not null"`
	UnitPrice       decimal.Decimal `json:"unit_price" gorm:"type:decimal(15,4);not null"`
	TaxRate         decimal.Decimal `json:"tax_rate" gorm:"type:decimal(5,4);default:0"`
	Subtotal        decimal.Decimal `json:"subtotal" gorm:"type:decimal(15,4);not null"`
	TaxAmount       decimal.Decimal `json:"tax_amount" gorm:"type:decimal(15,4);default:0"`
	TotalAmount     decimal.Decimal `json:"total_amount" gorm:"type:decimal(15,4);not null"`
	Condition       string          `json:"condition" gorm:"type:varchar(50)"`
	Notes           string          `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ReturnItem) TableName() string {
	return "return_items"
}