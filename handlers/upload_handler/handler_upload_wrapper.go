package uploadhandlers

import "github.com/STaninnat/ecom-backend/handlers"

type HandlersUploadConfig struct {
	*handlers.HandlersConfig
}

type imageUploadResponse struct {
	Message  string `json:"message"`
	ImageURL string `json:"image_url"`
}
