package uploadawshandlers

import (
	"github.com/STaninnat/ecom-backend/handlers"
)

type HandlersUploadAWSConfig struct {
	*handlers.HandlersConfig
}

type imageUploadS3Response struct {
	Message  string `json:"message"`
	ImageURL string `json:"image_url"`
}
