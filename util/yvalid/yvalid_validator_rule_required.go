

package yvalid

import (
	"github.com/AmarsDing/lib/internal/empty"
	"github.com/AmarsDing/lib/util/yconv"
	"github.com/AmarsDing/lib/util/yutil"
	"reflect"
	"strings"
)

// checkRequired checks `value` using required rules.
// It also supports require checks for `value` of type: slice, map.
func (v *Validator) checkRequired(value interface{}, ruleKey, rulePattern string, dataMap map[string]interface{}) bool {
	required := false
	switch ruleKey {
	// Required.
	case "required":
		required = true

	// Required unless all given field and its value are equal.
	// Example: required-if: id,1,age,18
	case "required-if":
		required = false
		var (
			array      = strings.Split(rulePattern, ",")
			foundValue interface{}
		)
		// It supports multiple field and value pairs.
		if len(array)%2 == 0 {
			for i := 0; i < len(array); {
				tk := array[i]
				tv := array[i+1]
				_, foundValue = gutil.MapPossibleItemByKey(dataMap, tk)
				if strings.Compare(tv, yconv.String(foundValue)) == 0 {
					required = true
					break
				}
				i += 2
			}
		}

	// Required unless all given field and its value are not equal.
	// Example: required-unless: id,1,age,18
	case "required-unless":
		required = true
		var (
			array      = strings.Split(rulePattern, ",")
			foundValue interface{}
		)
		// It supports multiple field and value pairs.
		if len(array)%2 == 0 {
			for i := 0; i < len(array); {
				tk := array[i]
				tv := array[i+1]
				_, foundValue = gutil.MapPossibleItemByKey(dataMap, tk)
				if strings.Compare(tv, yconv.String(foundValue)) == 0 {
					required = false
					break
				}

				i += 2
			}
		}

	// Required if any of given fields are not empty.
	// Example: required-with:id,name
	case "required-with":
		required = false
		var (
			array      = strings.Split(rulePattern, ",")
			foundValue interface{}
		)
		for i := 0; i < len(array); i++ {
			_, foundValue = gutil.MapPossibleItemByKey(dataMap, array[i])
			if !empty.IsEmpty(foundValue) {
				required = true
				break
			}
		}

	// Required if all of given fields are not empty.
	// Example: required-with:id,name
	case "required-with-all":
		required = true
		var (
			array      = strings.Split(rulePattern, ",")
			foundValue interface{}
		)
		for i := 0; i < len(array); i++ {
			_, foundValue = gutil.MapPossibleItemByKey(dataMap, array[i])
			if empty.IsEmpty(foundValue) {
				required = false
				break
			}
		}

	// Required if any of given fields are empty.
	// Example: required-with:id,name
	case "required-without":
		required = false
		var (
			array      = strings.Split(rulePattern, ",")
			foundValue interface{}
		)
		for i := 0; i < len(array); i++ {
			_, foundValue = gutil.MapPossibleItemByKey(dataMap, array[i])
			if empty.IsEmpty(foundValue) {
				required = true
				break
			}
		}

	// Required if all of given fields are empty.
	// Example: required-with:id,name
	case "required-without-all":
		required = true
		var (
			array      = strings.Split(rulePattern, ",")
			foundValue interface{}
		)
		for i := 0; i < len(array); i++ {
			_, foundValue = gutil.MapPossibleItemByKey(dataMap, array[i])
			if !empty.IsEmpty(foundValue) {
				required = false
				break
			}
		}
	}
	if required {
		reflectValue := reflect.ValueOf(value)
		for reflectValue.Kind() == reflect.Ptr {
			reflectValue = reflectValue.Elem()
		}
		switch reflectValue.Kind() {
		case reflect.String, reflect.Map, reflect.Array, reflect.Slice:
			return reflectValue.Len() != 0
		}
		return yconv.String(value) != ""
	} else {
		return true
	}
}
