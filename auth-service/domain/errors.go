// Package domain contains the core entities and repository interfaces for the auth-service.
package domain

import "errors"

var (
	// ErrEmailAlreadyExists is returned when registering with a duplicate email.
	ErrEmailAlreadyExists = errors.New("email already registered")
	// ErrInvalidCredentials is returned when login credentials are wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserNotFound is returned when no user matches the query.
	ErrUserNotFound = errors.New("user not found")
)
