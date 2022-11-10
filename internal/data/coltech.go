// Filename: internal/data/coltech.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"coltech.osborncollins.net/internal/validator"
)

type Coltech struct {
	ID          int64     `json:"id"`
	Created_on  time.Time `json:"created_on"`
	Summary     string    `json:"summary"`
	Description string    `json:"desription"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	Assigned_to string    `json:"assigned_to"`
	Category    string    `json:"category"`
	Department  string    `json:"department"`
	Closed_on   time.Time `json:"closed_on"`
	Created_by  string    `json:"created_by"`
	Due_on      time.Time `json:"due_on"`
	Version     int32     `json:"version"`
}

func ValidateColtech(v *validator.Validator, coltech *Coltech) {

	// Use the check() method to execute our validation checks
	// Summary validation
	v.Check(coltech.Summary != "", "summary", "must be provided")
	v.Check(len(coltech.Summary) <= 300, "summary", "must not be more than 300 bytes long")

	// Description Validation
	v.Check(coltech.Description != "", "description", "must be provided")
	v.Check(len(coltech.Description) <= 1000, "level", "must not be more than 1000 bytes long")

	// Priority validation
	v.Check(coltech.Priority != "", "priority", "must be provided")
	v.Check(len(coltech.Priority) <= 100, "priority", "must not be more than 100 bytes long")

	// Category validation
	v.Check(coltech.Category != "", "category", "must be provided")
	v.Check(len(coltech.Category) <= 200, "category", "must not be more than 200 bytes long")

	// Department validation
	v.Check(coltech.Department != "", "department", "must be provided")
	v.Check(len(coltech.Department) <= 200, "department", "must not be more than 200 bytes long")

	// Created_by validation
	v.Check(coltech.Created_by != "", "created_by", "must be provided")
	v.Check(len(coltech.Created_by) <= 300, "created_by", "must not be more than 300 bytes long")

}

// Define a ColtechModel which wraps a sql.DB connection pool
type ColtechModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new coltech item
func (m ColtechModel) Insert(coltech *Coltech) error {
	query := `
	INSERT INTO tblcoltech (summary, description, priority, category, department, created_by)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_on, version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Collect the data fields into a slice
	args := []interface{}{coltech.Summary, coltech.Description, coltech.Priority,
		coltech.Category, coltech.Department, coltech.Created_by,
	}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&coltech.ID, &coltech.Created_on, &coltech.Version)
}

// GET() allows us to retrieve a specific coltech item
func (m ColtechModel) Get(id int64) (*Coltech, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create query
	query := `
		SELECT id, created_on, summary, description, priority, status, assigned_to, category, department, closed_on, created_by, due_on, version
		FROM tblcoltech
		WHERE id = $1
	`
	// Declare a Coltech variable to hold the return data
	var coltech Coltech
	// Execute Query using the QueryRow
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(

		&coltech.ID,
		&coltech.Created_on,
		&coltech.Summary,
		&coltech.Description,
		&coltech.Priority,
		&coltech.Status,
		&coltech.Assigned_to,
		&coltech.Category,
		&coltech.Department,
		&coltech.Closed_on,
		&coltech.Created_by,
		&coltech.Due_on,
		&coltech.Version,
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &coltech, nil
}

// Update() allows us to edit/alter a coltech item in the list
func (m ColtechModel) Update(coltech *Coltech) error {
	query := `
		UPDATE tblcoltech 
		set summary = $1, description = $2, 
		priority = $3, status = $4, assigned_to = $5,
		category = $6, department = $7, closed_on = $8,
		created_by = $9, due_on = $10, 
		version = version + 1
		WHERE id = $11
		AND version = $12
		RETURNING version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()

	args := []interface{}{
		coltech.ID,
		coltech.Summary,
		coltech.Description,
		coltech.Priority,
		coltech.Status,
		coltech.Assigned_to,
		coltech.Category,
		coltech.Department,
		coltech.Closed_on,
		coltech.Created_by,
		coltech.Due_on,
		coltech.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&coltech.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete() removes a specific coltech item from the list
func (m ColtechModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM tblcoltech
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query
	results, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operations. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// The GetAll() returns a list of all the coltech items sorted by ID
func (m ColtechModel) GetAll(created_by string, assigned_to string, priority string, status string, filters Filters) ([]*Coltech, Metadata, error) {
	// Construct the query
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, created_on, summary, description, priority, status, assigned_to, category, department, closed_on, created_by, due_on, version
		FROM tblcoltech
		WHERE (to_tsvector('simple',created_by) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple',assigned_to) @@ plainto_tsquery('simple', $2) OR $2 = '')
		AND (to_tsvector('simple',priority) @@ plainto_tsquery('simple', $3) OR $3 = '')
		AND (to_tsvector('simple',status) @@ plainto_tsquery('simple', $4) OR $4 = '')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortOrder())

	// Create a 3-second-timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{created_by, assigned_to, priority, status, filters.limit(), filters.offset()}
	// Execute query
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	// Close the result set
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice to hold the coltech data
	coltechs := []*Coltech{}
	// Iterate over the rows in the results set
	for rows.Next() {
		var coltech Coltech
		// Scan the values from the row in to the Coltech struct
		err := rows.Scan(
			&totalRecords,
			&coltech.ID,
			&coltech.Summary,
			&coltech.Description,
			&coltech.Priority,
			&coltech.Status,
			&coltech.Assigned_to,
			&coltech.Category,
			&coltech.Department,
			&coltech.Closed_on,
			&coltech.Created_by,
			&coltech.Due_on,
			&coltech.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// Add the coltech to our slice
		coltechs = append(coltechs, &coltech)
	}
	// Check for errors after looping through the results set
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Return the slice of Coltechs
	return coltechs, metadata, nil
}
