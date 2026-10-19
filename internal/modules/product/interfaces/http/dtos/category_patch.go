package dtos

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/shopspring/decimal"

	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

// PatchCategoryRequest holds the fields PATCH /categories/:id may change. A
// field that was not sent stays nil; the ones listed in the cleared set were
// sent as null or empty.
type PatchCategoryRequest struct {
	Name         *string          `json:"name" validate:"omitempty,min=2,max=100"`
	Description  *string          `json:"description" validate:"omitempty,max=500"`
	DefaultTax   *string          `json:"default_tax" validate:"omitempty,max=60"`
	TargetMargin *decimal.Decimal `json:"target_margin"`
}

// patchCategoryFields lists the body keys PATCH accepts. Optional ones can be
// cleared with null; a cleared default tax or target margin is inherited from
// the parent category.
var patchCategoryFields = map[string]bool{
	"name":          false,
	"description":   true,
	"default_tax":   true,
	"target_margin": true,
}

// patchCategoryReadOnlyFields are category attributes that exist but cannot be
// changed through PATCH; any other unknown key is reported as unknown.
var patchCategoryReadOnlyFields = map[string]bool{
	"id":             true,
	"company_id":     true,
	"parent_id":      true,
	"status":         true,
	"products_count": true,
	"created_at":     true,
	"updated_at":     true,
	"deleted_at":     true,
}

// ParseCategoryPatch reads the body of PATCH /categories/:id. It must be a JSON
// object with at least one editable field and nothing else. The returned
// details list every field that cannot be applied.
func ParseCategoryPatch(body []byte) (PatchCategoryRequest, map[string]bool, map[string]string, error) {
	var req PatchCategoryRequest
	if len(bytes.TrimSpace(body)) == 0 {
		return req, nil, nil, sharedErrors.New(sharedErrors.CodeValidation, "at least one editable field is required", 400)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		return req, nil, nil, sharedErrors.New(sharedErrors.CodeBadRequest, "request body must be a JSON object", 400)
	}
	if len(raw) == 0 {
		return req, nil, nil, sharedErrors.New(sharedErrors.CodeValidation, "at least one editable field is required", 400)
	}

	cleared := map[string]bool{}
	details := map[string]string{}
	for key, value := range raw {
		optional, editable := patchCategoryFields[key]
		if !editable {
			if patchCategoryReadOnlyFields[key] {
				details[key] = key + " is not editable"
			} else {
				details[key] = key + " is not a known field"
			}
			continue
		}

		if key == "target_margin" {
			var margin *decimal.Decimal
			if err := json.Unmarshal(value, &margin); err != nil {
				details[key] = key + " must be a number"
				continue
			}
			if margin == nil {
				cleared[key] = true
				continue
			}
			req.TargetMargin = margin
			continue
		}

		var s *string
		if err := json.Unmarshal(value, &s); err != nil {
			details[key] = key + " must be a string"
			continue
		}
		if s != nil {
			trimmed := strings.TrimSpace(*s)
			s = &trimmed
		}
		if s == nil || *s == "" {
			if optional {
				cleared[key] = true
				continue
			}
			details[key] = key + " cannot be empty"
			continue
		}
		switch key {
		case "name":
			req.Name = s
		case "description":
			req.Description = s
		case "default_tax":
			req.DefaultTax = s
		}
	}
	return req, cleared, details, nil
}
