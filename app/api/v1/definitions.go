package apiv1

import (
	"embed"

	. "goa.design/goa/v3/dsl"
)

//go:embed gen/http/*.json gen/http/*.yaml
var OpenAPIFS embed.FS

const (
	ErrCodeUnauthenticated = "unauthenticated"
	ErrCodeAccessDenied    = "access_denied"
)

var Canvas = ResultType("application/vnd.pikcel.canvas`", "Canvas", func() {
	Field(1, "id", String)
	Field(2, "width", Int32)
	Field(3, "height", Int32)
	Required("id", "width", "height")
})
