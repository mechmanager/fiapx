// Package domain contains the core entities and port interfaces for the upload-service.
package domain

import "errors"

// ErrInvalidVideoFormat is returned when the uploaded file has an unsupported extension.
var ErrInvalidVideoFormat = errors.New("unsupported video format")
