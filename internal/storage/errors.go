package storage

import "errors"

var (
	ErrStorageDisabled = errors.New("storage is disabled")
	ErrInvalidRecord   = errors.New("invalid storage record")
	ErrEmptyBatch      = errors.New("empty telemetry batch")
)
