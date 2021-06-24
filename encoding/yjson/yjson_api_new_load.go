package yjson

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/AmarsDing/lib/internal/json"

	"github.com/AmarsDing/lib/encoding/yini"
	"github.com/AmarsDing/lib/encoding/ytoml"
	"github.com/AmarsDing/lib/encoding/yxml"
	"github.com/AmarsDing/lib/encoding/yyaml"
	"github.com/AmarsDing/lib/internal/rwmutex"
	"github.com/AmarsDing/lib/os/yfile"
	"github.com/AmarsDing/lib/text/yregex"
	"github.com/AmarsDing/lib/util/yconv"
)

// New creates a Json object with any variable type of <data>, but <data> should be a map
// or slice for data access reason, or it will make no sense.
//
// The parameter <safe> specifies whether using this Json object in concurrent-safe context,
// which is false in default.
func New(data interface{}, safe ...bool) *Json {
	return NewWithTag(data, "json", safe...)
}

// NewWithTag creates a Json object with any variable type of <data>, but <data> should be a map
// or slice for data access reason, or it will make no sense.
//
// The parameter <tags> specifies priority tags for struct conversion to map, multiple tags joined
// with char ','.
//
// The parameter <safe> specifies whether using this Json object in concurrent-safe context, which
// is false in default.
func NewWithTag(data interface{}, tags string, safe ...bool) *Json {
	option := Options{
		Tags: tags,
	}
	if len(safe) > 0 && safe[0] {
		option.Safe = true
	}
	return NewWithOptions(data, option)
}

// NewWithOptions creates a Json object with any variable type of <data>, but <data> should be a map
// or slice for data access reason, or it will make no sense.
func NewWithOptions(data interface{}, options Options) *Json {
	var j *Json
	switch data.(type) {
	case string, []byte:
		if r, err := loadContentWithOptions(data, options); err == nil {
			j = r
		} else {
			j = &Json{
				p:  &data,
				c:  byte(defaultSplitChar),
				vc: false,
			}
		}
	default:
		var (
			rv   = reflect.ValueOf(data)
			kind = rv.Kind()
		)
		if kind == reflect.Ptr {
			rv = rv.Elem()
			kind = rv.Kind()
		}
		switch kind {
		case reflect.Slice, reflect.Array:
			i := interface{}(nil)
			i = yconv.Interfaces(data)
			j = &Json{
				p:  &i,
				c:  byte(defaultSplitChar),
				vc: false,
			}
		case reflect.Map, reflect.Struct:
			i := interface{}(nil)
			i = yconv.MapDeep(data, options.Tags)
			j = &Json{
				p:  &i,
				c:  byte(defaultSplitChar),
				vc: false,
			}
		default:
			j = &Json{
				p:  &data,
				c:  byte(defaultSplitChar),
				vc: false,
			}
		}
	}
	j.mu = rwmutex.New(options.Safe)
	return j
}

// Load loads content from specified file <path>, and creates a Json object from its content.
func Load(path string, safe ...bool) (*Json, error) {
	if p, err := yfile.Search(path); err != nil {
		return nil, err
	} else {
		path = p
	}
	option := Options{}
	if len(safe) > 0 && safe[0] {
		option.Safe = true
	}
	return doLoadContentWithOptions(yfile.Ext(path), yfile.GetBytesWithCache(path), option)
}

// LoadJson creates a Json object from given JSON format content.
func LoadJson(data interface{}, safe ...bool) (*Json, error) {
	option := Options{}
	if len(safe) > 0 && safe[0] {
		option.Safe = true
	}
	return doLoadContentWithOptions("json", yconv.Bytes(data), option)
}

// LoadXml creates a Json object from given XML format content.
func LoadXml(data interface{}, safe ...bool) (*Json, error) {
	option := Options{}
	if len(safe) > 0 && safe[0] {
		option.Safe = true
	}
	return doLoadContentWithOptions("xml", yconv.Bytes(data), option)
}

// LoadIni creates a Json object from given INI format content.
func LoadIni(data interface{}, safe ...bool) (*Json, error) {
	option := Options{}
	if len(safe) > 0 && safe[0] {
		option.Safe = true
	}
	return doLoadContentWithOptions("ini", yconv.Bytes(data), option)
}

// LoadYaml creates a Json object from given YAML format content.
func LoadYaml(data interface{}, safe ...bool) (*Json, error) {
	option := Options{}
	if len(safe) > 0 && safe[0] {
		option.Safe = true
	}
	return doLoadContentWithOptions("yaml", yconv.Bytes(data), option)
}

// LoadToml creates a Json object from given TOML format content.
func LoadToml(data interface{}, safe ...bool) (*Json, error) {
	option := Options{}
	if len(safe) > 0 && safe[0] {
		option.Safe = true
	}
	return doLoadContentWithOptions("toml", yconv.Bytes(data), option)
}

// LoadContent creates a Json object from given content, it checks the data type of <content>
// automatically, supporting data content type as follows:
// JSON, XML, INI, YAML and TOML.
func LoadContent(data interface{}, safe ...bool) (*Json, error) {
	content := yconv.Bytes(data)
	if len(content) == 0 {
		return New(nil, safe...), nil
	}
	return LoadContentType(checkDataType(content), content, safe...)
}

// LoadContentType creates a Json object from given type and content,
// supporting data content type as follows:
// JSON, XML, INI, YAML and TOML.
func LoadContentType(dataType string, data interface{}, safe ...bool) (*Json, error) {
	content := yconv.Bytes(data)
	if len(content) == 0 {
		return New(nil, safe...), nil
	}
	//ignore UTF8-BOM
	if content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}
	option := Options{}
	if len(safe) > 0 && safe[0] {
		option.Safe = true
	}
	return doLoadContentWithOptions(dataType, content, option)
}

// IsValidDataType checks and returns whether given <dataType> a valid data type for loading.
func IsValidDataType(dataType string) bool {
	if dataType == "" {
		return false
	}
	if dataType[0] == '.' {
		dataType = dataType[1:]
	}
	switch dataType {
	case "json", "js", "xml", "yaml", "yml", "toml", "ini":
		return true
	}
	return false
}

func loadContentWithOptions(data interface{}, options Options) (*Json, error) {
	content := yconv.Bytes(data)
	if len(content) == 0 {
		return NewWithOptions(nil, options), nil
	}
	return loadContentTypeWithOptions(checkDataType(content), content, options)
}

func loadContentTypeWithOptions(dataType string, data interface{}, options Options) (*Json, error) {
	content := yconv.Bytes(data)
	if len(content) == 0 {
		return NewWithOptions(nil, options), nil
	}
	//ignore UTF8-BOM
	if content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}
	return doLoadContentWithOptions(dataType, content, options)
}

// doLoadContent creates a Json object from given content.
// It supports data content type as follows:
// JSON, XML, INI, YAML and TOML.
func doLoadContentWithOptions(dataType string, data []byte, options Options) (*Json, error) {
	var (
		err    error
		result interface{}
	)
	if len(data) == 0 {
		return NewWithOptions(nil, options), nil
	}
	if dataType == "" {
		dataType = checkDataType(data)
	}
	switch dataType {
	case "json", ".json", ".js":

	case "xml", ".xml":
		if data, err = yxml.ToJson(data); err != nil {
			return nil, err
		}

	case "yml", "yaml", ".yml", ".yaml":
		if data, err = yyaml.ToJson(data); err != nil {
			return nil, err
		}

	case "toml", ".toml":
		if data, err = ytoml.ToJson(data); err != nil {
			return nil, err
		}
	case "ini", ".ini":
		if data, err = yini.ToJson(data); err != nil {
			return nil, err
		}
	default:
		err = errors.New("unsupported type for loading")
	}
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if options.StrNumber {
		decoder.UseNumber()
	}
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	switch result.(type) {
	case string, []byte:
		return nil, fmt.Errorf(`json decoding failed for content: %s`, string(data))
	}
	return NewWithOptions(result, options), nil
}

// checkDataType automatically checks and returns the data type for <content>.
// Note that it uses regular expression for loose checking, you can use LoadXXX/LoadContentType
// functions to load the content for certain content type.
func checkDataType(content []byte) string {
	if json.Valid(content) {
		return "json"
	} else if yregex.IsMatch(`^<.+>[\S\s]+<.+>\s*$`, content) {
		return "xml"
	} else if !yregex.IsMatch(`[\n\r]*[\s\t\w\-\."]+\s*=\s*"""[\s\S]+"""`, content) &&
		!yregex.IsMatch(`[\n\r]*[\s\t\w\-\."]+\s*=\s*'''[\s\S]+'''`, content) &&
		((yregex.IsMatch(`^[\n\r]*[\w\-\s\t]+\s*:\s*".+"`, content) || yregex.IsMatch(`^[\n\r]*[\w\-\s\t]+\s*:\s*\w+`, content)) ||
			(yregex.IsMatch(`[\n\r]+[\w\-\s\t]+\s*:\s*".+"`, content) || yregex.IsMatch(`[\n\r]+[\w\-\s\t]+\s*:\s*\w+`, content))) {
		return "yml"
	} else if !yregex.IsMatch(`^[\s\t\n\r]*;.+`, content) &&
		!yregex.IsMatch(`[\s\t\n\r]+;.+`, content) &&
		!yregex.IsMatch(`[\n\r]+[\s\t\w\-]+\.[\s\t\w\-]+\s*=\s*.+`, content) &&
		(yregex.IsMatch(`[\n\r]*[\s\t\w\-\."]+\s*=\s*".+"`, content) || yregex.IsMatch(`[\n\r]*[\s\t\w\-\."]+\s*=\s*\w+`, content)) {
		return "toml"
	} else if yregex.IsMatch(`\[[\w\.]+\]`, content) &&
		(yregex.IsMatch(`[\n\r]*[\s\t\w\-\."]+\s*=\s*".+"`, content) || yregex.IsMatch(`[\n\r]*[\s\t\w\-\."]+\s*=\s*\w+`, content)) {
		// Must contain "[xxx]" section.
		return "ini"
	} else {
		return ""
	}
}
