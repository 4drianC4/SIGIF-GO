package entity

type DocumentType string

const (
	DocumentTypeNationalID DocumentType = "national_id"
	DocumentTypeTaxID DocumentType = "tax_id"
	DocumentTypePassport DocumentType = "passport"
	DocumentTypeOther DocumentType = "other"
)

func (d DocumentType) IsValid() bool {
	switch d {
	case DocumentTypeNationalID, DocumentTypeTaxID, DocumentTypePassport, DocumentTypeOther:
		return true
	}
	return false
}
