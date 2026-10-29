package models

import (
	"database/sql"
	"time"

	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// Define a new User type. Notice how the field names and types align
// with the columns in the database "users" table?
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

// Define a new UserModel type which wraps a database connection pool.
type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Retrieve the user ID and hashed password associated with the email.
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}

		return 0, err
	}

	// Check whether the hashed password matches the supplied password.
	err = bcrypt.CompareHashAndPassword(
		hashedPassword,
		[]byte(password),
	)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}

		return 0, err
	}

	// The credentials are correct, so return the user ID.
	return id, nil
}

func (m *UserModel) Insert(name, email, password string) error {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `
		INSERT INTO users (name, email, hashed_password, created)
		VALUES (?, ?, ?, UTC_TIMESTAMP())
	`

	// Insert the user details and hashed password into the users table.
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		// Check whether the error is a MySQL error.
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			// MySQL error 1062 means a duplicate entry.
			if mySQLError.Number == 1062 &&
				strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}

		return err
	}

	return nil
}

// We'll use the Exists method to check if a user exists with a specific ID.
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt :=
		"SELECT EXISTS(SELECT true FROM users WHERE id = ?)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
