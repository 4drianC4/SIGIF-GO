package gorm

import (
	"errors"

	"gorm.io/gorm"
)

// isDuplicateKey detecta una violación de índice único. Cubre la carrera entre
// la comprobación previa del servicio y el INSERT de dos peticiones simultáneas.
func isDuplicateKey(db *gorm.DB, err error) bool {
	if translator, ok := db.Dialector.(gorm.ErrorTranslator); ok {
		err = translator.Translate(err)
	}
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
