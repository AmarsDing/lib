

package yjson

import "github.com/AmarsDing/lib/util/yconv"

// ToMap converts current Json object to map[string]interface{}.
// It returns nil if fails.
// Deprecated, use Map instead.
func (j *Json) ToMap() map[string]interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.Map(*(j.p))
}

// ToArray converts current Json object to []interface{}.
// It returns nil if fails.
// Deprecated, use Array instead.
func (j *Json) ToArray() []interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.Interfaces(*(j.p))
}

// ToStruct converts current Json object to specified object.
// The <pointer> should be a pointer type of *struct.
// Deprecated, use Struct instead.
func (j *Json) ToStruct(pointer interface{}, mapping ...map[string]string) error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.Struct(*(j.p), pointer, mapping...)
}

// ToStructDeep converts current Json object to specified object recursively.
// The <pointer> should be a pointer type of *struct.
// Deprecated, use Struct instead.
func (j *Json) ToStructDeep(pointer interface{}, mapping ...map[string]string) error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.StructDeep(*(j.p), pointer, mapping...)
}

// ToStructs converts current Json object to specified object slice.
// The <pointer> should be a pointer type of []struct/*struct.
// Deprecated, use Structs instead.
func (j *Json) ToStructs(pointer interface{}, mapping ...map[string]string) error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.Structs(*(j.p), pointer, mapping...)
}

// ToStructsDeep converts current Json object to specified object slice recursively.
// The <pointer> should be a pointer type of []struct/*struct.
// Deprecated, use Structs instead.
func (j *Json) ToStructsDeep(pointer interface{}, mapping ...map[string]string) error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.StructsDeep(*(j.p), pointer, mapping...)
}

// ToScan automatically calls Struct or Structs function according to the type of parameter
// <pointer> to implement the converting..
// Deprecated, use Scan instead.
func (j *Json) ToScan(pointer interface{}, mapping ...map[string]string) error {
	return yconv.Scan(*(j.p), pointer, mapping...)
}

// ToScanDeep automatically calls StructDeep or StructsDeep function according to the type of
// parameter <pointer> to implement the converting..
// Deprecated, use Scan instead.
func (j *Json) ToScanDeep(pointer interface{}, mapping ...map[string]string) error {
	return yconv.ScanDeep(*(j.p), pointer, mapping...)
}

// ToMapToMap converts current Json object to specified map variable.
// The parameter of <pointer> should be type of *map.
// Deprecated, use MapToMap instead.
func (j *Json) ToMapToMap(pointer interface{}, mapping ...map[string]string) error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.MapToMap(*(j.p), pointer, mapping...)
}

// ToMapToMaps converts current Json object to specified map variable slice.
// The parameter of <pointer> should be type of []map/*map.
// Deprecated, use MapToMaps instead.
func (j *Json) ToMapToMaps(pointer interface{}, mapping ...map[string]string) error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.MapToMaps(*(j.p), pointer, mapping...)
}

// ToMapToMapsDeep converts current Json object to specified map variable slice recursively.
// The parameter of <pointer> should be type of []map/*map.
// Deprecated, use MapToMaps instead.
func (j *Json) ToMapToMapsDeep(pointer interface{}, mapping ...map[string]string) error {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return yconv.MapToMapsDeep(*(j.p), pointer, mapping...)
}
