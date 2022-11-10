// Filename: cmd/api/routes.go

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
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

	return router
}
