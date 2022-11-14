// Filename: cmd/api/coltech.go

package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"coltech.osborncollins.net/internal/data"
	"coltech.osborncollins.net/internal/validator"
)

// createCOLTECHItemHandler for the "POST" /v1/coltech_items" endpoint
func (app *application) createCOLTECHItemHandler(w http.ResponseWriter, r *http.Request) {
	// Our Target decode destination
	var input struct {
		Summary      string `json:"summary"`
		Description  string `json:"description"`
		Department   string `json:"department"`
		Category     string `json:"category"`
		Priority_val string `json:"priority_val"`
		Created_by   string `json:"created_by"`
	}
	// Initialize a new json.Decoder instance
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//Copy the values from the input struct to a new coltech struct
	coltech := &data.Coltech{
		Summary:      input.Summary,
		Description:  input.Description,
		Department:   input.Department,
		Category:     input.Category,
		Priority_val: input.Priority_val,
		Created_by:   input.Created_by,
	}
	// initialize a new Validator instance
	v := validator.New()

	//Check the map to determine if there were any validation errors
	if data.ValidateColtech(v, coltech); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Create a Coltech Object
	err = app.models.Coltechs.Insert(coltech)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// Create a location header for the newly created resource/Coltech object
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/coltech_items/%d", coltech.ID))
	// Write the JSON response with 201 - created status code with the body
	// being the actual coltech data and the header being the headers map
	err = app.writeJSON(w, http.StatusCreated, envelope{"coltech": coltech}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showCOLTECHItemHandler the "GET" /v1/coltech_items/:id" endpoint
func (app *application) showCOLTECHItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the specific coltech item
	coltech, err := app.models.Coltechs.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the response by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"coltech": coltech}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCOLTECHItemHandler(w http.ResponseWriter, r *http.Request) {
	// This method does a partial replacement
	// Get the id for the coltech item that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the original record from the database
	coltech, err := app.models.Coltechs.Get(id)
	// Error handling
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Create an input struct to hold data read in from the client
	// We update the input struct to use pointers because pointers have a
	// default value of nil false
	// if a field remains nil then we know that the client did not update it
	var input struct {
		Summary      *string    `json:"summary"`
		Description  *string    `json:"desription"`
		Priority_val *string    `json:"priority_val"`
		Status_val   *string    `json:"status_val"`
		Assigned_to  *string    `json:"assigned_to"`
		Category     *string    `json:"category"`
		Department   *string    `json:"department"`
		Closed_on    *time.Time `json:"closed_on"`
		Created_by   *string    `json:"created_by"`
		Due_on       *time.Time `json:"due_on"`
	}

	//Initalize a new json.Decoder instance
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check for updates
	if input.Summary != nil {
		coltech.Summary = *input.Summary
	}
	if input.Description != nil {
		coltech.Description = *input.Description
	}
	if input.Priority_val != nil {
		coltech.Priority_val = *input.Priority_val
	}
	if input.Status_val != nil {
		coltech.Status_val = *input.Status_val
	}
	if input.Assigned_to != nil {
		coltech.Assigned_to = *input.Assigned_to
	}
	if input.Category != nil {
		coltech.Category = *input.Category
	}
	if input.Department != nil {
		coltech.Department = *input.Department
	}
	if input.Closed_on != nil {
		coltech.Closed_on = *input.Closed_on
	}
	if input.Created_by != nil {
		coltech.Created_by = *input.Created_by
	}
	if input.Due_on != nil {
		coltech.Due_on = *input.Due_on
	}

	// Perform Validation on the updated coltech item. If validation fails then
	// we send a 422 - unprocessable entity response to the client
	// initialize a new Validator instance
	v := validator.New()

	//Check the map to determine if there were any validation errors
	if data.ValidateColtech(v, coltech); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Pass the update coltech record to the Update() method
	err = app.models.Coltechs.Update(coltech)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"coltech": coltech}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// The deleteCOLTECHItemHandler() allows the user to delete a coltechs item from the databse by using the ID
func (app *application) deleteCOLTECHItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the coltech item from the database. Send a 404 Not Found status code to the
	// client if there is no matching record
	err = app.models.Coltechs.Delete(id)
	// Error handling
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return 200 Status OK to the client with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "coltech item successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// The listCOLTECHItemsHandler() allows the client to see a listing of coltech items
// based on a set criteria
func (app *application) listCOLTECHItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Create an input struct to hold our query parameter
	var input struct {
		Created_by   string
		Assigned_to  string
		Priority_val string
		Status_val   string
		data.Filters
	}
	// Initialize a validator
	v := validator.New()
	// Get the URL values map
	qs := r.URL.Query()
	// use the helper methods to extract values
	input.Created_by = app.readString(qs, "created_by", "")
	input.Assigned_to = app.readString(qs, "assigned_to", "")
	input.Priority_val = app.readString(qs, "priority_val", "")
	input.Status_val = app.readString(qs, "status_val", "")
	// Get the page information using the read int method
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// Get the sort information
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Specify the allowed sort values
	input.Filters.SortList = []string{"id", "created_by", "priority_val", "assigned_to", "status_val", "-id", "-created_by", "-priority_val", "-assigned_to", "-status_val"}
	// Check for validation errors
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get a listing of all coltech items
	coltechs, metadata, err := app.models.Coltechs.GetAll(input.Created_by, input.Assigned_to, input.Status_val, input.Priority_val, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a JSON response containing all the coltech items
	err = app.writeJSON(w, http.StatusOK, envelope{"coltechs": coltechs, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
