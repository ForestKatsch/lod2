package page

import (
	"html/template"
	"lod2/auth"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/sprig/v3"
)

var templateLibrary *template.Template

// A map from the path within `pages/` to the template.
var pageCache map[string]*template.Template

// If false, page templates won't be cached and will be reloaded on each request.
const enablePageCache = false

func init() {
	templateLibrary = loadAllFromDir("templates/library/")
}

func dynamicFuncs() template.FuncMap {
	// Get Sprig's default functions
	funcs := sprig.FuncMap()

	funcs["hasRole"] = func(user *auth.UserInfo, scopeName string, levelName string) bool {
		if user == nil {
			return false
		}
		scope, ok := auth.NameToAccessScope[scopeName]
		if !ok {
			return false
		}
		level, ok := auth.NameToAccessLevel[levelName]
		if !ok {
			return false
		}
		return auth.UserHasRole(user.Roles, scope, level)
	}

	return funcs
}

// Loads a single template from the provided directory and adds it to rootTemplate.
func loadFromFile(rootTemplate *template.Template, name string, filename string) *template.Template {
	file, err := os.ReadFile(filename)

	// Stuff will likely break since we'll be missing files.
	if err != nil {
		log.Printf("! unable to read file '%v': %v", filename, err)
	}

	// log.Printf("parsing template '%v'...", filename)
	newTemplate := template.Must(rootTemplate.New(name).Parse(string(file)))

	return newTemplate
}

// Loads all files from the provided directory. The file path relative to `dir` is used as the template name.
func loadAllFromDir(dir string) *template.Template {
	t := template.New("/")

	t.Funcs(dynamicFuncs())

	err := walkDir(dir, func(path string) {
		// remove the directory from the path (keeping in mind the path has been normalized!)
		name := path[len(dir):]

		loadFromFile(t, name, path)
	})

	// Stuff will likely break since we have _no_ templates.
	if err != nil {
		log.Printf("! unable to walk template directory '%v': %v", dir, err)
	}

	return t
}

// take in a dir to walk and a function to call for each file
func walkDir(dir string, fn func(string)) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Traverse into directories.
		if !info.IsDir() {
			fn(path)
		}

		return nil
	})

	return err
}

// Given the path to a file (relative to `templates/pages/`), returns a template.
// Handles caching and lazy parsing.
func loadPage(path string) (*template.Template, error) {
	if cachedTemplate, exists := pageCache[path]; exists {
		return cachedTemplate, nil
	}

	templ, err := templateLibrary.Clone()

	if err != nil {
		log.Printf("! unable to clone template library for page '%s': %v", path, err)
		return nil, err
	}

	return loadFromFile(templ, path, filepath.Join("templates/pages/", path)), nil
}

type MetaData struct {
	Status   int
	Referrer string
	Now      time.Time
	User     *auth.UserInfo
	Roles    []auth.Role
	Hostname string
}

// renders a single template
func Render(w http.ResponseWriter, r *http.Request, path string, data map[string]interface{}) error {
	templ, err := loadPage(path)

	// Create page data and add some last-second metadata.
	pageData := make(map[string]interface{})
	for k, v := range data {
		pageData[k] = v
	}

	meta := MetaData{
		Referrer: r.Referer(),
		Now:      time.Now(),
		Hostname: r.Host,
	}

	meta.User = auth.GetCurrentUserInfo(r.Context())
	if meta.User != nil {
		meta.Roles = meta.User.Roles
	}

	pageData["Meta"] = meta

	pageData["Const"] = map[string]interface{}{
		"AccessLevelNames": auth.AccessLevelToName,
		"AccessScopeNames": auth.AccessScopeToName,
	}

	err = templ.ExecuteTemplate(w, path, pageData)

	if err != nil {
		log.Printf("! unable to execute template '%v': %v", path, err)
		return err
	}

	return nil
}
