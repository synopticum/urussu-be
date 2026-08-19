// Package swagger serves the OpenAPI schema embedded from the generated
// gen/openapiv2 output (see gen/openapiv2/embed.go) together with a
// Swagger UI rendered from it.
package swagger

import (
	"embed"
	"io/fs"
	"net/http"

	"urussu-be/gen/openapiv2"
)

// ui holds the Swagger UI static assets (swagger-ui-dist 5.32.14, the two
// files needed to render the spec offline) plus the index page wiring them
// to /swagger.json.
//
//go:embed ui
var ui embed.FS

// JsonHandler serves the embedded OpenAPI schema.
func JsonHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(openapiv2.SwaggerSchema)
	})
}

// UIHandler serves the embedded Swagger UI under /swagger/.
func UIHandler() http.Handler {
	// Cannot fail: the ui directory is embedded above.
	assets, _ := fs.Sub(ui, "ui")
	return http.StripPrefix("/swagger/", http.FileServerFS(assets))
}
