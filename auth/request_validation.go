package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Example request struct with validation tags
// type RegisterRequest struct {
// 	Email    string `json:"email" validate:"required,email"`
// 	Password string `json:"password" validate:"required,min=8"`
// }

// DecodeAndValidate decodes a JSON request body into the provided struct type and validates it using struct tags.
// Returns an error if decoding or validation fails.
func DecodeAndValidate[T any](w http.ResponseWriter, r *http.Request) (*T, error) {
	defer r.Body.Close()

	var params T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&params); err != nil {
		return nil, errors.New("invalid request format")
	}

	if err := validate.Struct(params); err != nil {
		return nil, errors.New("validation failed: " + err.Error())
	}

	return &params, nil
}
