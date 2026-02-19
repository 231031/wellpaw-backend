package utils

import (
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct[T any](payload *T) (*model.HTTPResponse, error) {
	if err := validate.Struct(payload); err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}, err
	}
	return nil, nil
}
