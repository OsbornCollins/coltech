// Filename: internal/data/users.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"coltech.osborncollins.net/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	ID         int64     `json:"id"`
	Created_on time.Time `json:"created_on"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Activated  bool      `json:"activated"`
	Version    int       `json:"-"`
}

// Create a custom password type
type password struct {
	plaintext *string
	hash      []byte
}

// The set() method stores the hash of the plaintext password.

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// The Matches() method checks if the supplied password is correct
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// Validate the client request
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRx), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be atleast 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	// Validate the email
	ValidateEmail(v, user.Email)
	// Validate the password
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	// Ensure a hash of the password was created
	if user.Password.hash == nil {
		panic("missing password hash for the user")
	}
}

// Create our user model
type UserModel struct {
	DB *sql.DB
}

// Create a new user
func (m UserModel) Insert(user *User) error {
	// Create our query
	query := `
		INSERT INTO tblusers (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_on, version
	`
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Created_on, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraints "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

// Get user based on their email
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, created_on, name, email, password_hash, activated, version
	FROM tblUsers
	WHERE email = $1
	`
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Created_on,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

// The client can update their information
func (m UserModel) Update(user *User) error {
	query := `
		UPDATE tblusers
		SET name = $1, email = $2, password_hash = $3,
		activated = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
	`
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraints "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}
