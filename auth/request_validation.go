package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"github.com/STaninnat/ecom-backend/middlewares"
)

func DecodeAndValidate[T any](w http.ResponseWriter, r *http.Request) (*T, bool) {
	defer r.Body.Close()

	var params T
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Println("Decode error: ", err)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return nil, false
	}

	v := reflect.ValueOf(params)
	for i := range v.NumField() {
		if v.Field(i).Interface() == "" {
			log.Println("Invalid request format: missing fields")
			middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
			return nil, false
		}
	}

	return &params, true
}
