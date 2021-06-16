package structs

import (
	"reflect"
)

// Type wraps reflect.Type for additional features.
type Type struct {
	reflect.Type
}

// Field contains information of a struct field .
type Field struct {
	Value    reflect.Value       // The underlying value of the field.
	Field    reflect.StructField // The underlying field of the field.
	TagValue string              // Retrieved tag value. There might be more than one tags in the field, but only one can be retrieved according to calling function rules.
}
