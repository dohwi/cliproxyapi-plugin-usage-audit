package plugin

import (
	"errors"
	"fmt"
)

// apiError is a client-facing request error. message stays a human-readable
// English sentence (existing API clients read "error"); code is a stable
// snake_case identifier the dashboard maps to its own translations (err.<code>).
type apiError struct {
	code    string
	message string
}

func (e *apiError) Error() string { return e.message }

func newAPIError(code, format string, args ...any) error {
	return &apiError{code: code, message: fmt.Sprintf(format, args...)}
}

// Stable error codes. Changing one is a breaking change for the dashboard.
const (
	codeBadRequest          = "bad_request"
	codeNotFound            = "not_found"
	codeStoreNotReady       = "store_not_ready"
	codeInvalidTime         = "invalid_time_format"
	codeInvalidTimeZone     = "invalid_time_zone"
	codeInvalidTZOffset     = "invalid_timezone_offset"
	codeInvalidRange        = "invalid_range"
	codeFutureFrom          = "future_from"
	codeInvalidStatus       = "invalid_status"
	codeInvalidLimit        = "invalid_limit"
	codeInvalidPage         = "invalid_page"
	codeSummaryRange        = "summary_range_too_large"
	codeSummaryBucketLimit  = "summary_bucket_limit"
	codeInvalidCursor       = "invalid_cursor"
	codeCursorExpired       = "cursor_expired"
	codeReadFailed          = "read_failed"
	codeTruncateFailed      = "truncate_failed"
	codeInternalMarshalFail = "marshal_failed"
)

// errorBody is the JSON error shape of every management/resource route:
// {"error": <English message>, "code": <snake_case code>}.
func errorBody(code, message string) map[string]any {
	return map[string]any{"error": message, "code": code}
}

// errorBodyFrom renders err, falling back to bad_request for untyped errors.
func errorBodyFrom(err error) map[string]any {
	var ae *apiError
	if errors.As(err, &ae) {
		return errorBody(ae.code, ae.message)
	}
	return errorBody(codeBadRequest, err.Error())
}
