package yview

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	htmltpl "html/template"
	"strconv"
	"strings"
	texttpl "text/template"

	"github.com/AmarsDing/lib/encoding/yhash"
	"github.com/AmarsDing/lib/errors/yerror"
	"github.com/AmarsDing/lib/internal/intlog"
	"github.com/AmarsDing/lib/os/yfsnotify"
	"github.com/AmarsDing/lib/os/ymlock"
	"github.com/AmarsDing/lib/text/ystr"
	"github.com/AmarsDing/lib/util/yconv"
	"github.com/AmarsDing/lib/util/yutil"

	"github.com/AmarsDing/lib/os/yres"

	"github.com/AmarsDing/lib/container/ymap"
	"github.com/AmarsDing/lib/os/yfile"
	"github.com/AmarsDing/lib/os/ylog"
	"github.com/AmarsDing/lib/os/yspath"
)

const (
	// Template name for content parsing.
	templateNameForContentParsing = "TemplateContent"
)

// fileCacheItem is the cache item for template file.
type fileCacheItem struct {
	path    string
	folder  string
	content string
}

var (
	// Templates cache map for template folder.
	// Note that there's no expiring logic for this map.
	templates = ymap.NewStrAnyMap(true)

	// Try-folders for resource template file searching.
	resourceTryFolders = []string{"template/", "template", "/template", "/template/"}
)

// Parse parses given template file <file> with given template variables <params>
// and returns the parsed template content.
func (view *View) Parse(ctx context.Context, file string, params ...Params) (result string, err error) {
	var tpl interface{}
	// It caches the file, folder and its content to enhance performance.
	r := view.fileCacheMap.GetOrSetFuncLock(file, func() interface{} {
		var (
			path     string
			folder   string
			content  string
			resource *yres.File
		)
		// Searching the absolute file path for <file>.
		path, folder, resource, err = view.searchFile(file)
		if err != nil {
			return nil
		}
		if resource != nil {
			content = yconv.UnsafeBytesToStr(resource.Content())
		} else {
			content = yfile.GetContentsWithCache(path)
		}
		// Monitor template files changes using fsnotify asynchronously.
		if resource == nil {
			if _, err := yfsnotify.AddOnce("gview.Parse:"+folder, folder, func(event *yfsnotify.Event) {
				// CLEAR THEM ALL.
				view.fileCacheMap.Clear()
				templates.Clear()
				yfsnotify.Exit()
			}); err != nil {
				intlog.Error(err)
			}
		}
		return &fileCacheItem{
			path:    path,
			folder:  folder,
			content: content,
		}
	})
	if r == nil {
		return
	}
	item := r.(*fileCacheItem)
	// It's not necessary continuing parsing if template content is empty.
	if item.content == "" {
		return "", nil
	}
	// Get the template object instance for <folder>.
	tpl, err = view.getTemplate(item.path, item.folder, fmt.Sprintf(`*%s`, yfile.Ext(item.path)))
	if err != nil {
		return "", err
	}
	// Using memory lock to ensure concurrent safety for template parsing.
	ymlock.LockFunc("gview.Parse:"+item.path, func() {
		if view.config.AutoEncode {
			tpl, err = tpl.(*htmltpl.Template).Parse(item.content)
		} else {
			tpl, err = tpl.(*texttpl.Template).Parse(item.content)
		}
		if err != nil && item.path != "" {
			err = yerror.Wrap(err, item.path)
		}
	})
	if err != nil {
		return "", err
	}
	// Note that the template variable assignment cannot change the value
	// of the existing <params> or view.data because both variables are pointers.
	// It needs to merge the values of the two maps into a new map.
	variables := yutil.MapMergeCopy(params...)
	if len(view.data) > 0 {
		yutil.MapMerge(variables, view.data)
	}
	view.setI18nLanguageFromCtx(ctx, variables)

	buffer := bytes.NewBuffer(nil)
	if view.config.AutoEncode {
		newTpl, err := tpl.(*htmltpl.Template).Clone()
		if err != nil {
			return "", err
		}
		if err := newTpl.Execute(buffer, variables); err != nil {
			return "", err
		}
	} else {
		if err := tpl.(*texttpl.Template).Execute(buffer, variables); err != nil {
			return "", err
		}
	}

	// TODO any graceful plan to replace "<no value>"?
	result = ystr.Replace(buffer.String(), "<no value>", "")
	result = view.i18nTranslate(ctx, result, variables)
	return result, nil
}

// ParseDefault parses the default template file with params.
func (view *View) ParseDefault(ctx context.Context, params ...Params) (result string, err error) {
	return view.Parse(ctx, view.config.DefaultFile, params...)
}

// ParseContent parses given template content <content>  with template variables <params>
// and returns the parsed content in []byte.
func (view *View) ParseContent(ctx context.Context, content string, params ...Params) (string, error) {
	// It's not necessary continuing parsing if template content is empty.
	if content == "" {
		return "", nil
	}
	err := (error)(nil)
	key := fmt.Sprintf("%s_%v_%v", templateNameForContentParsing, view.config.Delimiters, view.config.AutoEncode)
	tpl := templates.GetOrSetFuncLock(key, func() interface{} {
		if view.config.AutoEncode {
			return htmltpl.New(templateNameForContentParsing).Delims(
				view.config.Delimiters[0],
				view.config.Delimiters[1],
			).Funcs(view.funcMap)
		}
		return texttpl.New(templateNameForContentParsing).Delims(
			view.config.Delimiters[0],
			view.config.Delimiters[1],
		).Funcs(view.funcMap)
	})
	// Using memory lock to ensure concurrent safety for content parsing.
	hash := strconv.FormatUint(yhash.DJBHash64([]byte(content)), 10)
	ymlock.LockFunc("gview.ParseContent:"+hash, func() {
		if view.config.AutoEncode {
			tpl, err = tpl.(*htmltpl.Template).Parse(content)
		} else {
			tpl, err = tpl.(*texttpl.Template).Parse(content)
		}
	})
	if err != nil {
		return "", err
	}
	// Note that the template variable assignment cannot change the value
	// of the existing <params> or view.data because both variables are pointers.
	// It needs to merge the values of the two maps into a new map.
	variables := yutil.MapMergeCopy(params...)
	if len(view.data) > 0 {
		yutil.MapMerge(variables, view.data)
	}
	view.setI18nLanguageFromCtx(ctx, variables)

	buffer := bytes.NewBuffer(nil)
	if view.config.AutoEncode {
		newTpl, err := tpl.(*htmltpl.Template).Clone()
		if err != nil {
			return "", err
		}
		if err := newTpl.Execute(buffer, variables); err != nil {
			return "", err
		}
	} else {
		if err := tpl.(*texttpl.Template).Execute(buffer, variables); err != nil {
			return "", err
		}
	}
	// TODO any graceful plan to replace "<no value>"?
	result := ystr.Replace(buffer.String(), "<no value>", "")
	result = view.i18nTranslate(ctx, result, variables)
	return result, nil
}

// getTemplate returns the template object associated with given template file <path>.
// It uses template cache to enhance performance, that is, it will return the same template object
// with the same given <path>. It will also automatically refresh the template cache
// if the template files under <path> changes (recursively).
func (view *View) getTemplate(filePath, folderPath, pattern string) (tpl interface{}, err error) {
	// Key for template cache.
	key := fmt.Sprintf("%s_%v", filePath, view.config.Delimiters)
	result := templates.GetOrSetFuncLock(key, func() interface{} {
		tplName := filePath
		if view.config.AutoEncode {
			tpl = htmltpl.New(tplName).Delims(
				view.config.Delimiters[0],
				view.config.Delimiters[1],
			).Funcs(view.funcMap)
		} else {
			tpl = texttpl.New(tplName).Delims(
				view.config.Delimiters[0],
				view.config.Delimiters[1],
			).Funcs(view.funcMap)
		}
		// Firstly checking the resource manager.
		if !yres.IsEmpty() {
			if files := yres.ScanDirFile(folderPath, pattern, true); len(files) > 0 {
				var err error
				if view.config.AutoEncode {
					t := tpl.(*htmltpl.Template)
					for _, v := range files {
						_, err = t.New(v.FileInfo().Name()).Parse(string(v.Content()))
						if err != nil {
							err = view.formatTemplateObjectCreatinyerror(v.Name(), tplName, err)
							return nil
						}
					}
				} else {
					t := tpl.(*texttpl.Template)
					for _, v := range files {
						_, err = t.New(v.FileInfo().Name()).Parse(string(v.Content()))
						if err != nil {
							err = view.formatTemplateObjectCreatinyerror(v.Name(), tplName, err)
							return nil
						}
					}
				}
				return tpl
			}
		}

		// Secondly checking the file system.
		var (
			files []string
		)
		files, err = yfile.ScanDir(folderPath, pattern, true)
		if err != nil {
			return nil
		}
		if view.config.AutoEncode {
			t := tpl.(*htmltpl.Template)
			for _, file := range files {
				if _, err = t.Parse(yfile.GetContents(file)); err != nil {
					err = view.formatTemplateObjectCreatinyerror(file, tplName, err)
					return nil
				}
			}
		} else {
			t := tpl.(*texttpl.Template)
			for _, file := range files {
				if _, err = t.Parse(yfile.GetContents(file)); err != nil {
					err = view.formatTemplateObjectCreatinyerror(file, tplName, err)
					return nil
				}
			}
		}
		return tpl
	})
	if result != nil {
		return result, nil
	}
	return
}

// formatTemplateObjectCreatinyerror formats the error that creted from creating template object.
func (view *View) formatTemplateObjectCreatinyerror(filePath, tplName string, err error) error {
	if err != nil {
		return yerror.NewSkip(1, ystr.Replace(err.Error(), tplName, filePath))
	}
	return nil
}

// searchFile returns the found absolute path for <file> and its template folder path.
// Note that, the returned <folder> is the template folder path, but not the folder of
// the returned template file <path>.
func (view *View) searchFile(file string) (path string, folder string, resource *yres.File, err error) {
	// Firstly checking the resource manager.
	if !yres.IsEmpty() {
		// Try folders.
		for _, folderPath := range resourceTryFolders {
			if resource = yres.Get(folderPath + file); resource != nil {
				path = resource.Name()
				folder = folderPath
				return
			}
		}
		// Search folders.
		view.paths.RLockFunc(func(array []string) {
			for _, v := range array {
				v = strings.TrimRight(v, "/"+yfile.Separator)
				if resource = yres.Get(v + "/" + file); resource != nil {
					path = resource.Name()
					folder = v
					break
				}
				if resource = yres.Get(v + "/template/" + file); resource != nil {
					path = resource.Name()
					folder = v + "/template"
					break
				}
			}
		})
	}

	// Secondly checking the file system.
	if path == "" {
		view.paths.RLockFunc(func(array []string) {
			for _, folderPath := range array {
				folderPath = strings.TrimRight(folderPath, yfile.Separator)
				if path, _ = yspath.Search(folderPath, file); path != "" {
					folder = folderPath
					break
				}
				if path, _ = yspath.Search(folderPath+yfile.Separator+"template", file); path != "" {
					folder = folderPath + yfile.Separator + "template"
					break
				}
			}
		})
	}

	// Error checking.
	if path == "" {
		buffer := bytes.NewBuffer(nil)
		if view.paths.Len() > 0 {
			buffer.WriteString(fmt.Sprintf("[gview] cannot find template file \"%s\" in following paths:", file))
			view.paths.RLockFunc(func(array []string) {
				index := 1
				for _, folderPath := range array {
					folderPath = strings.TrimRight(folderPath, "/")
					if folderPath == "" {
						folderPath = "/"
					}
					buffer.WriteString(fmt.Sprintf("\n%d. %s", index, folderPath))
					index++
					buffer.WriteString(fmt.Sprintf("\n%d. %s", index, strings.TrimRight(folderPath, "/")+yfile.Separator+"template"))
					index++
				}
			})
		} else {
			buffer.WriteString(fmt.Sprintf("[gview] cannot find template file \"%s\" with no path set/add", file))
		}
		if errorPrint() {
			ylog.Error(buffer.String())
		}
		err = errors.New(fmt.Sprintf(`template file "%s" not found`, file))
	}
	return
}
