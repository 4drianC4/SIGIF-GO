package gorm

import "strings"

// isDuplicateKey reports whether err is a PostgreSQL unique (or exclusion)
// constraint violation, used to translate the permission module/operation code
// index into a 409 conflict.
func isDuplicateKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
