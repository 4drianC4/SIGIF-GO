package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type createdCustomerCredit struct {
	ID            string `json:"id"`
	CreditLimit   string `json:"credit_limit"`
	CreditBalance string `json:"credit_balance"`
	PointsAccrued int    `json:"points_accrued"`
}

func TestCreateCustomerCreditLimitRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "omitted", want: "0"},
		{name: "null", value: "null", want: "0"},
		{name: "zero", value: "0", want: "0"},
		{name: "number", value: "2000", want: "2000"},
		{name: "decimal string", value: `"2000.25"`, want: "2000.25"},
		{name: "decimal number", value: "2000.25", want: "2000.25"},
		{name: "trailing zeros", value: `"2000.2500"`, want: "2000.25"},
		{name: "smallest positive", value: `"0.01"`, want: "0.01"},
		{name: "maximum column value", value: `"9999999999.99"`, want: "9999999999.99"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(t)
			token := s.token(t, companyA, "customers.create", "customers.read", "customers.list")
			body := map[string]any{"legal_name": "Credit customer"}
			if tt.value != "" {
				body["credit_limit"] = json.RawMessage(tt.value)
			}
			status, resp := s.send(t, fiber.MethodPost, "/api/v1/customers/", token, body)
			if status != fiber.StatusCreated || !resp.Success {
				t.Fatalf("create status = %d, response = %+v", status, resp)
			}
			var created createdCustomerCredit
			if err := json.Unmarshal(resp.Data, &created); err != nil {
				t.Fatalf("decode created customer: %v", err)
			}
			if created.CreditLimit != tt.want || created.CreditBalance != "0" || created.PointsAccrued != 0 {
				t.Fatalf("created credit = %+v, want limit %s and zero balance/points", created, tt.want)
			}
			stored := s.stored(t, uuid.MustParse(created.ID))
			if !stored.CreditLimit.Equal(decimal.RequireFromString(tt.want)) ||
				!stored.CreditBalance.IsZero() || stored.PointsAccrued != 0 {
				t.Fatalf("stored credit limit = %s, balance = %s, points = %d", stored.CreditLimit, stored.CreditBalance, stored.PointsAccrued)
			}
			status, resp = s.send(t, fiber.MethodGet, "/api/v1/customers/"+created.ID, token, nil)
			if status != fiber.StatusOK || !resp.Success {
				t.Fatalf("get status = %d, response = %+v", status, resp)
			}
			var fetched createdCustomerCredit
			if err := json.Unmarshal(resp.Data, &fetched); err != nil {
				t.Fatalf("decode fetched customer: %v", err)
			}
			if fetched != created {
				t.Fatalf("fetched credit = %+v, want %+v", fetched, created)
			}
			status, resp, items := s.list(t, "", token)
			if status != fiber.StatusOK || !resp.Success || len(items) != 1 {
				t.Fatalf("list status = %d, response = %+v, items = %+v", status, resp, items)
			}
			if items[0].ID != created.ID || items[0].CreditLimit != tt.want {
				t.Fatalf("listed customer = %+v, want id %s and credit limit %s", items[0], created.ID, tt.want)
			}
		})
	}
}

func TestCreateCustomerRejectsInvalidCreditLimitWithoutWriting(t *testing.T) {
	tests := []struct {
		name  string
		value string
		code  string
	}{
		{name: "negative number", value: "-1", code: "VALIDATION_ERROR"},
		{name: "negative string", value: `"-0.01"`, code: "VALIDATION_ERROR"},
		{name: "extra decimal places", value: `"2000.251"`, code: "VALIDATION_ERROR"},
		{name: "column overflow", value: `"10000000000"`, code: "VALIDATION_ERROR"},
		{name: "invalid string", value: `"abc"`, code: "BAD_REQUEST"},
		{name: "empty string", value: `""`, code: "BAD_REQUEST"},
		{name: "boolean", value: "true", code: "BAD_REQUEST"},
		{name: "object", value: "{}", code: "BAD_REQUEST"},
		{name: "array", value: "[]", code: "BAD_REQUEST"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(t)
			token := s.token(t, companyA, "customers.create")
			status, resp := s.send(t, fiber.MethodPost, "/api/v1/customers/", token, map[string]any{
				"legal_name":   "Credit customer",
				"credit_limit": json.RawMessage(tt.value),
			})
			if status != fiber.StatusBadRequest || resp.Success || resp.Error == nil || resp.Error.Code != tt.code {
				t.Fatalf("status = %d, response = %+v, want 400/%s", status, resp, tt.code)
			}
			if tt.code == "VALIDATION_ERROR" && resp.Error.Details["credit_limit"] == "" {
				t.Fatalf("missing credit_limit validation detail: %+v", resp.Error)
			}
			if s.customers.Count() != 0 || s.customers.Writes != 0 {
				t.Fatal("invalid credit limit caused a repository write")
			}
		})
	}
}
