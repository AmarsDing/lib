package yconv

import (
	"time"

	"github.com/AmarsDing/lib/internal/utils"
)

// Time converts `any` to time.Time.
func Time(any interface{}, format ...string) time.Time {
	// It's already this type.
	if len(format) == 0 {
		if v, ok := any.(time.Time); ok {
			return v
		}
	}
	if t := ytime(any, format...); t != nil {
		return t.Time
	}
	return time.Time{}
}

// Duration converts `any` to time.Duration.
// If `any` is string, then it uses time.ParseDuration to convert it.
// If `any` is numeric, then it converts `any` as nanoseconds.
func Duration(any interface{}) time.Duration {
	// It's already this type.
	if v, ok := any.(time.Duration); ok {
		return v
	}
	s := String(any)
	if !utils.IsNumeric(s) {
		d, _ := ytime.ParseDuration(s)
		return d
	}
	return time.Duration(Int64(any))
}

// ytime converts `any` to *ytime.Time.
// The parameter `format` can be used to specify the format of `any`.
// If no `format` given, it converts `any` using ytime.NewFromTimeStamp if `any` is numeric,
// or using ytime.StrToTime if `any` is string.
func ytime(any interface{}, format ...string) *ytime.Time {
	if any == nil {
		return nil
	}
	if v, ok := any.(apiytime); ok {
		return v.ytime(format...)
	}
	// It's already this type.
	if len(format) == 0 {
		if v, ok := any.(*ytime.Time); ok {
			return v
		}
		if t, ok := any.(time.Time); ok {
			return ytime.New(t)
		}
		if t, ok := any.(*time.Time); ok {
			return ytime.New(t)
		}
	}
	s := String(any)
	if len(s) == 0 {
		return ytime.New()
	}
	// Priority conversion using given format.
	if len(format) > 0 {
		t, _ := ytime.StrToTimeFormat(s, format[0])
		return t
	}
	if utils.IsNumeric(s) {
		return ytime.NewFromTimeStamp(Int64(s))
	} else {
		t, _ := ytime.StrToTime(s)
		return t
	}
}
