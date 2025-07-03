package utilsuploaders

import "testing"

func TestAllowedImageExtensions(t *testing.T) {
	extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	for _, ext := range extensions {
		if _, ok := AllowedImageExtensions[ext]; !ok {
			t.Errorf("AllowedImageExtensions missing %q", ext)
		}
	}
	if _, ok := AllowedImageExtensions[".exe"]; ok {
		t.Errorf(".exe should not be allowed")
	}
}
