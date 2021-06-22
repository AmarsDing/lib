package yvar

import (
	"github.com/AmarsDing/lib/util/yconv"
)

// Struct maps value of <v> to <pointer>.
// The parameter <pointer> should be a pointer to a struct instance.
// The parameter <mapping> is used to specify the key-to-attribute mapping rules.
func (v *Var) Struct(pointer interface{}, mapping ...map[string]string) error {
	return yconv.Struct(v.Val(), pointer, mapping...)
}

// Struct maps value of <v> to <pointer> recursively.
// The parameter <pointer> should be a pointer to a struct instance.
// The parameter <mapping> is used to specify the key-to-attribute mapping rules.
// Deprecated, use Struct instead.
func (v *Var) StructDeep(pointer interface{}, mapping ...map[string]string) error {
	return yconv.StructDeep(v.Val(), pointer, mapping...)
}

// Structs converts and returns <v> as given struct slice.
func (v *Var) Structs(pointer interface{}, mapping ...map[string]string) error {
	return yconv.Structs(v.Val(), pointer, mapping...)
}

// StructsDeep converts and returns <v> as given struct slice recursively.
// Deprecated, use Struct instead.
func (v *Var) StructsDeep(pointer interface{}, mapping ...map[string]string) error {
	return yconv.StructsDeep(v.Val(), pointer, mapping...)
}

// Scan automatically calls Struct or Structs function according to the type of parameter
// <pointer> to implement the converting.
// It calls function Struct if <pointer> is type of *struct/**struct to do the converting.
// It calls function Structs if <pointer> is type of *[]struct/*[]*struct to do the converting.
func (v *Var) Scan(pointer interface{}, mapping ...map[string]string) error {
	return yconv.Scan(v.Val(), pointer, mapping...)
}

// ScanDeep automatically calls StructDeep or StructsDeep function according to the type of
// parameter <pointer> to implement the converting.
// It calls function StructDeep if <pointer> is type of *struct/**struct to do the converting.
// It calls function StructsDeep if <pointer> is type of *[]struct/*[]*struct to do the converting.
func (v *Var) ScanDeep(pointer interface{}, mapping ...map[string]string) error {
	return yconv.ScanDeep(v.Val(), pointer, mapping...)
}
