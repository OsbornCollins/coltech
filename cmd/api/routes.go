// Filename: cmd/api/routes.go

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Create a new httprouter router instance
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/coltech_items", app.listCOLTECHItemsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/coltech_items", app.createCOLTECHItemHandler)
	router.HandlerFunc(http.MethodGet, "/v1/coltech_items/:id", app.showCOLTECHItemHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/coltech_items/:id", app.updateCOLTECHItemHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/coltech_items/:id", app.deleteCOLTECHItemHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
