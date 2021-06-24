package yview

import (
	"context"

	"github.com/AmarsDing/lib"
	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/internal/intlog"

	"github.com/AmarsDing/lib/container/yarray"
	"github.com/AmarsDing/lib/os/ycmd"
	"github.com/AmarsDing/lib/os/yfile"
	"github.com/AmarsDing/lib/os/ylog"
)

// View object for template engine.
type View struct {
	paths        *yarray.StrArray       // Searching array for path, NOT concurrent-safe for performance purpose.
	data         map[string]interface{} // Global template variables.
	funcMap      map[string]interface{} // Global template function map.
	fileCacheMap *ymap.StrAnyMap        // File cache map.
	config       Config                 // Extra configuration for the view.
}

type (
	Params  = map[string]interface{} // Params is type for template params.
	FuncMap = map[string]interface{} // FuncMap is type for custom template functions.
)

var (
	// Default view object.
	defaultViewObj *View
)

// checkAndInitDefaultView checks and initializes the default view object.
// The default view object will be initialized just once.
func checkAndInitDefaultView() {
	if defaultViewObj == nil {
		defaultViewObj = New()
	}
}

// ParseContent parses the template content directly using the default view object
// and returns the parsed content.
func ParseContent(ctx context.Context, content string, params ...Params) (string, error) {
	checkAndInitDefaultView()
	return defaultViewObj.ParseContent(ctx, content, params...)
}

// New returns a new view object.
// The parameter <path> specifies the template directory path to load template files.
func New(path ...string) *View {
	view := &View{
		paths:        yarray.NewStrArray(),
		data:         make(map[string]interface{}),
		funcMap:      make(map[string]interface{}),
		fileCacheMap: ymap.NewStrAnyMap(true),
		config:       DefaultConfig(),
	}
	if len(path) > 0 && len(path[0]) > 0 {
		if err := view.SetPath(path[0]); err != nil {
			intlog.Error(err)
		}
	} else {
		// Customized dir path from env/cmd.
		if envPath := ycmd.GetOptWithEnv("lib.gview.path").String(); envPath != "" {
			if yfile.Exists(envPath) {
				if err := view.SetPath(envPath); err != nil {
					intlog.Error(err)
				}
			} else {
				if errorPrint() {
					ylog.Errorf("Template directory path does not exist: %s", envPath)
				}
			}
		} else {
			// Dir path of working dir.
			if err := view.SetPath(yfile.Pwd()); err != nil {
				intlog.Error(err)
			}
			// Dir path of binary.
			if selfPath := yfile.SelfDir(); selfPath != "" && yfile.Exists(selfPath) {
				if err := view.AddPath(selfPath); err != nil {
					intlog.Error(err)
				}
			}
			// Dir path of main package.
			if mainPath := yfile.MainPkgPath(); mainPath != "" && yfile.Exists(mainPath) {
				if err := view.AddPath(mainPath); err != nil {
					intlog.Error(err)
				}
			}
		}
	}
	view.SetDelimiters("{{", "}}")
	// default build-in variables.
	view.data["lib"] = map[string]interface{}{
		"version": lib.VERSION,
	}
	// default build-in functions.
	view.BindFuncMap(FuncMap{
		"eq":         view.buildInFuncEq,
		"ne":         view.buildInFuncNe,
		"lt":         view.buildInFuncLt,
		"le":         view.buildInFuncLe,
		"gt":         view.buildInFuncGt,
		"ge":         view.buildInFuncGe,
		"text":       view.buildInFuncText,
		"html":       view.buildInFuncHtmlEncode,
		"htmlencode": view.buildInFuncHtmlEncode,
		"htmldecode": view.buildInFuncHtmlDecode,
		"encode":     view.buildInFuncHtmlEncode,
		"decode":     view.buildInFuncHtmlDecode,
		"url":        view.buildInFuncUrlEncode,
		"urlencode":  view.buildInFuncUrlEncode,
		"urldecode":  view.buildInFuncUrlDecode,
		"date":       view.buildInFuncDate,
		"substr":     view.buildInFuncSubStr,
		"strlimit":   view.buildInFuncStrLimit,
		"concat":     view.buildInFuncConcat,
		"replace":    view.buildInFuncReplace,
		"compare":    view.buildInFuncCompare,
		"hidestr":    view.buildInFuncHideStr,
		"highlight":  view.buildInFuncHighlight,
		"toupper":    view.buildInFuncToUpper,
		"tolower":    view.buildInFuncToLower,
		"nl2br":      view.buildInFuncNl2Br,
		"include":    view.buildInFuncInclude,
		"dump":       view.buildInFuncDump,
		"map":        view.buildInFuncMap,
		"maps":       view.buildInFuncMaps,
		"json":       view.buildInFuncJson,
	})

	return view
}
