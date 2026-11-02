package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Response envelope for API errors.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// JSON sends a JSON response with the provided HTTP status code and payload.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, `{"error":"failed to encode JSON response"}`, http.StatusInternalServerError)
		}
	}
}

// Error sends a standardized JSON error response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{
		Error: message,
		Code:  status,
	})
}

// ReadJSON decodes request body into dst with 1MB size limit and unknown field rejection.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Restrict request body size to 1MB to prevent unbounded memory allocation DOS attacks
	maxBytes := int64(1048576)
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("malformed JSON at position %d", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("malformed JSON: unexpected end of body")
		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("invalid value for field %q at position %d", unmarshalTypeError.Field, unmarshalTypeError.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown field %s in JSON body", fieldName)
		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")
		case err.Error() == "http: request body too large":
			return errors.New("request body exceeds maximum allowed size of 1MB")
		default:
			return err
		}
	}

	// Ensure there is only a single JSON value in the body
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain only a single JSON object")
	}

	return nil
}
