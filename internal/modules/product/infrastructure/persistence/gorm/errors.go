package gorm

import (
	"errors"

	"gorm.io/gorm"
)

func isDuplicateKey(db *gorm.DB, err error) bool {
	if translator, ok := db.Dialector.(gorm.ErrorTranslator); ok {
		err = translator.Translate(err)
	}
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
