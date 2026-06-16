package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
)

// FmtAPIError formats an SDK API error response into a Go error.
func FmtAPIError(statusCode int, title, detail *string) error {
	msg := fmt.Sprintf("API error (status %d)", statusCode)
	if title != nil {
		msg += ": " + *title
	}
	if detail != nil {
		msg += " — " + *detail
	}
	return fmt.Errorf("%s", msg)
}

// FromV2 extracts an *aruba.HTTPError from err and formats it via FmtAPIError.
// Returns the original error unchanged if it is not an HTTPError. Nil-safe on
// ErrResp/Title/Detail. Field-level validation errors are appended when present.
// When debug is true the raw response body is printed to stderr.
func FromV2(err error, debug bool) error {
	if err == nil {
		return nil
	}
	var httpErr *aruba.HTTPError
	if !errors.As(err, &httpErr) {
		return err
	}
	var title, detail *string
	if httpErr.ErrResp != nil {
		title = httpErr.ErrResp.Title
		detail = httpErr.ErrResp.Detail
	}
	base := FmtAPIError(httpErr.StatusCode, title, detail)

	// Append per-field validation errors so callers see exactly what is wrong.
	// The SDK's ValidationError uses "field"/"message" JSON tags; some API endpoints
	// return "fieldName"/"errorMessage" instead. Parse the raw body for the alternate
	// shape when the SDK-parsed entries are all empty strings.
	var parts []string
	if httpErr.ErrResp != nil {
		for _, ve := range httpErr.ErrResp.Errors {
			switch {
			case ve.Field != "" && ve.Message != "":
				parts = append(parts, ve.Field+": "+ve.Message)
			case ve.Message != "":
				parts = append(parts, ve.Message)
			case ve.Field != "":
				parts = append(parts, ve.Field)
			}
		}
	}
	if len(parts) == 0 && len(httpErr.Body) > 0 {
		var rawBody struct {
			Errors []struct {
				FieldName    string `json:"fieldName"`
				ErrorMessage string `json:"errorMessage"`
			} `json:"errors"`
		}
		if json.Unmarshal(httpErr.Body, &rawBody) == nil {
			for _, ve := range rawBody.Errors {
				if ve.FieldName != "" && ve.ErrorMessage != "" {
					parts = append(parts, ve.FieldName+": "+ve.ErrorMessage)
				} else if ve.ErrorMessage != "" {
					parts = append(parts, ve.ErrorMessage)
				}
			}
		}
	}
	if len(parts) > 0 {
		base = fmt.Errorf("%w — validation: %s", base, strings.Join(parts, "; "))
	}

	if debug && len(httpErr.Body) > 0 {
		fmt.Fprintf(os.Stderr, "[DEBUG] response body: %s\n", httpErr.Body)
	}

	return base
}
