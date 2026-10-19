package entity

// Optional is a field of a partial update (PATCH): not sent (Set is false),
// cleared (Set is true and Value is nil) or given a value.
type Optional[T any] struct {
	Set   bool
	Value *T
}

// Some returns an Optional that sets the field to value.
func Some[T any](value T) Optional[T] {
	return Optional[T]{Set: true, Value: &value}
}

// Cleared returns an Optional that removes the value of the field.
func Cleared[T any]() Optional[T] {
	return Optional[T]{Set: true}
}
