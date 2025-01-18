package page

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var templ map[string]*template.Template

func init() {
	templ = loadFrom("templates/")
}

func loadFrom(dir string) map[string]*template.Template {
	templates := make(map[string]*template.Template)

	t := template.New("__root")

	t.Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("02 Jan 2006 15:04:05 MST")
		},
	})

	err := walkDir(dir, func(path string) {
		// remove the directory from the path (keeping in mind the path has been normalized!)
		name := path[len(dir):]

		// now read the file as a string
		file, err := os.ReadFile(path)

		if err != nil {
			fmt.Printf("Error reading file %v: %v\n", path, err)
			panic(err)
		}

		fmt.Printf("Parsing template '%v'...\n", name)
		templates[name] = template.Must(t.New(name).Parse(string(file)))
	})

	if err != nil {
		panic(err)
	}

	return templates
}

// take in a dir to walk and a function to call for each file
func walkDir(dir string, fn func(string)) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fn(path)
		}
		return nil
	})

	return err
}

// renders a template
func Execute(w http.ResponseWriter, name string, data map[string]interface{}) error {
	t, ok := templ[name]

	if !ok {
		return fmt.Errorf("template not found: %v", name)
	}

	pageData := make(map[string]interface{})
	for k, v := range data {
		pageData[k] = v
	}
	pageData["Meta"] = map[string]interface{}{
		"Now": time.Now(),
	}
	return t.Execute(w, pageData)
}
