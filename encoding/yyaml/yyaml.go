
// Package gyaml provides accessing and converting for YAML content.
package yyaml

import (
	"github.com/AmarsDing/lib/internal/json"
	"gopkg.in/yaml.v3"

	"github.com/AmarsDing/lib/util/yconv"
)

func Encode(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func Decode(v []byte) (interface{}, error) {
	var result map[string]interface{}
	if err := yaml.Unmarshal(v, &result); err != nil {
		return nil, err
	}
	return yconv.MapDeep(result), nil
}

func DecodeTo(v []byte, result interface{}) error {
	return yaml.Unmarshal(v, result)
}

func ToJson(v []byte) ([]byte, error) {
	if r, err := Decode(v); err != nil {
		return nil, err
	} else {
		return json.Marshal(r)
	}
}
