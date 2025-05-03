package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

func DecodeAndValidate[T any](w http.ResponseWriter, r *http.Request) (*T, error) {
	defer r.Body.Close()

	var params T
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return nil, errors.New("invalid request format")
	}

	v := reflect.ValueOf(params)
	for i := range v.NumField() {
		if v.Field(i).Interface() == "" {
			return nil, errors.New("missing required fields")
		}
	}

	return &params, nil
}
