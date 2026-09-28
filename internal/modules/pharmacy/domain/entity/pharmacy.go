package entity

import (
	"time"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sigif/sigif-go/internal/shared/clock"
)

type PrescriptionStatus string

const (
	PrescriptionStatusPending   PrescriptionStatus = "pending"
	PrescriptionStatusVerified  PrescriptionStatus = "verified"
	PrescriptionStatusDispensed PrescriptionStatus = "dispensed"
	PrescriptionStatusPartial   PrescriptionStatus = "partial"
	PrescriptionStatusCancelled PrescriptionStatus = "cancelled"
	PrescriptionStatusExpired   PrescriptionStatus = "expired"
)

type PrescriptionType string

const (
	PrescriptionTypeSimple      PrescriptionType = "simple"
	PrescriptionTypeControlled  PrescriptionType = "controlled"
	PrescriptionTypePsychotropic PrescriptionType = "psychotropic"
	PrescriptionTypeNarcotic    PrescriptionType = "narcotic"
	PrescriptionTypeSpecial     PrescriptionType = "special"
)

type Prescription struct {
	ID                  uuid.UUID           `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID            uuid.UUID           `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID           *uuid.UUID          `json:"company_id" gorm:"type:uuid;index"`
	CustomerID          *uuid.UUID          `json:"customer_id" gorm:"type:uuid;index"`
	DoctorID            *uuid.UUID          `json:"doctor_id" gorm:"type:uuid;index"`
	PrescriptionNumber  string              `json:"prescription_number" gorm:"type:varchar(100);uniqueIndex:idx_prescriptions_tenant_number;not null"`
	Status              PrescriptionStatus  `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Type                PrescriptionType    `json:"type" gorm:"type:varchar(20);default:'simple'"`
	PrescriptionDate    time.Time           `json:"prescription_date" gorm:"not null;index"`
	ExpiryDate          *time.Time          `json:"expiry_date" gorm:"index"`
	DispenseDate        *time.Time          `json:"dispense_date" gorm:"index"`
	DoctorName          string              `json:"doctor_name" gorm:"type:varchar(255)"`
	DoctorLicense       string              `json:"doctor_license" gorm:"type:varchar(100)"`
	DoctorSpecialty     string              `json:"doctor_specialty" gorm:"type:varchar(100)"`
	ClinicName          string              `json:"clinic_name" gorm:"type:varchar(255)"`
	ClinicAddress       string              `json:"clinic_address" gorm:"type:text"`
	Diagnosis           string              `json:"diagnosis" gorm:"type:text"`
	Notes               string              `json:"notes" gorm:"type:text"`
	VerifiedBy          *uuid.UUID          `json:"verified_by" gorm:"type:uuid"`
	VerifiedAt          *time.Time          `json:"verified_at"`
	DispensedBy         *uuid.UUID          `json:"dispensed_by" gorm:"type:uuid"`
	DispensedAt         *time.Time          `json:"dispensed_at"`
	CreatedBy           uuid.UUID           `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt           time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time           `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt           *time.Time          `json:"deleted_at,omitempty" gorm:"index"`

	Items    []*PrescriptionItem `json:"items" gorm:"foreignKey:PrescriptionID"`
}

func NewPrescription(clock clock.Clock, tenantID, createdBy uuid.UUID, prescriptionNumber string, companyID, customerID, doctorID *uuid.UUID, prescriptionDate time.Time, prescriptionType PrescriptionType, doctorName, doctorLicense, doctorSpecialty, clinicName, clinicAddress, diagnosis, notes string, expiryDate *time.Time) *Prescription {
	now := clock.NowUTC()
	return &Prescription{
		TenantID:            tenantID,
		CompanyID:           companyID,
		CustomerID:          customerID,
		DoctorID:            doctorID,
		PrescriptionNumber:  prescriptionNumber,
		Status:              PrescriptionStatusPending,
		Type:                prescriptionType,
		PrescriptionDate:    prescriptionDate,
		ExpiryDate:          expiryDate,
		DoctorName:          doctorName,
		DoctorLicense:       doctorLicense,
		DoctorSpecialty:     doctorSpecialty,
		ClinicName:          clinicName,
		ClinicAddress:       clinicAddress,
		Diagnosis:           diagnosis,
		Notes:               notes,
		CreatedBy:           createdBy,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func (p *Prescription) Verify(clock clock.Clock, verifiedBy uuid.UUID) {
	p.Status = PrescriptionStatusVerified
	p.VerifiedBy = &verifiedBy
	now := clock.NowUTC()
	p.VerifiedAt = &now
	p.UpdatedAt = now
}

func (p *Prescription) Dispense(clock clock.Clock, dispensedBy uuid.UUID) {
	p.Status = PrescriptionStatusDispensed
	p.DispensedBy = &dispensedBy
	now := clock.NowUTC()
	p.DispensedAt = &now
	p.UpdatedAt = now
}

func (p *Prescription) Cancel(clock clock.Clock, cancelledBy uuid.UUID, reason string) {
	p.Status = PrescriptionStatusCancelled
	p.Notes = p.Notes + "\nCancelled: " + reason
	now := clock.NowUTC()
	p.UpdatedAt = now
}

func (Prescription) TableName() string {
	return "prescriptions"
}

type PrescriptionItem struct {
	ID                 uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PrescriptionID     uuid.UUID       `json:"prescription_id" gorm:"type:uuid;not null;index"`
	ProductID          uuid.UUID       `json:"product_id" gorm:"type:uuid;not null;index"`
	ProductName        string          `json:"product_name" gorm:"type:varchar(255);not null"`
	Dosage             string          `json:"dosage" gorm:"type:varchar(100)"`
	Form               string          `json:"form" gorm:"type:varchar(50)"`
	Route              string          `json:"route" gorm:"type:varchar(50)"`
	Frequency          string          `json:"frequency" gorm:"type:varchar(100)"`
	Duration           string          `json:"duration" gorm:"type:varchar(100)"`
	Quantity           decimal.Decimal `json:"quantity" gorm:"type:decimal(15,4);not null"`
	DispensedQuantity  decimal.Decimal `json:"dispensed_quantity" gorm:"type:decimal(15,4);default:0"`
	Instructions       string          `json:"instructions" gorm:"type:text"`
	SubstitutionAllowed bool           `json:"substitution_allowed" gorm:"default:true"`
	Notes              string          `json:"notes" gorm:"type:text"`
	CreatedAt          time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func (PrescriptionItem) TableName() string {
	return "prescription_items"
}

type Doctor struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID       *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	FirstName       string     `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName        string     `json:"last_name" gorm:"type:varchar(100);not null"`
	LicenseNumber   string     `json:"license_number" gorm:"type:varchar(100);uniqueIndex:idx_doctors_tenant_license;not null"`
	Specialty       string     `json:"specialty" gorm:"type:varchar(100)"`
	Email           string     `json:"email" gorm:"type:varchar(255)"`
	Phone           string     `json:"phone" gorm:"type:varchar(50)"`
	ClinicName      string     `json:"clinic_name" gorm:"type:varchar(255)"`
	ClinicAddress   string     `json:"clinic_address" gorm:"type:text"`
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func NewDoctor(clock clock.Clock, tenantID uuid.UUID, firstName, lastName, licenseNumber string, companyID *uuid.UUID) *Doctor {
	now := clock.NowUTC()
	return &Doctor{
		TenantID:      tenantID,
		CompanyID:     companyID,
		FirstName:     firstName,
		LastName:      lastName,
		LicenseNumber: licenseNumber,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (d *Doctor) FullName() string {
	return d.FirstName + " " + d.LastName
}

func (Doctor) TableName() string {
	return "doctors"
}

type ControlledSubstanceLog struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID            uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	CompanyID           *uuid.UUID `json:"company_id" gorm:"type:uuid;index"`
	PrescriptionID      *uuid.UUID `json:"prescription_id" gorm:"type:uuid;index"`
	ProductID           uuid.UUID  `json:"product_id" gorm:"type:uuid;not null;index"`
	BatchID             *uuid.UUID `json:"batch_id" gorm:"type:uuid;index"`
	Action              string     `json:"action" gorm:"type:varchar(50);not null"`
	Quantity            decimal.Decimal `json:"quantity" gorm:"type:decimal(15,4);not null"`
	PreviousStock       decimal.Decimal `json:"previous_stock" gorm:"type:decimal(15,4);not null"`
	NewStock            decimal.Decimal `json:"new_stock" gorm:"type:decimal(15,4);not null"`
	PerformedBy         uuid.UUID  `json:"performed_by" gorm:"type:uuid;not null"`
	AuthorizedBy        *uuid.UUID `json:"authorized_by" gorm:"type:uuid"`
	Notes               string     `json:"notes" gorm:"type:text"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`

	}

func NewControlledSubstanceLog(clock clock.Clock, tenantID, productID, performedBy uuid.UUID, action string, quantity, previousStock, newStock decimal.Decimal, companyID, prescriptionID, batchID *uuid.UUID, authorizedBy *uuid.UUID, notes string) *ControlledSubstanceLog {
	now := clock.NowUTC()
	return &ControlledSubstanceLog{
		TenantID:       tenantID,
		CompanyID:      companyID,
		PrescriptionID: prescriptionID,
		ProductID:      productID,
		BatchID:        batchID,
		Action:         action,
		Quantity:       quantity,
		PreviousStock:  previousStock,
		NewStock:       newStock,
		PerformedBy:    performedBy,
		AuthorizedBy:   authorizedBy,
		Notes:          notes,
		CreatedAt:      now,
	}
}

func (ControlledSubstanceLog) TableName() string {
	return "controlled_substance_logs"
}