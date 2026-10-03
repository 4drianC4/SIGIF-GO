package entity

// DocumentType representa el tipo de documento de identificación del cliente.
type DocumentType string

const (
	// DocumentTypeNationalID corresponde a la cédula de identidad (CI).
	DocumentTypeNationalID DocumentType = "national_id"
	// DocumentTypeTaxID corresponde al número de identificación tributaria (NIT).
	DocumentTypeTaxID DocumentType = "tax_id"
	// DocumentTypePassport corresponde a un pasaporte.
	DocumentTypePassport DocumentType = "passport"
	// DocumentTypeOther agrupa cualquier otro tipo de documento.
	DocumentTypeOther DocumentType = "other"
)

// IsValid verifica que el valor sea uno de los tipos oficiales.
func (d DocumentType) IsValid() bool {
	switch d {
	case DocumentTypeNationalID, DocumentTypeTaxID, DocumentTypePassport, DocumentTypeOther:
		return true
	}
	return false
}
