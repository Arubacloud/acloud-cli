package cmd

import "errors"

// ErrParsingFailed is returned when one or more Cobra flag reads fail during
// ParseFromCobraCommand. The constructor wraps the underlying errors:
//
//	fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
var ErrParsingFailed = errors.New("parsing flags failed")

// ErrValidationFailed is returned when an Args struct's Validate method detects
// an invalid combination of field values. The constructor wraps the underlying errors:
//
//	fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
var ErrValidationFailed = errors.New("validation failed")
