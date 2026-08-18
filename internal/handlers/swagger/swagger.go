// Package swagger serves the OpenAPI schema embedded from the generated
// gen/openapiv2 output (see gen/openapiv2/embed.go).
package swagger

import (
	"net/http"

	"urussu-be/gen/openapiv2"
)

// Handler serves the embedded OpenAPI schema.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(openapiv2.SwaggerSchema)
	})
}
