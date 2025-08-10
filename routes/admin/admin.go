package admin

import (
	"lod2/auth"
	"lod2/db"
	"lod2/middleware"
	"lod2/page"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.AuthRequiredMiddleware())

	r.Mount("/users", userRouter())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "admin/index.html", map[string]interface{}{})
	})

	r.Get("/sql", func(w http.ResponseWriter, r *http.Request) {
		if !auth.VerifyRole(r.Context(), auth.DangerousSql, auth.Edit) {
			page.Render401(w, r)
			return
		}
		page.Render(w, r, "admin/sql.html", map[string]interface{}{})
	})

	r.Post("/db/execute", func(w http.ResponseWriter, r *http.Request) {
		if !auth.VerifyRole(r.Context(), auth.DangerousSql, auth.Edit) {
			page.Render401(w, r)
			return
		}

		r.ParseForm()
		query := r.Form.Get("query")
		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			page.Render(w, r, "admin/db/fragment-execute.html", map[string]interface{}{"Error": "query required"})
			return
		}

		rows, err := db.DB.Query(query)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			page.Render(w, r, "admin/db/fragment-execute.html", map[string]interface{}{"Error": err.Error()})
			return
		}

		defer rows.Close()
		// The page itself has {{ range $index, $row := $.Rows }}
		// and needs to display the column names too.

		columns, err := rows.Columns()

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			page.Render(w, r, "admin/db/fragment-execute.html", map[string]interface{}{"Error": err.Error()})
			return
		}

		// Add data to be passed to the template
		data := map[string]interface{}{
			"Columns": columns,
			"Rows":    []map[string]interface{}{},
		}

		for rows.Next() {
			columnMap := make(map[string]interface{})
			columnPointers := make([]interface{}, len(columns))
			for i := range columnPointers {
				columnPointers[i] = new(interface{})
			}

			if err := rows.Scan(columnPointers...); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				page.Render(w, r, "admin/db/fragment-execute.html", map[string]interface{}{"Error": err.Error()})
				return
			}

			for i, colName := range columns {
				if val, ok := (*(columnPointers[i].(*interface{}))).(interface{ Valid() bool }); ok && !val.Valid() {
					columnMap[colName] = ""
				} else {
					columnMap[colName] = *(columnPointers[i].(*interface{}))
				}
			}
			data["Rows"] = append(data["Rows"].([]map[string]interface{}), columnMap)
		}

		if err := rows.Err(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			page.Render(w, r, "admin/db/fragment-execute.html", map[string]interface{}{"Error": err.Error()})
			return
		}

		log.Printf("Rows: %d", len(data["Rows"].([]map[string]interface{})))

		page.Render(w, r, "admin/db/fragment-execute.html", data)
	})

	return r
}
