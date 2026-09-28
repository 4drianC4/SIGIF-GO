package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DashboardMetric struct {
	TenantID    uuid.UUID       `json:"tenant_id"`
	CompanyID   *uuid.UUID      `json:"company_id,omitempty"`
	PeriodStart time.Time       `json:"period_start"`
	PeriodEnd   time.Time       `json:"period_end"`

	TotalSales         decimal.Decimal `json:"total_sales"`
	TotalOrders        int64           `json:"total_orders"`
	AverageOrderValue  decimal.Decimal `json:"average_order_value"`
	TotalCustomers     int64           `json:"total_customers"`
	NewCustomers       int64           `json:"new_customers"`
	TotalProducts      int64           `json:"total_products"`
	LowStockProducts   int64           `json:"low_stock_products"`
	OutOfStockProducts int64           `json:"out_of_stock_products"`
	TotalInventoryValue decimal.Decimal `json:"total_inventory_value"`
	GrossProfit        decimal.Decimal `json:"gross_profit"`
	ProfitMargin       decimal.Decimal `json:"profit_margin"`

	TopProducts        []TopProductMetric `json:"top_products"`
	TopCategories      []TopCategoryMetric `json:"top_categories"`
	SalesByDay         []DailySalesMetric `json:"sales_by_day"`
	SalesByPaymentMethod []PaymentMethodMetric `json:"sales_by_payment_method"`
	SalesByHour        []HourlySalesMetric `json:"sales_by_hour"`
}

type TopProductMetric struct {
	ProductID   uuid.UUID       `json:"product_id"`
	ProductName string          `json:"product_name"`
	SKU         string          `json:"sku"`
	Quantity    decimal.Decimal `json:"quantity"`
	Revenue     decimal.Decimal `json:"revenue"`
	Margin      decimal.Decimal `json:"margin"`
}

type TopCategoryMetric struct {
	CategoryID   uuid.UUID       `json:"category_id"`
	CategoryName string          `json:"category_name"`
	Quantity     decimal.Decimal `json:"quantity"`
	Revenue      decimal.Decimal `json:"revenue"`
}

type DailySalesMetric struct {
	Date       time.Time       `json:"date"`
	Sales      decimal.Decimal `json:"sales"`
	Orders     int64           `json:"orders"`
	Customers  int64           `json:"customers"`
}

type PaymentMethodMetric struct {
	Method  string          `json:"method"`
	Amount  decimal.Decimal `json:"amount"`
	Count   int64           `json:"count"`
}

type HourlySalesMetric struct {
	Hour   int             `json:"hour"`
	Sales  decimal.Decimal `json:"sales"`
	Orders int64           `json:"orders"`
}

type ReportType string

const (
	ReportTypeSales          ReportType = "sales"
	ReportTypeInventory      ReportType = "inventory"
	ReportTypeCustomers      ReportType = "customers"
	ReportTypeProducts       ReportType = "products"
	ReportTypeProfitability  ReportType = "profitability"
	ReportTypeTax            ReportType = "tax"
	ReportTypeCustom         ReportType = "custom"
)

type Report struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID   *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Type        ReportType `json:"type" gorm:"type:varchar(50);not null"`
	Description string     `json:"description" gorm:"type:text"`
	Query       string     `json:"query" gorm:"type:text"`
	Parameters  string     `json:"parameters" gorm:"type:jsonb"`
	Schedule    string     `json:"schedule" gorm:"type:varchar(100)"`
	Format      string     `json:"format" gorm:"type:varchar(20);default:'json'"`
	Recipients  string     `json:"recipients" gorm:"type:jsonb"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	LastRunAt   *time.Time `json:"last_run_at" gorm:"index"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Report) TableName() string {
	return "reports"
}

type ReportExecution struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ReportID      uuid.UUID  `json:"report_id" gorm:"type:uuid;not null;index"`
	Status        string     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	StartedAt     time.Time  `json:"started_at" gorm:"not null"`
	CompletedAt   *time.Time `json:"completed_at"`
	ResultURL     string     `json:"result_url" gorm:"type:varchar(500)"`
	ErrorMessage  string     `json:"error_message" gorm:"type:text"`
	RowCount      int64      `json:"row_count"`
	ExecutionTime int64      `json:"execution_time_ms"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`

	Report *Report `json:"-" gorm:"foreignKey:ReportID"`
}

func (ReportExecution) TableName() string {
	return "report_executions"
}