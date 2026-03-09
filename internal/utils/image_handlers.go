package utils

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func ValidateBase64Image(image string) error {
	decodedImage, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		decodedImage, err = base64.RawStdEncoding.DecodeString(image)
		if err != nil {
			return ErrInvalidImageFile
		}
	}

	if len(decodedImage) == 0 {
		return ErrInvalidImageFile
	}

	if !strings.HasPrefix(http.DetectContentType(decodedImage), "image/") {
		return ErrInvalidImageFile
	}

	return nil
}

func DetectContentType(contentType string) string {
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	case "image/gif":
		ext = ".gif"
	}

	return ext
}
