package entity

import "github.com/shopspring/decimal"

type ProductDetail struct {
	*Product
	StockTotal decimal.Decimal
}
