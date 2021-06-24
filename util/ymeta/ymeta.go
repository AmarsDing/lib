// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package gmeta provides embedded meta data feature for struct.
package ymeta

import (
	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/container/yvar"
	"github.com/AmarsDing/lib/internal/structs"
)

// Meta is used as an embedded attribute for struct to enabled meta data feature.
type Meta struct{}

const (
	// metaAttributeName is the attribute name of meta data in struct.
	metaAttributeName = "Meta"
)

var (
	// metaDataCacheMap is a cache map for struct type to enhance the performance.
	metaDataCacheMap = ymap.NewStrAnyMap(true)
)

// Data retrieves and returns all meta data from `object`.
// It automatically parses and caches the tag string from "Mata" attribute as its meta data.
func Data(object interface{}) map[string]interface{} {
	reflectType, err := structs.StructType(object)
	if err != nil {
		panic(err)
	}
	return metaDataCacheMap.GetOrSetFuncLock(reflectType.Signature(), func() interface{} {
		if field, ok := reflectType.FieldByName(metaAttributeName); ok {
			var (
				tags = structs.ParseTag(string(field.Tag))
				data = make(map[string]interface{}, len(tags))
			)
			for k, v := range tags {
				data[k] = v
			}
			return data
		}
		return map[string]interface{}{}
	}).(map[string]interface{})
}

// Get retrieves and returns specified meta data by `key` from `object`.
func Get(object interface{}, key string) *yvar.Var {
	return yvar.New(Data(object)[key])
}
