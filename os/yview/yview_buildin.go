package yview

import (
	"context"
	"fmt"
	"strings"

	"github.com/AmarsDing/lib/internal/json"
	"github.com/AmarsDing/lib/util/yutil"

	"github.com/AmarsDing/lib/encoding/yhtml"
	"github.com/AmarsDing/lib/encoding/yurl"
	"github.com/AmarsDing/lib/os/ytime"
	"github.com/AmarsDing/lib/text/ystr"
	"github.com/AmarsDing/lib/util/yconv"

	htmltpl "html/template"
)

// buildInFuncDump implements build-in template function: dump
func (view *View) buildInFuncDump(values ...interface{}) (result string) {
	result += "<!--\n"
	for _, v := range values {
		result += yutil.Export(v) + "\n"
	}
	result += "-->\n"
	return result
}

// buildInFuncMap implements build-in template function: map
func (view *View) buildInFuncMap(value ...interface{}) map[string]interface{} {
	if len(value) > 0 {
		return yconv.Map(value[0])
	}
	return map[string]interface{}{}
}

// buildInFuncMaps implements build-in template function: maps
func (view *View) buildInFuncMaps(value ...interface{}) []map[string]interface{} {
	if len(value) > 0 {
		return yconv.Maps(value[0])
	}
	return []map[string]interface{}{}
}

// buildInFuncEq implements build-in template function: eq
func (view *View) buildInFuncEq(value interface{}, others ...interface{}) bool {
	s := yconv.String(value)
	for _, v := range others {
		if strings.Compare(s, yconv.String(v)) == 0 {
			return true
		}
	}
	return false
}

// buildInFuncNe implements build-in template function: ne
func (view *View) buildInFuncNe(value, other interface{}) bool {
	return strings.Compare(yconv.String(value), yconv.String(other)) != 0
}

// buildInFuncLt implements build-in template function: lt
func (view *View) buildInFuncLt(value, other interface{}) bool {
	s1 := yconv.String(value)
	s2 := yconv.String(other)
	if ystr.IsNumeric(s1) && ystr.IsNumeric(s2) {
		return yconv.Int64(value) < yconv.Int64(other)
	}
	return strings.Compare(s1, s2) < 0
}

// buildInFuncLe implements build-in template function: le
func (view *View) buildInFuncLe(value, other interface{}) bool {
	s1 := yconv.String(value)
	s2 := yconv.String(other)
	if ystr.IsNumeric(s1) && ystr.IsNumeric(s2) {
		return yconv.Int64(value) <= yconv.Int64(other)
	}
	return strings.Compare(s1, s2) <= 0
}

// buildInFuncGt implements build-in template function: gt
func (view *View) buildInFuncGt(value, other interface{}) bool {
	s1 := yconv.String(value)
	s2 := yconv.String(other)
	if ystr.IsNumeric(s1) && ystr.IsNumeric(s2) {
		return yconv.Int64(value) > yconv.Int64(other)
	}
	return strings.Compare(s1, s2) > 0
}

// buildInFuncGe implements build-in template function: ge
func (view *View) buildInFuncGe(value, other interface{}) bool {
	s1 := yconv.String(value)
	s2 := yconv.String(other)
	if ystr.IsNumeric(s1) && ystr.IsNumeric(s2) {
		return yconv.Int64(value) >= yconv.Int64(other)
	}
	return strings.Compare(s1, s2) >= 0
}

// buildInFuncInclude implements build-in template function: include
// Note that configuration AutoEncode does not affect the output of this function.
func (view *View) buildInFuncInclude(file interface{}, data ...map[string]interface{}) htmltpl.HTML {
	var m map[string]interface{} = nil
	if len(data) > 0 {
		m = data[0]
	}
	path := yconv.String(file)
	if path == "" {
		return ""
	}
	// It will search the file internally.
	content, err := view.Parse(context.TODO(), path, m)
	if err != nil {
		return htmltpl.HTML(err.Error())
	}
	return htmltpl.HTML(content)
}

// buildInFuncText implements build-in template function: text
func (view *View) buildInFuncText(html interface{}) string {
	return yhtml.StripTags(yconv.String(html))
}

// buildInFuncHtmlEncode implements build-in template function: html
func (view *View) buildInFuncHtmlEncode(html interface{}) string {
	return yhtml.Entities(yconv.String(html))
}

// buildInFuncHtmlDecode implements build-in template function: htmldecode
func (view *View) buildInFuncHtmlDecode(html interface{}) string {
	return yhtml.EntitiesDecode(yconv.String(html))
}

// buildInFuncUrlEncode implements build-in template function: url
func (view *View) buildInFuncUrlEncode(url interface{}) string {
	return yurl.Encode(yconv.String(url))
}

// buildInFuncUrlDecode implements build-in template function: urldecode
func (view *View) buildInFuncUrlDecode(url interface{}) string {
	if content, err := yurl.Decode(yconv.String(url)); err == nil {
		return content
	} else {
		return err.Error()
	}
}

// buildInFuncDate implements build-in template function: date
func (view *View) buildInFuncDate(format interface{}, timestamp ...interface{}) string {
	t := int64(0)
	if len(timestamp) > 0 {
		t = yconv.Int64(timestamp[0])
	}
	if t == 0 {
		t = ytime.Timestamp()
	}
	return ytime.NewFromTimeStamp(t).Format(yconv.String(format))
}

// buildInFuncCompare implements build-in template function: compare
func (view *View) buildInFuncCompare(value1, value2 interface{}) int {
	return strings.Compare(yconv.String(value1), yconv.String(value2))
}

// buildInFuncSubStr implements build-in template function: substr
func (view *View) buildInFuncSubStr(start, end, str interface{}) string {
	return ystr.SubStrRune(yconv.String(str), yconv.Int(start), yconv.Int(end))
}

// buildInFuncStrLimit implements build-in template function: strlimit
func (view *View) buildInFuncStrLimit(length, suffix, str interface{}) string {
	return ystr.StrLimitRune(yconv.String(str), yconv.Int(length), yconv.String(suffix))
}

// buildInFuncConcat implements build-in template function: concat
func (view *View) buildInFuncConcat(str ...interface{}) string {
	var s string
	for _, v := range str {
		s += yconv.String(v)
	}
	return s
}

// buildInFuncReplace implements build-in template function: replace
func (view *View) buildInFuncReplace(search, replace, str interface{}) string {
	return ystr.Replace(yconv.String(str), yconv.String(search), yconv.String(replace), -1)
}

// buildInFuncHighlight implements build-in template function: highlight
func (view *View) buildInFuncHighlight(key, color, str interface{}) string {
	return ystr.Replace(yconv.String(str), yconv.String(key), fmt.Sprintf(`<span style="color:%v;">%v</span>`, color, key))
}

// buildInFuncHideStr implements build-in template function: hidestr
func (view *View) buildInFuncHideStr(percent, hide, str interface{}) string {
	return ystr.HideStr(yconv.String(str), yconv.Int(percent), yconv.String(hide))
}

// buildInFuncToUpper implements build-in template function: toupper
func (view *View) buildInFuncToUpper(str interface{}) string {
	return ystr.ToUpper(yconv.String(str))
}

// buildInFuncToLower implements build-in template function: toupper
func (view *View) buildInFuncToLower(str interface{}) string {
	return ystr.ToLower(yconv.String(str))
}

// buildInFuncNl2Br implements build-in template function: nl2br
func (view *View) buildInFuncNl2Br(str interface{}) string {
	return ystr.Nl2Br(yconv.String(str))
}

// buildInFuncJson implements build-in template function: json ,
// which encodes and returns <value> as JSON string.
func (view *View) buildInFuncJson(value interface{}) (string, error) {
	b, err := json.Marshal(value)
	return yconv.UnsafeBytesToStr(b), err
}
