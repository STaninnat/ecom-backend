package handlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/middlewares"
)

func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	middlewares.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func HandlerError(w http.ResponseWriter, r *http.Request) {
	middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
}
