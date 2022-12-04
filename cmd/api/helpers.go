// Filename: cmd/api/helpers.go

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"coltech.osborncollins.net/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// Define a new type name envelope
type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	// Use the "ParamsFromContext()" function to get the request context as a slice
	params := httprouter.ParamsFromContext(r.Context())
	// Get the value of the "id" parameter
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	//Convert our map into a JSON object
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	// Add a newline to make viewing on the terminal easier
	js = append(js, '\n')
	// Add the headers
	for key, value := range headers {
		w.Header()[key] = value
	}
	// Specify that we will serve our responses using JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Write the []byte slice containing the JSON response body
	w.Write(js)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use http.MaxBytesreader() to limit the size of the request body to
	// 1 MB 2^20
	maxBytes := 1_048_576
	// Decode the request body into the target destination
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)

	// Check for a bad request
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		// Switch to check for the errors
		switch {
		//Check for syntax errors
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON(at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		// Check for wrong types passed by the client
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
			// Empty body
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		//Unmappable fields
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		// Request too large
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		// Pass non-nil pointer error
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
			// Default
		default:
			return err
		}
	}

	//call Decode() again
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// The readString() method returns a string value from the query parameters
// string or it returns a default value if no matching key is found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// The readInt() method converts a string value from the query string to an integer value
// If the value cannot be converted to an integer then a validation error is added to
// the validations errors map.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	// Get the value
	value := qs.Get(key)
	if value == "" {
		return defaultValue
	}
	// Perform the conversion to an integer
	intValue, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return intValue
}

// Background accepts a function as its parameter
func (app *application) background(fn func()) {
	// Increment the WaitGroup counter
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()
}
