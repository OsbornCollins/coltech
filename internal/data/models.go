// Filename: internal/data/models.go

package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Create a Wrapper for our data models

type Models struct {
	Coltechs ColtechModel
	Users    UserModel
}

// NewModels() allows us to create a new Models
func NewModels(db *sql.DB) Models {
	return Models{
		Coltechs: ColtechModel{DB: db},
		Users:    UserModel{DB: db},
	}
}
