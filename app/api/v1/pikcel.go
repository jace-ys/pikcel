package apiv1

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("pikcel", func() {
	Title("Pikcel")
	Description("A production-ready Go service deployed on Kubernetes")
	Version("1.0.0")
	Server("pikcel", func() {
		Services("api")
		Host("local-http", func() {
			URI("http://localhost:8080")
		})
		Host("local-grpc", func() {
			URI("grpc://localhost:8081")
		})
	})
})

var _ = Service("api", func() {
	Error(ErrCodeUnauthenticated)
	Error(ErrCodeAccessDenied)

	HTTP(func() {
		Path("/api/v1")
		Response(ErrCodeUnauthenticated, StatusUnauthorized)
		Response(ErrCodeAccessDenied, StatusForbidden)
	})

	GRPC(func() {
		Response(ErrCodeUnauthenticated, CodeUnauthenticated)
		Response(ErrCodeAccessDenied, CodePermissionDenied)
	})

	Method("CanvasGet", func() {
		NoSecurity()

		Result(Canvas)

		HTTP(func() {
			GET("/canvas")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})

	Files("/openapi.json", "gen/http/openapi3.json")
})
