package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func WriteJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func RespondWithError(w http.ResponseWriter, code int, err error) {
	if code > 499 {
		BaseLogger.Error().Int("status", code).Err(err).Msg("Responded with 5XX error")
	}

	type errResponse struct {
		Error string `json:"error"`
	}

	WriteJSON(w, code, errResponse{
		Error: err.Error(),
	})
}

func ParseJson(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(payload); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("malformed JSON at position %d", syntaxErr.Offset)
		case errors.As(err, &unmarshalTypeErr):
			BaseLogger.Warn().Str("value", unmarshalTypeErr.Value).Msg("unmarshal type error")
			return fmt.Errorf("invalid value for field '%s', expected type '%s', received type '%s'",
				unmarshalTypeErr.Field,
				unmarshalTypeErr.Type,
				unmarshalTypeErr.Value,
			)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("empty request body")
		case strings.Contains(err.Error(), "json: unknown field"):
			fieldName := extractUnknownFieldName(err.Error())
			return fmt.Errorf("unknown field '%s' in request body", fieldName)
		default:
			return fmt.Errorf("invalid JSON payload")
		}
	}

	return nil
}

func ValidatePayload(payload any) error {
	err := validate.Struct(payload)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			var msgs []string
			for _, fe := range validationErrs {
				msgs = append(msgs, fmt.Sprintf("field '%s' failed on '%s' validation", fe.Field(), fe.Tag()))
			}
			return fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
		}
		return err
	}
	return nil
}

func extractUnknownFieldName(errMsg string) string {
	start := strings.Index(errMsg, "\"") + 1
	end := strings.LastIndex(errMsg, "\"")
	if start > 0 && end > start {
		return errMsg[start:end]
	}
	return "unknown"
}
