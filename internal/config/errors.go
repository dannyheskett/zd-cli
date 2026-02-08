package config

import "errors"

var (
	// ErrNoCurrentInstance is returned when no instance is currently selected
	ErrNoCurrentInstance = errors.New("no current instance set")

	// ErrInstanceNotFound is returned when an instance cannot be found
	ErrInstanceNotFound = errors.New("instance not found")

	// ErrInstanceExists is returned when trying to add an instance that already exists
	ErrInstanceExists = errors.New("instance already exists")

	// ErrInvalidInstanceName is returned when an instance name is invalid
	ErrInvalidInstanceName = errors.New("invalid instance name")

	// ErrConfigNotFound is returned when the config file doesn't exist
	ErrConfigNotFound = errors.New("config file not found")
)
