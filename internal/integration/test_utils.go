package integration

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// addURLParams adds URL parameters to a request for chi router testing
func addURLParams(req *http.Request, key, value string) *http.Request {
	// Create a chi route context and add the URL parameter
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	return req.WithContext(ctx)
}
