package backend

import (
	"errors"
	"time"
)

var (
	// ErrBackendNotProvided error
	ErrBackendNotProvided = errors.New("backend proxy not provided")
	// ErrHealthCheckStatusNotOk error
	ErrHealthCheckStatusNotOk = errors.New("health check response status not correct")
)

const (
	// RTTErrorHappened means health check failed
	RTTErrorHappened = time.Hour * 24 * 365 * 99
	// RTTErrorNotCheck now
	RTTErrorNotCheck = time.Hour * 24 * 365 * 98
)
